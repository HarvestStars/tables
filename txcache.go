package main

import (
	"github.com/HarvestStars/tables/gredis"
	"github.com/HarvestStars/tables/logging"
)

func rawTxsInCache(beg int, end int, height int, hash string) (string, bool) {
	if height != 0 {
		// unconfirmed tx's height is 0
		if height < beg || height > end {
			return "", false
		}
	}
	var raw string
	var err error
	if gredis.Exists("rawtx_" + hash) {
		raw, err = gredis.Get("rawtx_" + hash)
		if err != nil {
			logging.GetLogger().Error(err)
			return "", false
		}
	} else {
		raw, err = node.BlockchainTransactionGet(hash)
		if err != nil {
			logging.GetLogger().Error(err)
			return "", false
		}
		logging.GetLogger().Infof("raw tx %s is saved in redis", hash)
		gredis.Set("rawtx_"+hash, raw, 0)
	}
	return raw, true
}
