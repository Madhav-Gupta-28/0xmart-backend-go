package utils

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockchainEventListener struct {
	client      *ethclient.Client
	isListening bool
}

// NewBlockchainEventListener creates a new BlockchainEventListener
func NewBlockchainEventListener() *BlockchainEventListener {
	return &BlockchainEventListener{}
}

// Start begins listening for blockchain events
func (b *BlockchainEventListener) Start() error {
	if b.isListening {
		return errors.New("already listening")
	}

	websocketURL := os.Getenv("WEB3_WEBSOCKET_URL")
	client, err := ethclient.Dial(websocketURL)
	if err != nil {
		log.Println("Failed to connect to the Ethereum client:", err)
		return err
	}

	b.client = client
	b.isListening = true

	// Example: Subscribe to new blocks
	headers := make(chan *types.Header)
	sub, err := b.client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Println("Failed to subscribe to new blocks:", err)
		return err
	}

	go func() {
		for {
			select {
			case err := <-sub.Err():
				log.Println("Subscription error:", err)
				b.Restart()
			case header := <-headers:
				log.Println("New block:", header.Number.String())
			}
		}
	}()

	return nil
}

// Restart stops and starts the listener
func (b *BlockchainEventListener) Restart() error {
	if !b.isListening {
		return errors.New("not currently listening")
	}

	log.Println("Restarting blockchain event listener...")
	b.isListening = false
	time.Sleep(5 * time.Second)
	return b.Start()
}
