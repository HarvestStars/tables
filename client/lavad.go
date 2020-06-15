package main

import (
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
)

// LavaService 全节点钱包服务
type LavaService struct {
	masterKey string
}

// TODO...
func (ls *LavaService) loadMasterPubKey() error {

	return nil
}

// GetAddresses 投注地址
// TODO...
func (ls *LavaService) GetAddresses(index uint32) ([]string, error) {
	return nil, nil
}

// NewLavaService 创建
// TODO...
func NewLavaService(conf *LavaServiceConf) *LavaService {
	service := &LavaService{
		masterKey: conf.HDMasterPubKey,
	}
	if err := service.loadMasterPubKey(); err != nil {
		// TODO logger
		return nil
	}
	return service
}

// LavaServiceConf 全节点钱包服务
type LavaServiceConf struct {
	Host           string
	Port           int
	RPCUser        string
	RPCPassword    string
	HDMasterPubKey string
	StartSlot      uint32
}

//GenerateAddress generate bitcoin address
func GenerateAddress(key string, count int, start string, net *chaincfg.Params, isSigWit bool) ([]string, error) {
	master, err := hdkeychain.NewKeyFromString(key)
	if err != nil {
		return nil, err
	}
	strs := strings.Split(start, "/")
	for _, v := range strs {
		if v == "" {
			continue
		}
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		child, err := master.Child(uint32(i) + hdkeychain.HardenedKeyStart)
		if err != nil {
			return nil, err
		}
		master = child
	}
	result := make([]string, count)
	for i := 1; i <= count; i++ {
		child, err := master.Child(uint32(i) + hdkeychain.HardenedKeyStart)
		if err != nil {
			return nil, err
		}
		if isSigWit {
			pubKey, err := child.ECPubKey()
			if err != nil {
				return nil, err
			}
			addr, err := GetWitnessAddress(pubKey, net)
			if err != nil {
				return nil, err
			}
			result[i-1] = addr
		} else {
			pubHash, err := child.Address(net)
			if err != nil {
				return nil, err
			}
			result[i-1] = pubHash.EncodeAddress()
		}
	}
	return result, nil
}

// GetWitnessAddress get witness address
func GetWitnessAddress(pubKey *btcec.PublicKey, net *chaincfg.Params) (string, error) {
	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
	witAddr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
	witnessProgram, err := txscript.PayToAddrScript(witAddr)
	if err != nil {
		return "", err
	}
	address, err := btcutil.NewAddressScriptHash(witnessProgram, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	return address.EncodeAddress(), nil
}
