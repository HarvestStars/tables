package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/HarvestStars/tables/gredis"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/urfave/cli"
)

// NewApp new application
func NewApp() *cli.App {
	app := &cli.App{
		Commands: []cli.Command{
			{
				Name:  "initdb",
				Usage: "init database",
				Action: func(c *cli.Context) error {
					host := c.Args().First()
					if err := gredis.Setup(host, "123456"); err != nil {
						//fmt.Println("redis error with ", host)
						return err
					}

					count, err := strconv.Atoi(c.Args().Get(1))
					if err != nil {
						//fmt.Println(count, " error count")
						return err
					}
					if count%2 != 0 {
						//fmt.Println("count must be even")
						return nil
					}
					masterKey := c.Args().Get(2)
					addrs, err := GenerateAddress(masterKey, count, "/0/0", &chaincfg.MainNetParams, true)
					if err != nil {
						//fmt.Println("generate address error with ", masterKey)
						return err
					}
					//slot_index_long
					//slot_index_short
					index := 0
					for index < count/2 {
						longAddr := addrs[2*index]
						shortAddr := addrs[2*index+1]
						longKey := "slot_" + strconv.Itoa(index) + "_long"
						shortKey := "slot_" + strconv.Itoa(index) + "_short"
						index++
						fmt.Printf("set %s %s \n", longKey, longAddr)
						fmt.Printf("set %s %s \n", shortKey, shortAddr)
						err = gredis.Set(longKey, longAddr, 0)
						if err != nil {
							//fmt.Printf("longKey写入错误, %s \n", err)
						}
						err = gredis.Set(shortKey, shortAddr, 0)
						if err != nil {
							//fmt.Printf("shortKey写入错误, %s \n", err)
						}
					}
					return nil
				},
			},
			{
				Name:  "initdb2",
				Usage: "init database",
				Action: func(c *cli.Context) error {
					host := c.Args().First()
					if err := gredis.Setup(host, "123456"); err != nil {
						fmt.Println("redis error with ", host)
						return err
					}

					count, err := strconv.Atoi(c.Args().Get(1))
					if err != nil {
						fmt.Println(count, " error count")
						return err
					}
					if count%2 != 0 {
						fmt.Println("count must be even")
						return nil
					}
					masterKey := c.Args().Get(2)
					addrs, err := GenerateAddress(masterKey, count, "/0/0", &chaincfg.MainNetParams, true)
					if err != nil {
						fmt.Println("generate address error with ", masterKey)
						return err
					}
					//slot_index_long
					//slot_index_short
					index := 0
					for index < count/2 {
						index++
						longAddr := addrs[index]
						shortAddr := addrs[index+1]
						longKey := "height_" + strconv.Itoa(index+103212) + "_single"
						shortKey := "height_" + strconv.Itoa(index+103212) + "_double"

						gredis.Set(longKey, longAddr, 0)
						gredis.Set(shortKey, shortAddr, 0)
					}
					return nil
				},
			},
		},
	}
	app.Usage = "tables manager client"

	return app
}

func main() {
	app := NewApp()
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
