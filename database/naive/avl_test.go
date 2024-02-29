//go:build !race
// +build !race

package naive

import (
	"sort"
	"testing"
)

func TestAVL(t *testing.T) {

	keys := []string{"T", "X", "G", "L", "E", "Q", "M", "H", "O", "I", "B", "Z", "A", "V", "S", "R", "K", "P", "C", "D", "U", "F", "N", "W", "Y", "J"}
	avl := NewAVL()

	for _, k := range keys {
		avl.Insert(k, nil)
	}

	// Add some duplicates
	avl.Insert("T", nil)
	avl.Insert("D", nil)
	avl.Insert("N", nil)
	avl.Insert("P", nil)

	result := avl.GetKeys()
	sort.Strings(keys)

	if len(keys) != len(result) {
		t.Errorf("Expected and result slices differ in length; expected: %d, got: %d", len(keys), len(result))
	}

	for i, expectedKey := range keys {
		if expectedKey != string(result[i]) {
			t.Errorf("Key mismatch at index %d; expected: %s, got: %s", i, expectedKey, result[i])
		}
	}
}
