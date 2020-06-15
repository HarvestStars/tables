package main

import (
	"testing"
)

func TestAddr2ScriptHash(t *testing.T) {
	addr := "3EYPj1wgUxz82DPUJTbaeq4ipdF3WmewQK"
	want := "bf88657b58aa06c29545f27fc0163e6daf84742a7dc5f3ed26145fcea128de18"
	if Addr2ScriptHash(addr) != want {
		t.Error("get script hash failure")
	}
}

func TestDecodeRawTransaction(t *testing.T) {
	raw := "0200000001a4369df75a0b3695c518ede6e667a30fc3912aa43fd6ed6755b52eaed51981e6000000006a47304402202bf78149df1d8d77446f66eeebe212d21dfdf8f1775cd7fdf7f5900ef7fa44940220405902c918c0619421b6cda4afe239fe8c7b83d9bc33b86c2a3837d4e678cc4301210334135cbf970741851b20cda74d26fc71ab50455bfd087b25a91c2afc621acc83feffffff034def0b7a0b0000001976a91400023145087e223d10f3af05b60c067aea75e68e88acab11e30eda00000017a9145655cb60b59a7bf4e78bf9cb0abb9521ddd11592870000000000000000226a511f03ff9701b17576a91400023145087e223d10f3af05b60c067aea75e68e88acdd920100"
	long := AddressBalance{
		Addr:    "113dLR9a4pa1qeQtggQQ5c1j5wpVqhC1a",
		Balance: 0,
	}
	short := AddressBalance{
		Addr:    "39ZWp9Ha2VrStMgML27aLsXneWbXzQETjR",
		Balance: 0,
	}

	err := DecodeRawTransaction(raw, &long, &short)
	if err != nil {
		t.Error("boom")
	}
	if long.Balance != 0 || short.Balance != 936552632747 {
		t.Error("transaction decode error")
	}
}
