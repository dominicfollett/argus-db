package naive

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type NaiveDB struct {
	bst         *BST
	avl         *AVL
	callback    func(data any, params any) (any, any, error)
	avlChannel  chan Message
	rwLock      *sync.RWMutex
	totalOps    *atomic.Int64
	stopRoutine context.CancelFunc
}

// TODO: Add logging
func NewDB(callback func(data any, params any) (any, any, error)) *NaiveDB {
	totalOps := atomic.Int64{}
	totalOps.Store(1)

	context, cancel := context.WithCancel(context.Background())

	db := &NaiveDB{
		bst:         NewBST(),
		avl:         NewAVL(),
		callback:    callback,
		avlChannel:  make(chan Message),
		rwLock:      &sync.RWMutex{},
		totalOps:    &totalOps,
		stopRoutine: cancel,
	}

	// Start the AVL goroutine
	go func() {
		fmt.Println("Starting AVL goroutine")
		// We'll use channel closure to signal the end of the go routine
		// TODO: consider using a context here instead?
		for message := range db.avlChannel {
			db.rwLock.RLock()
			db.avl.Insert(message.key, message.data)
			db.rwLock.RUnlock()
		}
		fmt.Println("Channel closed. Exiting AVL goroutine.")
	}()

	// Start the switchover goroutine
	// Add context to this goroutine so that it can be stopped
	go func() {
		fmt.Println("Starting switchover routine")
		for {
			select {
			case <-context.Done():
				fmt.Println("Switchover routine stopped. Exiting")
				return
			default:
				triggerMetric := float64(db.bst.balanceFactorSum.Load()) / float64(db.totalOps.Load())

				if triggerMetric > 5 {
					fmt.Printf("[switchover routine] Trigger metric: %f\n", triggerMetric)

					// Obtain the r/w lock to ensure everyone's work is done
					db.rwLock.Lock()

					// Swap out the BST and AVL trees
					// What happens to the old BST?
					db.bst.root = db.avl.root

					// Create a new AVL tree
					db.avl = NewAVL()

					// TODO: is there a way to clear the avlChannel?

					// Reset the balance factor sum
					db.bst.balanceFactorSum.Store(0)

					// Reset the totalOps counter
					db.totalOps.Store(1)

					// Release the r/w lock
					db.rwLock.Unlock()
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()

	return db
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
	db.stopRoutine()

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

	// Increment the totalOps counter
	db.totalOps.Add(1)

	fmt.Println("Total operations:", db.totalOps.Load())

	return result, nil
}
