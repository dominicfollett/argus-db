package naive

import (
	"math/rand"
	"sort"
	"testing"
)

func TestAVL(t *testing.T) {

	keys := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	avl := NewAVL()

	// Shuffle the keys randomly
	rand.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	for _, k := range(keys) {
		avl.Insert([]byte(k), nil)
	}

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