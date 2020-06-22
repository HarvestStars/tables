package main

import (
	"encoding/json"
	"strconv"

	"github.com/HarvestStars/tables/gredis"
	"github.com/HarvestStars/tables/logging"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// AddressBalance comment
type AddressBalance struct {
	Addr    string `json:"address"`
	Balance int64  `json:"amount"`
}

type tableOrder struct {
	Amount int64          `json:"total"`
	Long   AddressBalance `json:"long"`
	Short  AddressBalance `json:"short"`
}

type participate struct {
	PoolEntrySet []memPoolEntryWithHash `json:"pooltxs"`
}

type memPoolEntryWithHash struct {
	Hash    string           `json:"txid"`
	Account map[string]int64 `json:"account"`
}

func calcLongShort() {
	logger := logging.GetLogger()
	logger.Infof("calc long and short, slot index: %d", slot)
	longAddr, shortAddr, beg, end := getLongShortAddr(slot)
	fetchInfo(longAddr, shortAddr, beg, end)
}

func fetchInfo(longAddr string, shortAddr string, beg int, end int) {
	longTxs, err := node.BlockchainAddressListUnspent(Addr2ScriptHash(longAddr))
	if err != nil {
		logging.GetLogger().Error(err)
		return
	}

	shortTxs, err := node.BlockchainAddressListUnspent(Addr2ScriptHash(shortAddr))
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
	long := &AddressBalance{
		Addr:    longAddr,
		Balance: 0,
	}
	short := &AddressBalance{
		Addr:    shortAddr,
		Balance: 0,
	}

	// 周期变化后，重置poolEntrySet
	if slotOld != slot {
		poolEntrySet = make([]memPoolEntryWithHash, 0, 1000)
		slotOld = slot
	}

	for _, tx := range txs {
		raw, ok := rawTxsInCache(beg, end, tx.Height, tx.Hash)
		if !ok {
			continue
		}

		// 新交易滚动功能，提供cache
		// 仅缓存两个周期的txs
		entryLong := make(map[string]int64)
		entryShort := make(map[string]int64)
		longAddr, shortAddr, _, _ := getLongShortAddr(slot)
		_, _ = addOneTxInParticipates(raw, entryLong, entryShort, longAddr, shortAddr, true)

		isInPoolSet := false
		for _, entry := range poolEntrySet {
			// 判断tx.Hash 是否已经存在poolEntrySet中
			if entry.Hash == tx.Hash {
				isInPoolSet = true
			}
		}

		if !isInPoolSet {
			// 更新redis和cache
			if len(entryLong) != 0 {
				//poolEntrySet[len(poolEntrySet)] = memPoolEntryWithHash{Hash: tx.Hash, Account: entryLong}
				poolEntrySet = append(poolEntrySet, memPoolEntryWithHash{Hash: tx.Hash, Account: entryLong})
			}
			if len(entryShort) != 0 {
				//poolEntrySet[len(poolEntrySet)] = memPoolEntryWithHash{Hash: tx.Hash, Account: entryShort}
				poolEntrySet = append(poolEntrySet, memPoolEntryWithHash{Hash: tx.Hash, Account: entryShort})
			}

			ParticipateSet := participate{PoolEntrySet: poolEntrySet}
			data, err := json.Marshal(ParticipateSet)
			if err != nil {
				logging.GetLogger().Error(err)
			}
			gredis.Set("participate_"+strconv.Itoa(slot), string(data), 0)
			if gredis.Exists("participate_" + strconv.Itoa(slot-2)) {
				gredis.Delete("participate_" + strconv.Itoa(slot-2))
			}
		}

		// DecodeAndAddBalance 进行order信息累加
		DecodeAndAddBalance(raw, long, short)
	}
	logging.GetLogger().Info("Last info: ", *long, *short)
	//write to redis
	//order_50:{"total":2000000,"long":{"addr":"3FbigTuPm8xg8NErwKFPvFMUh1to1vhhGf", "amout":1000000},"short":"3FbigTuPm8xg8NErwKFPvFMUh1to1vhhGf", "amout":1000000}}
	key := "order_" + strconv.Itoa(slot)
	order := tableOrder{
		Amount: long.Balance + short.Balance,
		Long:   *long,
		Short:  *short,
	}
	data, _ := json.Marshal(&order)
	gredis.Set(key, string(data), 0)
}

// DecodeAndAddBalance comment
func DecodeAndAddBalance(raw string, long *AddressBalance, short *AddressBalance) error {
	mtx := decodeTx(raw)
	for _, out := range mtx.TxOut {
		scriptClass, addrs, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
		if txscript.ScriptHashTy != scriptClass {
			continue
		}
		for _, v := range addrs {
			if v.String() == long.Addr {
				long.Balance += out.Value
			}
			if v.String() == short.Addr {
				short.Balance += out.Value
			}
		}
	}
	return nil
}
