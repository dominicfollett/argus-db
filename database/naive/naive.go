package naive

import (
	"context"
	"log/slog"
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

		db.logger.Debug("Starting AVL goroutine")
		// We'll use channel closure to signal the end of the go routine
		for message := range db.avlChannel {
			db.rwLock.RLock()

			db.avl.Insert(message.key, message.data)
			db.logger.Debug("avl goroutine, successfully inserted", "key", message.key)

			// TODO: consider running a function that removes records that are obsolete
			db.rwLock.RUnlock()
		}
		db.logger.Debug("Channel closed. Exiting AVL goroutine.")
	}()

	// Start the switchover goroutine
	db.wg.Add(1)
	go func() {
		defer db.wg.Done()

		db.logger.Debug("Starting switchover routine")
		for {
			select {
			case <-context.Done():
				db.logger.Debug("Switchover routine stopped. Exiting")
				return
			default:
				triggerMetric := float64(db.bst.balanceFactorSum.Load()) / float64(db.totalOps.Load())

				if triggerMetric > 5 {
					db.logger.Debug("switchover routine, threshold exceeded", "trigger metric", triggerMetric)

					// Obtain the r/w lock to ensure everyone's work is done
					db.rwLock.Lock()
					db.logger.Debug("switchover routine, naive db lock obtained")

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
					db.logger.Debug("switchover routine, naive db lock released")
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
	db.logger.Debug("naive db shutdown commencing...")

	// Wait until everyone has released the r/w lock / finished their operations
	db.rwLock.Lock()
	defer db.rwLock.Unlock()
	db.logger.Debug("naive db lock obtained")

	// Signal to the switchover goroutine to stop
	db.logger.Debug("terminating the switchover routine")
	db.stopRoutine()

	// Close the channel to signal the end of the goroutine
	db.logger.Debug("terminating the avl routine")
	close(db.avlChannel)

	db.wg.Wait()
	db.logger.Debug("naive db routines ended")
	db.logger.Info("naive db shutdown complete")
}

func (db *NaiveDB) Calculate(key string, params any) (any, error) {

	db.rwLock.RLock()
	// We must absolutely unlock the r/w lock before we return
	defer db.rwLock.RUnlock()
	db.logger.Debug("naive db calculate, lock obtained")

	node := db.bst.InSearch(key)
	// We must absolutely unlock the node before we return
	defer node.lock.Unlock()
	db.logger.Debug("naive db calculate, node found")

	// Apply the callback defined by the user of this DB
	data, result, err := db.callback(node.data, params)
	if err != nil {
		return false, err
	}
	db.logger.Debug("naive db calculate, callback function successfully applied")

	// Update the node's data
	node.data = data

	// Publish the message to the avlChannel for the goroutine to pick up
	db.avlChannel <- Message{key: key, data: data}
	db.logger.Debug("naive db calculate, avl message published")

	// Increment the totalOps counter
	db.totalOps.Add(1)

	db.logger.Debug("naive db calculate", "total operations", db.totalOps.Load())
	return result, nil
}
