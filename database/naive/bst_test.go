package naive

import (
	"strconv"
	"sync"
	"testing"
)

// TODO: Generated Tests -- revisit and refactor
func TestBSTConcurrentInserts(t *testing.T) {
	bst := NewBST() // Initialize your BST

	var wg sync.WaitGroup // Use WaitGroup to wait for all goroutines to finish
	const numInserts = 100 // Number of inserts per goroutine
	const concurrencyLevel = 10 // Number of goroutines

	// Perform concurrent inserts
	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numInserts; j++ {
				key := strconv.Itoa(goroutineID*numInserts + j) // Ensure unique keys
				bst.Insert(key, 1, 1) // Dummy tokens and capacity
			}
		}(i)
	}

	wg.Wait() // Wait for all goroutines to complete

	// After all inserts, check the size of the BST to ensure all inserts were successful
	expectedSize := numInserts * concurrencyLevel
	actualSize := len(bst.GetKeys())
	if actualSize != expectedSize {
		t.Errorf("Expected BST size to be %d, got %d", expectedSize, actualSize)
	}
}

func TestBSTConcurrentInsertsWithDuplicates(t *testing.T) {
	bst := NewBST() // Initialize your BST

	var wg sync.WaitGroup // Use WaitGroup to wait for all goroutines to finish
	const numInserts = 100  // Number of inserts per goroutine
	const concurrencyLevel = 10 // Number of goroutines
	const duplicateEvery = 5 // Insert a duplicate key every 'duplicateEvery' inserts

	// Perform concurrent inserts, including duplicates
	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numInserts; j++ {
				var key string
				if j%duplicateEvery == 0 {
					// Generate a duplicate key intentionally
					key = strconv.Itoa(j) // This will generate duplicates across goroutines
				} else {
					// Generate a unique key
					key = strconv.Itoa(goroutineID*numInserts + j)
				}
				bst.Insert(key, 1, 1) // Dummy tokens and capacity
			}
		}(i)
	}

	wg.Wait() // Wait for all goroutines to complete

	maxPossibleSize := numInserts*concurrencyLevel - (concurrencyLevel-1)*(numInserts/duplicateEvery)
	actualSize := len(bst.GetKeys())
	if actualSize > maxPossibleSize {
		t.Errorf("Expected BST size to be less than or equal to %d, got %d", maxPossibleSize, actualSize)
	}
}

func (node *Node) heightTestHelper(t *testing.T) int32 {
	if node == nil {
		return 0
	}

	left_height := node.left.heightTestHelper(t)

	right_height := node.right.heightTestHelper(t)

	acutal_height := 1 + max(left_height, right_height)

	if absInt32(acutal_height - node.getHeight()) > 2 {
		t.Errorf(
			"Expected height: %d and actual height: %d difference over the threshold! ",
			acutal_height,
			node.getHeight(),
		)
	}

	return acutal_height
}

func TestBSTHeightCalulcations(t *testing.T) {
	bst := NewBST()

	var wg sync.WaitGroup // Use WaitGroup to wait for all goroutines to finish
	const numInserts = 100  // Number of inserts per goroutine
	const concurrencyLevel = 100 // Number of goroutines
	const duplicateEvery = 5 // Insert a duplicate key every 'duplicateEvery' inserts

		// Perform concurrent inserts, including duplicates
		for i := 0; i < concurrencyLevel; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < numInserts; j++ {
					var key string
					if j%duplicateEvery == 0 {
						// Generate a duplicate key intentionally
						key = strconv.Itoa(j)
					} else {
						// Generate a unique key
						key = strconv.Itoa(goroutineID*numInserts + j)
					}
					bst.Insert(key, 1, 1) // Dummy tokens and capacity
				}
			}(i)
		}
	
		wg.Wait() // Wait for all goroutines to complete

		// Check the height of each node
		bst.root.heightTestHelper(t)
}