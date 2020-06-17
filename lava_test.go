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
	raw := "0200000001521e6ce5c7631df51e0586a2cf7d9cd6fa98cb68b577bbef22b4fcf6f9b9ae39010000006b483045022100dc03c007859cdcc26540ffc3568c674f59453f1b927d39be71d5d8a38117168a022001daf5b51ca548a6909b7797053291b873c60bcddb2361561bd30fa81193fe20012103faf0f27e2c14b29d4b36f72448148a28c263aef1a2487d5131c9e0c47179e181feffffff0200e1f5050000000017a914d4aa6a77b97cc513073738280bbafb574b459997870c451d97090000001976a9142209ed74db5ec6d0b8a0d9fca756298d7e07cfd188acb8990100"
	long := AddressBalance{
		Addr:    "3M5VKv3aacea1q89i1TMtGCKuNEy6WWjhL",
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
	if long.Balance != 12 || short.Balance != 0 {
		t.Error("transaction decode error")
	}
}
