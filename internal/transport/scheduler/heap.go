package scheduler

import "github.com/sandevgo/tuskbot/internal/core"

// taskHeap implements heap.Interface for Task items.
// We want the item with the earliest NextRun to be popped first (Min-Heap).
type taskHeap []*core.Task

func (h taskHeap) Len() int {
	return len(h)
}

func (h taskHeap) Less(i, j int) bool {
	// Min-Heap based on NextRun time
	return h[i].NextRun.Before(h[j].NextRun)
}

func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *taskHeap) Push(x interface{}) {
	*h = append(*h, x.(*core.Task))
}

func (h *taskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func (h *taskHeap) Peek() *core.Task {
	if len(*h) == 0 {
		return nil
	}
	return (*h)[0]
}
