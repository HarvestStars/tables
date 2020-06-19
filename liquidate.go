package main

import (
	"encoding/json"
	"strconv"

	"github.com/HarvestStars/tables/gredis"
	"github.com/HarvestStars/tables/logging"
	"github.com/HarvestStars/tables/setting"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

type beneficiary struct {
	LongBenefi  map[string]int64 `json:"LongBenefi"`
	ShortBenefi map[string]int64 `json:"ShortBenefi"`
}

func liquidate(height int) {
	// 判断是否已经清算完毕
	done := gredis.Exists("finish_liquidslot_" + strconv.Itoa(slot-1))
	if height%setting.LavadBaseSetting.BlocksInSlot > 5 && slot-1 >= 0 && !done {
		logging.GetLogger().Info("start liquidating on slot %d", slot-1)
		longAddr, shortAddr, beg, end := getLongShortAddr(slot - 1)
		longTxs, err := node.BlockchainAddressListUnspent(Addr2ScriptHash(longAddr))
		logging.GetLogger().Info("liquidating LONG address: %s, on slot %d", longAddr, slot-1)
		if err != nil {
			logging.GetLogger().Error(err)
			return
		}

		shortTxs, err := node.BlockchainAddressListUnspent(Addr2ScriptHash(shortAddr))
		logging.GetLogger().Info("liquidating SHORT address: %s, on slot %d", shortAddr, slot-1)
		if err != nil {
			logging.GetLogger().Error(err)
			return
		}

		// Duplication check for change addr
		txs := longTxs
		if len(txs) == 0 {
			txs = shortTxs
		} else {
			for _, v := range shortTxs {
				for _, tx := range txs {
					if v.Hash == tx.Hash {
						continue
					}
				}
				txs = append(txs, v)
			}
		}

		participateLong := make(map[string]int64)
		participateShort := make(map[string]int64)
		totalLongAmount := int64(0)
		totalShortAmount := int64(0)
		for _, tx := range txs {
			raw, ok := rawTxsInCache(beg, end, tx.Height, tx.Hash)
			if !ok {
				continue
			}
			logging.GetLogger().Infof("liquidating on txid: %s", tx.Hash)
			longAmount, shortAmount := addOneTxInParticipates(raw, participateLong, participateShort, longAddr, shortAddr)
			totalLongAmount += longAmount
			totalShortAmount += shortAmount
		}

		// 最终清算，获取受益人获利详情
		beneficiaryLong := make(map[string]int64)
		beneficiaryShort := make(map[string]int64)
		finnalLiquidate(totalLongAmount, totalShortAmount, participateLong, participateShort, beneficiaryLong, beneficiaryShort)

		// 序列化
		beneAll := beneficiary{LongBenefi: beneficiaryLong, ShortBenefi: beneficiaryShort}
		data, _ := json.Marshal(&beneAll)
		gredis.Set("liquid_"+strconv.Itoa(slot-1), string(data), 0)
		gredis.Set("finish_liquidslot_"+strconv.Itoa(slot-1), "done", 0)
	}
}

func addOneTxInParticipates(rawTx string, longSet map[string]int64, shortSet map[string]int64, longAddr string, shortAddr string) (int64, int64) {
	mtx := decodeTx(rawTx)
	senderLock := false
	var sender string
	longAmount := int64(0)
	shortAmount := int64(0)
	for _, out := range mtx.TxOut {
		scriptClass, addrs, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
		if txscript.ScriptHashTy != scriptClass {
			continue
		}
		for _, v := range addrs {
			if v.String() == longAddr {
				longAmount += out.Value
			}
			if v.String() == shortAddr {
				shortAmount += out.Value
			}
		}
	}

	for _, in := range mtx.TxIn {
		preRaw, err := node.BlockchainTransactionGet(in.PreviousOutPoint.Hash.String())
		if err != nil {
			logging.GetLogger().Error(err)
		}
		preTx := decodeTx(preRaw)
		senderScript := preTx.TxOut[in.PreviousOutPoint.Index].PkScript
		scriptClass, addrs, _, _ := txscript.ExtractPkScriptAddrs(senderScript, &chaincfg.MainNetParams)
		if (scriptClass == txscript.NullDataTy) || (scriptClass == txscript.NonStandardTy) {
			logging.GetLogger().Info("sender address decode error, the script is null or nonstandard")
			continue
		}

		if !senderLock {
			sender = addrs[0].String()
			logging.GetLogger().Infof("liquidating, sender is found and locked: %s", sender)
			senderLock = true
		}
	}

	// update 两方账目表
	if _, ok := longSet[sender]; ok {
		longSet[sender] += longAmount
	} else {
		if longAmount != 0 {
			longSet[sender] = longAmount
		}
	}
	if _, ok := shortSet[sender]; ok {
		shortSet[sender] += shortAmount
	} else {
		if shortAmount != 0 {
			shortSet[sender] = shortAmount
		}
	}
	return longAmount, shortAmount
}

func finnalLiquidate(longTotal int64, shortTotal int64, longPartiSet map[string]int64, shortPartiSet map[string]int64, longBenefSet map[string]int64, shortBenefSet map[string]int64) {
	for longAddr, amount := range longPartiSet {
		longBenefSet[longAddr] = (longTotal + shortTotal) * amount / longTotal / 100000000
	}
	for shorAddr, amount := range shortPartiSet {
		shortBenefSet[shorAddr] = (longTotal + shortTotal) * amount / shortTotal / 100000000
	}
}
