package hset

import (
	"container/heap"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapSetBasedHeapPopReturnLowestValues(t *testing.T) {
	h := Empty()
	heap.Init(&h)
	heap.Push(&h, "b")
	heap.Push(&h, "a")
	heap.Push(&h, "a")
	value := heap.Pop(&h)
	assert.Equal(t, "b", value)
}

func TestMapSetBasedHeapPopReturnLowestValuesWithMoreThan2Values(t *testing.T) {
	h := Empty()
	heap.Init(&h)
	heap.Push(&h, "a")
	heap.Push(&h, "a")
	heap.Push(&h, "b")
	heap.Push(&h, "c")
	heap.Push(&h, "d")
	value := heap.Pop(&h)
	assert.Equal(t, value, "b")
}

func TestMapSetBasedHeapMaxOccuringItem(t *testing.T) {
	h := Empty()
	heap.Init(&h)
	heap.Push(&h, "a")
	heap.Push(&h, "a")
	heap.Push(&h, "b")
	heap.Push(&h, "c")
	heap.Push(&h, "d")
	value := h.MaxOccuringItem()
	assert.Equal(t, value, "a")
}

func TestMapSetBasedHeapMaxOccuringItemIsStable(t *testing.T) {
	h := Empty()
	h.Add("a")
	h.Add("a")
	h.Add("b")
	h.Add("c")
	h.Add("c")
	heap.Init(&h)
	value := h.MaxOccuringItem()
	assert.Equal(t, value, "c")
}

// func BenchmarkMapSetBasedHeapMaxOccuringItem(b *testing.B) {
// 	var passedC = 0
// 	var passedA = 0
// 	for i := 0; i < b.N; i++ {
// 		h := Empty()
// 		h.Add("a")
// 		h.Add("a")
// 		h.Add("b")
// 		h.Add("c")
// 		h.Add("c")
// 		heap.Init(&h)
// 		value := h.MaxOccuringItem()
// 		if value == "c" {
// 			passedC++
// 		} else {
// 			passedA++
// 		}
// 	}
//
// 	fmt.Printf("A passed = %d\nC Passed = %d\n", passedA, passedC)
// }
