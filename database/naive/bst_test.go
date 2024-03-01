package naive

import (
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"testing"
)

// TODO: Generated Tests -- revisit and refactor
func TestBSTConcurrentInserts(t *testing.T) {
	bst := NewBST()

	var wg sync.WaitGroup
	const numInserts = 100      // Number of inserts per goroutine
	const concurrencyLevel = 10 // Number of goroutines

	// Perform concurrent inserts
	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numInserts; j++ {
				key := strconv.Itoa(goroutineID*numInserts + j)
				bst.Insert(key, 1, 1)
			}
		}(i)
	}

	wg.Wait()

	// After all inserts, check the size of the BST to ensure all inserts were successful
	expectedSize := numInserts * concurrencyLevel
	actualSize := len(bst.GetKeys())
	if actualSize != expectedSize {
		t.Errorf("Expected BST size to be %d, got %d", expectedSize, actualSize)
	}
}

func TestBSTConcurrentInsertsWithDuplicates(t *testing.T) {
	bst := NewBST()

	var wg sync.WaitGroup
	const numInserts = 100
	const concurrencyLevel = 10
	const duplicateEvery = 5

	// Perform concurrent inserts, including duplicates
	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numInserts; j++ {
				var key string
				if j%duplicateEvery == 0 {
					// Generate a duplicate key intentionally -- I'm not crazy about this
					key = strconv.Itoa(j)
				} else {
					// Generate a unique key
					key = strconv.Itoa(goroutineID*numInserts + j)
				}
				bst.Insert(key, 1, 1)
			}
		}(i)
	}

	wg.Wait()

	maxPossibleSize := numInserts*concurrencyLevel - ((concurrencyLevel - 1) * (numInserts / duplicateEvery))
	actualSize := len(bst.GetKeys())
	if actualSize != maxPossibleSize {
		t.Errorf("Expected BST size to be equal to %d, got %d", maxPossibleSize, actualSize)
	}
}

func (node *Node) heightTestHelper(t *testing.T) int32 {
	if node == nil {
		return 0
	}

	left_height := node.left.heightTestHelper(t)
	right_height := node.right.heightTestHelper(t)

	expected_height := 1 + max(left_height, right_height)

	if absInt32(expected_height-node.getHeight()) != 0 {
		t.Errorf(
			"Difference between expected height: %d, and actual height: %d, is greater than zero! "+
				"Left height: %d, Right height: %d",
			expected_height,
			node.getHeight(),
			left_height,
			right_height,
		)
	}

	return expected_height
}

func (node *Node) heightTestHelperCount(t *testing.T, count int) (int32, int) {
	if node == nil {
		return 0, count
	}

	left_height, c := node.left.heightTestHelperCount(t, count)
	count = c
	right_height, c := node.right.heightTestHelperCount(t, count)
	count = c

	expected_height := 1 + max(left_height, right_height)

	if absInt32(expected_height-node.getHeight()) != 0 {
		count += 1
	}

	return expected_height, count
}

func TestBSTHeightCalulcations(t *testing.T) {
	bst := NewBST()

	var wg sync.WaitGroup
	const numInserts = 100
	const concurrencyLevel = 10
	const duplicateEvery = 5

	// TODO: I think a problem with this test is that the unique keys are always increasing??
	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numInserts; j++ {
				var key string
				if j%duplicateEvery == 0 {
					key = strconv.Itoa(j)
				} else {
					key = strconv.Itoa(goroutineID*numInserts + j)
				}
				bst.Insert(key, 1, 1)
			}
		}(i)
	}
	wg.Wait()

	maxPossibleSize := numInserts*concurrencyLevel - ((concurrencyLevel - 1) * (numInserts / duplicateEvery))
	actualSize := len(bst.GetKeys())
	if actualSize != maxPossibleSize {
		t.Errorf("Expected BST size to be equal to %d, got %d", maxPossibleSize, actualSize)
	}

	// Check the height of each node
	bst.root.heightTestHelper(t)

	// However I'd like to understand why it appears that all differences between expected and actual is 1 and not any other number.
	// Can we prove this? At the very least we'd then know that the heights are be being consistently under estimated and not over estimated.
	_, count := bst.root.heightTestHelperCount(t, 0)
	if count != 0 {
		t.Errorf("Total differences %d", count)
	}
}

func TestWithRandomKeys(t *testing.T) {

	keys := []string{"T", "X", "G", "L", "E", "Q", "M", "H", "O", "I", "B", "Z", "A", "V", "S", "R", "K", "P", "C", "D", "U", "F", "N", "W", "Y", "J"}
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	bst := NewBST()

	var wg sync.WaitGroup
	concurrencyLevel := 6

	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < concurrencyLevel; j++ {
				// Use modulo to allow overflow and a bit of duplication
				offset := (goroutineID*concurrencyLevel + j) % len(keys)
				bst.Insert(keys[offset], 1, 1)
			}
		}(i)
	}
	wg.Wait()

	result := bst.GetKeys()
	sort.Strings(keys)

	if len(keys) != len(result) {
		t.Errorf("Expected and result slices differ in length; expected: %d, got: %d", len(keys), len(result))
	}

	for i, expectedKey := range keys {
		if expectedKey != string(result[i]) {
			t.Errorf("Key mismatch at index %d; expected: %s, got: %s", i, expectedKey, result[i])
		}
	}

	// Check the height of each node
	bst.root.heightTestHelper(t)

	// Count the number of differences between expected and actual heights
	_, count := bst.root.heightTestHelperCount(t, 0)
	if count != 0 {
		t.Errorf("Total differences %d", count)
	}
}
