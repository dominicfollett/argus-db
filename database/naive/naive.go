package naive

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

var TRIGGER_THRESHOLD float64 = 30

type NaiveDB struct {
	bst         *BST
	avl         *AVL
	callback    func(data any, params any) (any, any, error)
	avlChannel  chan Message
	rwLock      *sync.RWMutex
	avlLock     *sync.Mutex
	totalOps    *atomic.Int64
	stopRoutine context.CancelFunc
	wg          *sync.WaitGroup
	logger      *slog.Logger
}

// TODO: Add logging
func NewDB(callback func(data any, params any) (any, any, error), logger *slog.Logger) *NaiveDB {
	logger.Info("initializing naive DB...")

	totalOps := atomic.Int64{}
	totalOps.Store(1)

	context, cancel := context.WithCancel(context.Background())

	db := &NaiveDB{
		bst:         NewBST(),
		avl:         NewAVL(),
		callback:    callback,
		avlChannel:  make(chan Message),
		avlLock:     &sync.Mutex{},
		rwLock:      &sync.RWMutex{},
		totalOps:    &totalOps,
		stopRoutine: cancel,
		wg:          &sync.WaitGroup{},
		logger:      logger,
	}

	db.wg.Add(1)
	// Start the AVL goroutine
	go func() {
		defer db.wg.Done()

		db.logger.Info("Starting AVL goroutine")
		// We'll use channel closure to signal the end of the go routine
		for message := range db.avlChannel {
			db.avlLock.Lock()

			// if action == "delete" {
			// 	db.avl.Delete(message.key)
			// } else {
			// 	db.avl.Insert(message.key, message.data)
			// }
			db.avl.Insert(message.key, message.data)

			// TODO: consider running a function that removes records that are obsolete
			db.avlLock.Unlock()
		}
		db.logger.Info("Channel closed. Exiting AVL goroutine.")
	}()

	// Start the switchover goroutine
	db.wg.Add(1)
	go func() {
		defer db.wg.Done()

		db.logger.Info("Starting switchover routine")
		for {
			select {
			case <-context.Done():
				db.logger.Info("Switchover routine stopped. Exiting")
				return
			default:
				triggerMetric := float64(db.bst.balanceFactorSum.Load()) / float64(db.totalOps.Load())

				if triggerMetric > TRIGGER_THRESHOLD {
					db.logger.Debug("switchover routine, threshold exceeded", "trigger metric", triggerMetric)

					// Obtain the r/w lock to pause calculations
					db.rwLock.Lock()

					// Obtain the avl lock to pause avl inserts
					db.avlLock.Lock()
					db.logger.Debug("switchover routine, naive db locks obtained")

					// Swap out the BST and AVL trees
					// What happens to the old BST?
					db.bst.root = db.avl.root
					db.logger.Debug("switchover routine, tree successfully replaced")

					// Create a new AVL tree
					db.avl = NewAVL()

					// TODO: is there a way to clear the avlChannel?

					// Reset the balance factor sum
					db.bst.balanceFactorSum.Store(0)

					// Reset the totalOps counter
					db.totalOps.Store(1)
					db.logger.Debug("switchover routine, metrics reset")

					// Release the r/w lock
					db.rwLock.Unlock()
					// Release the avl lock
					db.avlLock.Unlock()

					db.logger.Info("switchover routine, naive db locks released")
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()

	return db
}

// Shutdown closes the AVL channel to stop the AVL goroutine
// it calls the cancel function 'stopRoutine' to tell the switchover routine to exit
func (db *NaiveDB) Shutdown() {
	// Wait until everyone has released the r/w lock / finished their operations
	db.rwLock.Lock()
	defer db.rwLock.Unlock()

	// Signal to the switchover goroutine to stop
	db.logger.Info("terminating the switchover routine")
	db.stopRoutine()

	// Close the channel to signal the end of the goroutine
	db.logger.Info("terminating the avl routine")
	close(db.avlChannel)

	db.wg.Wait()
	db.logger.Info("naive db shutdown complete")
}

func (db *NaiveDB) Calculate(key string, params any) (any, error) {
	db.rwLock.RLock()
	// We must absolutely unlock the r/w lock before we return
	defer db.rwLock.RUnlock()

	node := db.bst.InSearch(key)
	// We must absolutely unlock the node before we return
	defer node.lock.Unlock()

	// Apply the callback defined by the user of this DB
	data, result, err := db.callback(node.data, params)
	if err != nil {
		db.logger.Info("naive db calculate, callback function failed", "error", err)
		return false, err
	}

	// Update the node's data
	node.data = data

	// Publish the message to the avlChannel for the goroutine to pick up
	db.avlChannel <- Message{key: key, data: data}

	// Increment the totalOps counter
	db.totalOps.Add(1)
	return result, nil
}
