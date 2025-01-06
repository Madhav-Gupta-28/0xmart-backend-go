# 0xmart Blockchain Event Listener

A Go-based service that listens to smart contract events and stores them in MongoDB.

## Prerequisites

- Go 1.19 or higher
- MongoDB 6.0 or higher
- Alchemy API key (or other Web3 provider)

## Installation Steps

### 1. Install MongoDB (macOS)

### 2. Install Dependencies

### 3. Configure Environment
Create a `.env` file in the root directory:
```

## Running the Application
    go run main.go

1. **Start MongoDB** (in a separate terminal)
    sudo brew services start mongodb-community


2. # Connect to MongoDB
mongosh

# Switch to your database
use 0xmart

# View all transactions
db.transactions.find()

# View in a prettier format
db.transactions.find().pretty()

# Count total transactions
db.transactions.countDocuments()