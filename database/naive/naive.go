package naive

type NaiveDB struct {
	bloomFilter map[string]bool
	bst         *BST
	avl         *AVL
	callback    func(data any, params any) (any, any, error)
}

func NewDB(callback func(data any, params any) (any, any, error)) *NaiveDB {
	// Start the go routines that manage the BST and AVL trees, etc

	// Start the go routine that listens on the AVL channel and manages the AVL tree

	// We need a parent thread that monitors the BST's memory usage, and the balance factor metric
	// If the memory usage is too high, we need to block, swap out the BST and AVL trees, and write the AVL tree to disk
	// If the balance factor is too high, we need to block, swap out the BST and AVL trees, and instantiate a new AVL tree

	bloomFilter := make(map[string]bool)
	bst := NewBST()
	avl := NewAVL()

	// TODO we need to save the channel for the AVL tree in this struct
	return &NaiveDB{
		bloomFilter: bloomFilter,
		bst:         bst,
		avl:         avl,
		callback:    callback,
	}
}

// TODO: implement
func (db *NaiveDB) checkBloomFilter(key string) bool {
	return false
}

// TODO: implement
func (db *NaiveDB) Calculate(key string, params any) (any, error) {
	// PART 1 -----------------------------------------------------------------
	// Insert/Search and compute the result if it exists in the BST

	// TODO Hack
	db.bst.Insert(key)

	node := db.bst.Search(key)

	if node == nil {
		return false, nil
	}

	// We must absolutely unlock the node before we return
	defer node.lock.Unlock()

	data, result, err := db.callback(node.data, params)

	node.data = data

	if err != nil {
		return false, err
	}
	return result, nil

	// PART 2 -----------------------------------------------------------------
	// TODO: do we really need a bloom filter?
	// Check if the key is in the bloom filter

	// If it is, check the BST
	// If it is in the BST, find the node, calculcate the rate limiting, and return the result
	// push the key and node data into the avl queue (do this in a separate go routine)
	// push the key into the bloom filter (do this in a separate go routine)

	// For now, if it isn't in the bloom filter, we will insert the key into the BST
	// push the key and node data into the avl queue (do this in a separate go routine)
	// push the key into the bloom filter (do this in a separate go routine)

	// TODO FileSystem Part
	// Otherwise, check the filesystem
	// If it is in the filesystem, find the node, calculcate the rate limiting, and return the result
	// then call insert on the BST and push the key and node data into the avl queue (do this in a separate go routine)
	// push the key into the bloom filter (do this in a separate go routine)

	// Otherwise, insert the key into the BST and
	// push the key and node data into the avl queue (do this in a separate go routine)
	// push the key into the bloom filter (do this in a separate go routine)
}
