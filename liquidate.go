package main

import (
	"github.com/HarvestStars/tables/logging"
	"github.com/HarvestStars/tables/setting"
)

func liquidate(height int) {
	if height%setting.LavadBaseSetting.BlocksInSlot > 6 && slot-1 >= 0 {
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

		beneficiarySet := make(map[string]int32)
		for _, tx := range txs {
			raw, ok := rawTxsInCache(beg, end, tx.Height, tx.Hash)
			if !ok {
				continue
			}
			logging.GetLogger().Infof("liquidating on txid: %s, the tx raw: %s", tx.Hash, raw)
			benef, amount := liquidateOneTx(raw)
			beneficiarySet[benef] += amount
		}
	}
}

func liquidateOneTx(rawTx string) (string, int32) {

	return "", 0
}
