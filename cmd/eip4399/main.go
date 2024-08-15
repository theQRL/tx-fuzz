package main

import (
	"context"
	"fmt"

	txfuzz "github.com/rgeraldes24/tx-fuzz"
	"github.com/theQRL/go-qrllib/dilithium"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/core/types"
	"github.com/theQRL/go-zond/rpc"
	"github.com/theQRL/go-zond/zondclient"
)

var (
	address = "http://127.0.0.1:8545"
)

func main() {
	cl, acc := getRealBackend()
	backend := zondclient.NewClient(cl)
	sender := common.HexToAddress(txfuzz.ADDR)
	nonce, err := backend.PendingNonceAt(context.Background(), sender)
	if err != nil {
		panic(err)
	}
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Nonce: %v\n", nonce)

	gasTipCap, _ := backend.SuggestGasTipCap(context.Background())
	gasFeeCap, _ := backend.SuggestGasPrice(context.Background())
	tx := types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonce,
		Value:     common.Big1,
		Gas:       500000,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Data:      []byte{0x44, 0x44, 0x55},
	})
	signedTx, _ := types.SignTx(tx, types.NewShanghaiSigner(chainid), acc)
	backend.SendTransaction(context.Background(), signedTx)
}

func getRealBackend() (*rpc.Client, *dilithium.Dilithium) {
	// eth.sendTransaction({from:personal.listAccounts[0], to:"0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D", value: "100000000000000"})

	acc, err := dilithium.NewDilithiumFromHexSeed(txfuzz.SEED[2:])
	if err != nil {
		panic(err)
	}
	if addr := common.Address(acc.GetAddress()); addr.Hex() != txfuzz.ADDR {
		panic(fmt.Sprintf("wrong address want %s got %s", addr.Hex(), txfuzz.ADDR))
	}

	cl, err := rpc.Dial(address)
	if err != nil {
		panic(err)
	}
	return cl, acc
}
