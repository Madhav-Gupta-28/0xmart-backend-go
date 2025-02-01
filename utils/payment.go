package utils

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type PaymentProcessor struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
}

func NewPaymentProcessor(rpcURL, privateKeyHex string) (*PaymentProcessor, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %v", err)
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	return &PaymentProcessor{
		client:     client,
		privateKey: privateKey,
	}, nil
}

func (p *PaymentProcessor) ProcessPayment(toAddress string, amount *big.Int) (*types.Transaction, error) {
	ctx := context.Background()

	publicKey := p.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := p.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	gasPrice, err := p.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	to := common.HexToAddress(toAddress)
	gasLimit := uint64(21000)

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	chainID, err := p.client.NetworkID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain id: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), p.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	err = p.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	return signedTx, nil
}
