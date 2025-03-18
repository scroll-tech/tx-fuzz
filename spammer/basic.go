package spammer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const TX_TIMEOUT = 50 * time.Minute

func SendBasicTransactions(config *Config, key *ecdsa.PrivateKey, f *filler.Filler) error {
	backend := ethclient.NewClient(config.backend)
	sender := crypto.PubkeyToAddress(key.PublicKey)
	chainID, err := backend.ChainID(context.Background())
	if err != nil {
		log.Println("Could not get chainID, using default")
		chainID = big.NewInt(0x01000666)
	}
	allCount := config.N

	var lastTx *types.Transaction
	for i := uint64(0); i < allCount; i++ {
		nonce, err := backend.NonceAt(context.Background(), sender, big.NewInt(-1))
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		tx, err := txfuzz.RandomValidTx(config.backend, f, sender, nonce, nil, nil, config.accessList)
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		signedTx, err := types.SignTx(tx, types.NewCancunSigner(chainID), key)
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if err := backend.SendTransaction(context.Background(), signedTx); err != nil {
			log.Println("Could not send transaction", err)
			time.Sleep(10 * time.Millisecond)
			continue
		}
		lastTx = signedTx
		time.Sleep(10 * time.Millisecond)
	}
	log.Println("send over")
	if lastTx != nil {
		ctx, cancel := context.WithTimeout(context.Background(), TX_TIMEOUT)
		defer cancel()
		if _, err := bind.WaitMined(ctx, backend, lastTx); err != nil {
			fmt.Printf("Waiting for transactions to be mined failed: %v\n", err.Error())
		}
	}
	log.Println("Finished sending basic transactions")
	return nil
}
