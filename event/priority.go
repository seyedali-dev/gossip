// Package event. priority provides priority-based event processing.
package event

import (
	"container/heap"
	"sync"
)

// Priority levels for events.
const (
	PriorityLow    = 1
	PriorityNormal = 5
	PriorityHigh   = 10
)

// PriorityEvent wraps an event with priority information.
type PriorityEvent struct {
	Event    *Event
	Priority int
	index    int
}

// priorityQueue implements heap.Interface for priority-based event processing.
type priorityQueue []*PriorityEvent

// -------------------------------------------- Public Functions --------------------------------------------

// Len returns the number of events in the queue.
func (pq priorityQueue) Len() int {
	return len(pq)
}

// Less compares two events by priority (higher priority first).
func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority
}

// Swap swaps two events in the queue.
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push adds an event to the queue.
func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PriorityEvent)
	item.index = n
	*pq = append(*pq, item)
}

// Pop removes and returns the highest priority event.
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// -------------------------------------------- Priority Queue Manager --------------------------------------------

// PriorityQueue manages priority-based event processing.
type PriorityQueue struct {
	mu    sync.Mutex
	queue priorityQueue
}

// NewPriorityQueue creates a new priority queue.
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		queue: make(priorityQueue, 0),
	}
	heap.Init(&pq.queue)
	return pq
}

// Enqueue adds an event with the specified priority.
func (pq *PriorityQueue) Enqueue(event *Event, priority int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pe := &PriorityEvent{
		Event:    event,
		Priority: priority,
	}

	heap.Push(&pq.queue, pe)
}

// Dequeue removes and returns the highest priority event.
func (pq *PriorityQueue) Dequeue() (*Event, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.queue.Len() == 0 {
		return nil, false
	}

	pe := heap.Pop(&pq.queue).(*PriorityEvent)
	return pe.Event, true
}

// Size returns the number of events in the queue.
func (pq *PriorityQueue) Size() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return pq.queue.Len()
}
