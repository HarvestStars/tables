package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/HarvestStars/tables/gredis"
	"github.com/HarvestStars/tables/logging"
	"github.com/HarvestStars/tables/setting"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

func getLongShortAddr(targetSlot int) (string, string, int, int) {
	beg := targetSlot * setting.LavadBaseSetting.BlocksInSlot
	end := beg + setting.LavadBaseSetting.BlocksInSlot - 1
	longKey := "slot_" + strconv.Itoa(int(targetSlot)) + "_long"
	shortKey := "slot_" + strconv.Itoa(int(targetSlot)) + "_short"

	logger := logging.GetLogger()
	logger.Infof("get LONG and SHORT addr info, slot index: %d", targetSlot)
	longAddr, err := gredis.Get(longKey)
	if err != nil {
		logger.Error(err)
	}

	shortAddr, err := gredis.Get(shortKey)
	if err != nil {
		logger.Error(err)
	}

	return longAddr, shortAddr, beg, end
}

//Addr2ScriptHash Addr to ScriptHash
func Addr2ScriptHash(addr string) string {
	address, err := btcutil.DecodeAddress(addr, &chaincfg.MainNetParams)
	if err != nil {
		logging.GetLogger().Error(err)
		return ""
	}

	script, err := txscript.PayToAddrScript(address)
	if err != nil {
		logging.GetLogger().Error(err)
		return ""
	}

	h := sha256.New()
	h.Write(script)
	data := h.Sum(nil)

	return hex.EncodeToString(reverse(data))
}

// DecodeAndAddBalance comment
func DecodeAndAddBalance(raw string, long *AddressBalance, short *AddressBalance) error {
	serializedTx, err := hex.DecodeString(raw)
	if err != nil {
		panic(err)
	}
	var mtx wire.MsgTx
	err = mtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		panic(err)
	}
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

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
