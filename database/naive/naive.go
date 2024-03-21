package naive

import (
	"fmt"
	"sync"
)

type NaiveDB struct {
	bst        *BST
	avl        *AVL
	callback   func(data any, params any) (any, any, error)
	avlChannel chan Message
	rwLock     *sync.RWMutex
}

func NewDB(callback func(data any, params any) (any, any, error)) *NaiveDB {
	// We need a thread that monitors the BST's balance factor metric
	// If the balance factor is too high, we need to block, swap out the BST and AVL trees, and instantiate a new AVL tree

	bst := NewBST()
	avl := NewAVL()
	avlChannel := make(chan Message)
	rwLock := &sync.RWMutex{}

	go func() {
		// We'll use channel closure to signal the end of the go routine
		for message := range avlChannel {
			rwLock.RLock()
			avl.Insert(message.key, message.data)
			rwLock.RUnlock()
		}
		fmt.Println("Channel closed. Exiting AVL goroutine.")
	}()

	return &NaiveDB{
		bst:        bst,
		avl:        avl,
		callback:   callback,
		avlChannel: avlChannel,
		rwLock:     rwLock,
	}
}

// Shutdown closes the AVL channel and stops the AVL goroutine and ...
func (db *NaiveDB) Shutdown() {
	// Wait until everyone has released the r/w lock / finished their operations
	db.rwLock.Lock()
	defer db.rwLock.Unlock()

	// Close the channel to signal the end of the goroutine
	close(db.avlChannel)
	// Do we need to wait for the goroutine to finish?

	// Signal to the switchover goroutine to stop
	// Perhaps context.WithTimeout() and context.Done() would be useful here

	// Do we need to wait for the goroutine to finish?
}

func (db *NaiveDB) Calculate(key string, params any) (any, error) {

	// Obtain the r/w lock to ensure we're not in the middle of a tree swap
	db.rwLock.RLock()
	// We must absolutely unlock the r/w lock before we return
	defer db.rwLock.RUnlock()

	node := db.bst.InSearch(key)
	// We must absolutely unlock the node before we return
	defer node.lock.Unlock()

	// Apply the callback defined by the user of this DB
	data, result, err := db.callback(node.data, params)
	if err != nil {
		return false, err
	}

	// Update the node's data
	node.data = data

	// Publish the message to the avlChannel for the goroutine to pick up
	db.avlChannel <- Message{key: key, data: data}

	return result, nil
}
