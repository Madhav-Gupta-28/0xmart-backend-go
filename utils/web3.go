package utils

import (
	"context"
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

	log.Printf("🔌 Connecting to WebSocket: %s", websocketURL)
	log.Printf("📝 Watching contract: %s", contractAddress)

	client, err := ethclient.Dial(websocketURL)
	if err != nil {
		log.Printf("❌ Failed to connect to Ethereum client: %v", err)
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
		log.Printf("❌ Failed to subscribe to contract events: %v", err)
		return err
	}

	log.Printf("✅ Successfully connected to contract")
	log.Println("👂 Listening for contract events...")

	go func() {
		for {
			select {
			case err := <-sub.Err():
				log.Printf("❌ Subscription error: %v", err)
				b.Restart()
			case vLog := <-logs:
				log.Printf("📥 Received event: %+v", vLog)

				// Parse the event data
				tx := b.parseTransactionEvent(vLog)
				if tx != nil {
					log.Printf("📦 Parsed transaction: %+v", tx)

					// Store in MongoDB
					collection := database.DB.Collection("transactions")
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					result, err := collection.InsertOne(ctx, tx)
					cancel()

					if err != nil {
						log.Printf("❌ Failed to store transaction: %v", err)
						continue
					}

					log.Printf("✅ Stored transaction with ID: %v", result.InsertedID)
					log.Printf("📊 Transaction Details:")
					log.Printf("   Order ID: %d", tx.OrderID)
					log.Printf("   Customer: %s", tx.CustomerAddress)
					log.Printf("   Amount: %s", tx.Amount)
					log.Printf("   Status: %s", tx.Status)
					log.Println("----------------------------------------")
				} else {
					log.Printf("⚠️ Failed to parse event data")
				}
			}
		}
	}()

	return nil
}

func (b *BlockchainEventListener) parseTransactionEvent(vLog types.Log) *models.Transaction {
	log.Printf("🔍 Parsing event data")
	log.Printf("📑 Event topics: %v", vLog.Topics)

	// We need at least 4 topics: event signature and 3 indexed parameters
	if len(vLog.Topics) < 4 {
		log.Printf("⚠️ Not enough topics: %d", len(vLog.Topics))
		return nil
	}

	// Parse from Topics instead of Data
	// Topics[0] is event signature
	// Topics[1] is customer address
	// Topics[2] is order ID
	// Topics[3] is amount

	customerAddress := common.HexToAddress(vLog.Topics[1].Hex()).Hex()
	orderID := new(big.Int).SetBytes(vLog.Topics[2][:]).Uint64()
	amount := new(big.Int).SetBytes(vLog.Topics[3][:]).String()

	log.Printf("📝 Parsed values:")
	log.Printf("   Order ID: %d", orderID)
	log.Printf("   Customer: %s", customerAddress)
	log.Printf("   Amount: %s", amount)

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
