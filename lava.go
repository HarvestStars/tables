package main

import (
	"encoding/json"
	"time"

	"github.com/HarvestStars/go-electrum/electrum"
	"github.com/HarvestStars/tables/gredis"
	"github.com/HarvestStars/tables/logging"
	"github.com/HarvestStars/tables/setting"
)

var node *electrum.Node
var height int
var slot int
var slotOld int
var poolEntrySet []memPoolEntryWithHash

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
	poolEntrySet = make([]memPoolEntryWithHash, 0, 1000)
	slotOld = 0
	for {
		select {
		case <-stop:
			logging.GetLogger().Info("lava channel stop")
			stop <- 1
			return
		case header := <-headerChanel:
			updateHeader(header)
			liquidate(int(header.BlockHeight))
		case <-t.C:
			calcLongShort()
		}
	}
}

func updateHeader(header *electrum.BlockchainHeader) {
	height = int(header.BlockHeight)
	slot = height / setting.LavadBaseSetting.BlocksInSlot
	//write to redis
	info := struct {
		Height       int `json:"height"`
		Slot         int `json:"slot"`
		BlocksInSlot int `json:"blocksinslot"`
		DeadLine     int `json:"deadline"`
	}{
		Height:       height,
		Slot:         slot,
		BlocksInSlot: setting.LavadBaseSetting.BlocksInSlot,
		DeadLine:     setting.LavadBaseSetting.DeadLine,
	}
	data, _ := json.Marshal(&info)
	gredis.Set("blockchaininfo", string(data), 0)
	logging.GetLogger().Infof("update height:%d", height)
}
