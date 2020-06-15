package main

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
)

func TestImportMasterPubKey(t *testing.T) {
	master := "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8"
	wantAddr := "39SQGX6HB2bgS74p6bp144RjR4SRxfkKzH"
	extKey, err := hdkeychain.NewKeyFromString(master)
	if err != nil {
		t.Error(err)
	}

	if extKey.IsPrivate() {
		t.Error("key is private")
	}
	path := []uint32{
		'0', '0', '1',
	}
	for _, childNum := range path {
		extKey, _ = extKey.Child(childNum)
	}
	//addr, err := extKey.Address(&chaincfg.MainNetParams)
	//pubKey, _ := extKey.ECPubKey()
	//hash160 := btcutil.Hash160(pubKey.SerializeUncompressed())
	// hash160 = []byte{
	// 	0xe8, 0xc3, 0x00, 0xc8, 0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37,
	// 	0xc0, 0x51, 0x99, 0x29, 0x01, 0x9e, 0xf8, 0x6e, 0xb5, 0xb4}
	wif, _ := btcutil.DecodeWIF("Kyd5id9rUoZSvBdPmHkXqaEHkTWtZ9MLF32kVeAqapyjbbmw6zB3")
	pub2 := wif.PrivKey.PubKey()
	hash160 := btcutil.Hash160(pub2.SerializeUncompressed())
	addr, _ := btcutil.NewAddressScriptHashFromHash(hash160, &chaincfg.MainNetParams)
	addr1, _ := GetWitnessAddress(pub2, &chaincfg.MainNetParams)
	if addr.EncodeAddress() != wantAddr {
		t.Log(addr1)
		t.Error("error address")
	}
}

func TestImportMasterPubKey2(t *testing.T) {
	master := "xprv9s21ZrQH143K3i56kWKS4GjoEFKW6ctQeofrzvCTGys6RjEpEGoeeSQQ9oNJncTNm55P9V4u53qK2SvcEzs1SgFMoJwT5oKcqLXkt3LGTWt"
	wantAddr := "39SQGX6HB2bgS74p6bp144RjR4SRxfkKzH"
	addrs, _ := GenerateAddress(master, 1, "/0/0", &chaincfg.MainNetParams, true)
	if addrs[0] != wantAddr {
		t.Error("error address, tag=", addrs[0])
	}
}
