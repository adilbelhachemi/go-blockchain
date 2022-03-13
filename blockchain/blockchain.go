package blockchain

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
)

const dbPath = "./tmp/blocks"

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockchainIterator struct {
	currentHash []byte
	Database    *badger.DB
}

func InitBlockChain() *Blockchain {
	var lastHash []byte

	db, err := badger.Open(badger.DefaultOptions(dbPath))
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")

			genesis := Genesis()
			fmt.Println("Genesis proved")

			err = txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			lastHash, err = item.ValueCopy(nil)
			return err
		}
	})
	Handle(err)
	blockchain := Blockchain{
		LastHash: lastHash,
		Database: db,
	}

	return &blockchain
}

func (chain *Blockchain) AddBlock(data string) {
	var lastHash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)

		lastHash, err = item.ValueCopy(nil)
		Handle(err)

		return err
	})
	Handle(err)

	newBlock := CreateBlock(data, lastHash)
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash
		return err
	})
	Handle(err)
}

func (chain *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{
		currentHash: chain.LastHash,
		Database:    chain.Database,
	}
}

func (iter *BlockchainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.currentHash)
		encodedBlock, err := item.ValueCopy(nil)
		block = Deserialize(encodedBlock)

		return err
	})
	Handle(err)

	iter.currentHash = block.PrevHash
	return block
}
