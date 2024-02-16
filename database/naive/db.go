package naive

import (
	"bytes"
	"sync"
	"time"
)

type node struct {
	Key *[]byte // Should/Can we make this a fixed size? And why am I using a pointer?
	Height int
	Tokens int  // TODO Revise
	Time string // TODO Revise
	Left  *node
	Right *node
	Lock  *sync.Mutex // Synchronizes access to the 'values'
	BalanceLock *sync.Mutex 
}

type DB struct {
	AVL  *node

	// TBH this is probably only useful if we're taking a snapshot of the tree
	AVLWG   sync.WaitGroup    // Tracks threads accessing the AVL Tree
    AVLRWMutex  sync.RWMutex  // Controls access to the AVL Tree
}

func NewDatabase() *DB {
	return &DB{
		AVL: &node{
			Key: nil,
			Lock: &sync.Mutex{},
		},
		AVLWG: sync.WaitGroup{},
		AVLRWMutex: sync.RWMutex{}, 
	}
}

func (db *DB) Search(key []byte) error {
	// TODO Implement
	return nil
}

// Pre Order Traverses the tree and prints out the keys
func (db *DB) Print() {

}

/*
TODO:
- The goal is to keep the tree balanced, not to balance after every single insertion.
- Rotations only occur when a node's balance factor is violated post-insertion.

TODO
Better  Approaches for Concurrent AVL Trees:
Fine-Grained Locking: Employ more fine-grained locking strategies. 
Lock only the specific subtrees directly involved in an insertion or rebalancing operation. 
This allows more concurrent changes  in non-overlapping sections of the tree.

Also look at:
Lock-Free/Wait-Free Data Structures:  Explore advanced lock-free or wait-free data structures specifically designed for concurrent scenarios. These minimize or eliminate the reliance on traditional locks, though their implementation is often more complex.
Read-Write Locks: Employ  read-write locks (often called shared/exclusive locks) for better concurrency if you have many read operations interleaved with fewer modifications. Read locks allow multiple threads to hold locks simultaneously.
*/
func (n *node) __insert(key []byte, refill int, capacity int) error {
	
	// Insert
	if n.Key == nil {
		n.Lock.Lock()

		// Basically we've exited the blocking but n.Key might no longer be nil
		// so we need to check again
		if n.Key != nil {
			direction := bytes.Compare(*n.Key, key)

			// TODO convert to case
			// *n.Key == Key
			if direction == 0 {
				// Update -- we already have the lock
		
				// TODO do some stuff
				println("Okay")
		
				n.Lock.Unlock()
				return nil
			}

			n.Lock.Unlock()
		
			// *n.Key > Key
			if direction == 1 {
				return n.Right.__insert(key, refill, capacity)
			}
		
			// *n.Key < Key
			if direction == -1 {
				return n.Left.__insert(key, refill, capacity)
			}
		}

		n.Key = &key
		n.Left = &node{
			Key: nil,
			Lock: &sync.Mutex{},
		}
		n.Right = &node{
			Key: nil,
			Lock: &sync.Mutex{},
		}
		n.Tokens = 10
		n.Time = time.Now().String()

		n.Lock.Unlock()
		return nil
	}

	direction := bytes.Compare(*n.Key, key)

	// TODO convert to case
	// *n.Key == Key
	if direction == 0 {
		// Update
		n.Lock.Lock()

		// TODO do some stuff
		println("Okay")

		n.Lock.Unlock()
		return nil
	}

	// *n.Key > Key
	if direction == 1 {
		return n.Right.__insert(key, refill, capacity)
	}

	// *n.Key < Key
	if direction == -1 {
		return n.Left.__insert(key, refill, capacity)
	}

	return nil
}


func (db *DB) Insert(key []byte, refill int, capacity int) error {
	// TODO rethink this AVL tree locking
	//defer db.AVLWG.Done()
	
	//db.AVLWG.Add(1)
	//db.AVLRWMutex.RLock()
	
	err := db.AVL.__insert(key, refill, capacity)
	
	//db.AVLRWMutex.RUnlock()

	return err
}