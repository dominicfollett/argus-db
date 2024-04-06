//nolint:testpackage // Allow tests to access the naive package
package naive

import (
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"
)

// TestBSTHeightCalulcations tests the height of the BST after concurrent inserts.
func TestInsertBST(t *testing.T) {
	bst := NewBST()

	var wg sync.WaitGroup
	const numInserts = 100
	const concurrencyLevel = 10
	const duplicateEvery = 5

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
				bst.Insert(key)
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
	_, count := bst.root.heightTestHelper(t, 0)
	if count != 0 {
		t.Errorf("Total differences %d", count)
	}
}

func TestInsertBSTWithRandomKeys(t *testing.T) {
	keys := []string{
		"T", "X", "G", "L", "E", "Q", "M", "H", "O", "I", "B", "Z", "A", "V", "S",
		"R", "K", "P", "C", "D", "U", "F", "N", "W", "Y", "J",
	}
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
				bst.Insert(keys[offset])
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
		if expectedKey != result[i] {
			t.Errorf("Key mismatch at index %d; expected: %s, got: %s", i, expectedKey, result[i])
		}
	}

	// Check the height of each node
	_, count := bst.root.heightTestHelper(t, 0)
	if count != 0 {
		t.Errorf("Total differences %d", count)
	}
}

func (node *Node) heightTestHelper(t *testing.T, count int) (int32, int) {
	if node == nil {
		if node.getHeight() != -1 {
			t.Errorf("Expected height of nil node to be -1, got %d", node.getHeight())
		}
		return -1, count
	}

	leftHeight, count := node.left.heightTestHelper(t, count)
	rightHeight, count := node.right.heightTestHelper(t, count)

	expectedHeight := 1 + max(leftHeight, rightHeight)

	if absInt32(expectedHeight-node.getHeight()) != 0 {
		t.Errorf(
			"%s: Difference between expected height: %d, and actual height: %d, is greater than zero! "+
				"Left height: %d, Right height: %d",
			node.key,
			expectedHeight,
			node.getHeight(),
			leftHeight,
			rightHeight,
		)
		count++
	}

	return expectedHeight, count
}

func TestInSearchBST(t *testing.T) {
	bst := NewBST()

	var wg sync.WaitGroup
	const numInserts = 100
	const concurrencyLevel = 10
	const duplicateEvery = 5

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
				node := bst.InSearch(key)

				time.Sleep(time.Duration(rand.Intn(100)) * time.Nanosecond)

				node.lock.Unlock()
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
	_, count := bst.root.heightTestHelper(t, 0)
	if count != 0 {
		t.Errorf("Total differences %d", count)
	}
}
