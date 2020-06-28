package main

import (
	"github.com/HarvestStars/tables/gredis"
	"github.com/HarvestStars/tables/logging"
	"github.com/HarvestStars/tables/setting"
)

func rawTxsInCache(beg int, end int, targetSlot int, txHeight int, hash string, deadLine int) (string, bool) {
	if txHeight != 0 {
		// unconfirmed tx's height is 0
		if txHeight < beg || txHeight >= (end-deadLine) {
			return "", false
		}
	} else {
		// 对内存池中的tx，
		// case 1, 当前区块高度已经超过该tx本该入块的高度（针对有人作恶在n周期时不断向n-1地址打币）
		if height >= (targetSlot+1)*setting.LavadBaseSetting.BlocksInSlot {
			return "", false
		}

		// case 2, 正常逻辑下，当前区块高度已经到达本周期死区，直接返回
		if height >= ((slot+1)*setting.LavadBaseSetting.BlocksInSlot - deadLine) {
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
