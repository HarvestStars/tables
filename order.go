package main

import (
	"encoding/json"
	"strconv"

	"github.com/HarvestStars/tables/gredis"
	"github.com/HarvestStars/tables/logging"
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

	for _, tx := range txs {
		raw, ok := rawTxsInCache(beg, end, tx.Height, tx.Hash)
		if !ok {
			continue
		}
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
