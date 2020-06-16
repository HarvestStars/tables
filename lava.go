package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/EasonZhao/tables/gredis"
	"github.com/EasonZhao/tables/logging"
	"github.com/EasonZhao/tables/setting"
	"github.com/HarvestStars/go-electrum/electrum"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

var node *electrum.Node
var height int
var slot int
var lavadClient *http.Client
var lavadIP string
var lavadPort string

const (
	signalAddr = "3EjxUF2knRQE3D6mXzDC4PoEnJX8gpZi2W"
	doubleAddr = "3CT55u55LMa1Hat1PJukUrMACugRuVJ1bW"
)

func run(stop chan interface{}) {
	logger := logging.GetLogger()
	logger.Info("lava channel start")

	//
	node = electrum.NewNode()
	if err := node.ConnectTCP(setting.ElectrumBaseSetting.Host); err != nil {
		logger.Error(err)
	}
	headerChanel, err := node.BlockchainHeadersSubscribe()
	if err != nil {
		logger.Error(err)
	}

	//timer
	d := time.Duration(time.Second * 10)
	t := time.NewTicker(d)
	defer t.Stop()

	for {
		select {
		case <-stop:
			logging.GetLogger().Info("lava channel stop")
			stop <- 1
			return
		case header := <-headerChanel:
			updateHeader(header)
		case <-t.C:
			calcLongShort()
		}
	}
}

func updateHeader(header *electrum.BlockchainHeader) {
	height = int(header.BlockHeight)
	slot = height/setting.LavadBaseSetting.BlocksInSlot + setting.LavadBaseSetting.AddrOffset
	//write to redis
	info := struct {
		Height int `json:"height"`
		Slot   int `json:"slot"`
	}{
		Height: height,
		Slot:   slot,
	}
	data, _ := json.Marshal(&info)
	gredis.Set("blockchaininfo", string(data), 0)
	logging.GetLogger().Infof("update height:%d", height)
}

func calcLongShort() {
	//
	beg := slot * setting.LavadBaseSetting.BlocksInSlot
	end := beg + setting.LavadBaseSetting.BlocksInSlot - 1
	longKey := "slot_" + strconv.Itoa(int(slot)) + "_long"
	shortKey := "slot_" + strconv.Itoa(int(slot)) + "_short"

	logger := logging.GetLogger()
	longAddr, err := gredis.Get(longKey)
	if err != nil {
		logger.Error(err)
	}
	shortAddr, err := gredis.Get(shortKey)
	if err != nil {
		logger.Error(err)
	}

	logger.Infof("calc long and short, slot index: %d", slot)
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

	// mempool中获取两个地址的txs
	// longMemPootTxs, shortMemPoolTxs, errMemPool := getLongShortMemPoolTxs(longAddr, shortAddr)
	// if errMemPool != nil {
	// 	logging.GetLogger().Error(errMemPool)
	// 	return
	// }

	// longTxs = append(longTxs, longMemPootTxs...)
	// shortTxs = append(shortTxs, shortMemPoolTxs...)

	txs := longTxs
	if len(txs) == 0 {
		txs = shortTxs
	} else {
		for _, v := range shortTxs {
			for _, tx := range txs {
				if v.Hash == tx.Hash {
					continue
				}
				txs = append(txs, v)
			}
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
		if tx.Height < beg || tx.Height > end {
			continue
		}
		raw, err := node.BlockchainTransactionGet(tx.Hash)
		if err != nil {
			logging.GetLogger().Error(err)
			continue
		}

		DecodeRawTransaction(raw, long, short)
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

func getLongShortMemPoolTxs(longAddr string, shortAddr string) ([]*electrum.Transaction, []*electrum.Transaction, error) {
	type lavaReqBody struct {
		JSONRPC string        `json:"jsonrpc"`
		Method  string        `json:"method"`
		Params  []interface{} `json:"params"`
		ID      int32         `json:"id"`
	}

	reqBody := lavaReqBody{JSONRPC: "1.0", Method: "getrawmempool", Params: []interface{}{}, ID: 1}
	reqByte, _ := json.Marshal(reqBody)
	fmt.Println(string(reqByte))
	req, err := http.NewRequest("POST", setting.LavadBaseSetting.Host,
		strings.NewReader(string(reqByte)))
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, err := lavadClient.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	type rawTx map[string]interface{}

	type rawResult struct {
		Result []string `json:"result"`
		Error  string   `json:"error"`
		ID     int      `json:"id"`
	}
	result := &rawResult{}

	json.Unmarshal(body, result)

	return nil, nil, nil
}

type tableOrder struct {
	Amount int64          `json:"total"`
	Long   AddressBalance `json:"long"`
	Short  AddressBalance `json:"short"`
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

// AddressBalance comment
type AddressBalance struct {
	Addr    string `json:"address"`
	Balance int64  `json:"amount"`
}

// DecodeRawTransaction comment
func DecodeRawTransaction(raw string, long *AddressBalance, short *AddressBalance) error {
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
