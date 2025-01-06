package utils

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/database"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/models"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockchainEventListener struct {
	client      *ethclient.Client
	isListening bool
}

func NewBlockchainEventListener() *BlockchainEventListener {
	return &BlockchainEventListener{}
}

func (b *BlockchainEventListener) Start() error {
	if b.isListening {
		return errors.New("already listening")
	}

	websocketURL := os.Getenv("WEB3_WEBSOCKET_URL")
	contractAddress := os.Getenv("CONTRACT_ADDRESS")

	client, err := ethclient.Dial(websocketURL)
	if err != nil {
		log.Println("Failed to connect to the Ethereum client:", err)
		return err
	}

	b.client = client
	b.isListening = true

	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(contractAddress)},
	}

	logs := make(chan types.Log)
	sub, err := b.client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Printf("Failed to subscribe to contract events: %v", err)
		return err
	}

	log.Printf("🎉 Successfully connected to contract at: %s", contractAddress)
	log.Println("👂 Listening for contract events...")

	go func() {
		for {
			select {
			case err := <-sub.Err():
				log.Printf("❌ Subscription error: %v", err)
				b.Restart()
			case vLog := <-logs:
				// Parse the event data
				tx := b.parseTransactionEvent(vLog)
				if tx != nil {
					// Store in MongoDB
					collection := database.DB.Collection("transactions")
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					_, err := collection.InsertOne(ctx, tx)
					cancel()

					if err != nil {
						log.Printf("Failed to store transaction: %v", err)
						continue
					}

					log.Printf("\n✅ Stored New Transaction")
					log.Printf("Order ID: %d", tx.OrderID)
					log.Printf("Customer: %s", tx.CustomerAddress)
					log.Printf("Amount: %s", tx.Amount)
					log.Printf("Status: %s", tx.Status)
					log.Println("----------------------------------------")
				}
			}
		}
	}()

	return nil
}

func (b *BlockchainEventListener) parseTransactionEvent(vLog types.Log) *models.Transaction {
	// Assuming the event has orderId, customerAddress, and amount as parameters
	if len(vLog.Data) < 96 { // 3 * 32 bytes for three parameters
		return nil
	}

	// Parse the event data
	orderID := new(big.Int).SetBytes(vLog.Data[:32]).Uint64()
	customerAddress := common.BytesToAddress(vLog.Data[32:64]).Hex()
	amount := hex.EncodeToString(vLog.Data[64:96])

	return &models.Transaction{
		OrderID:         orderID,
		CustomerAddress: customerAddress,
		Amount:          amount,
		Timestamp:       time.Now(),
		Status:          "completed",
	}
}

func (b *BlockchainEventListener) Restart() error {
	if !b.isListening {
		return errors.New("not currently listening")
	}

	log.Println("🔄 Restarting blockchain event listener...")
	b.isListening = false
	if b.client != nil {
		b.client.Close()
	}
	time.Sleep(5 * time.Second)
	return b.Start()
}
