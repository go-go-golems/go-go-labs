package client

import (
	"context"
	"sync"
	"time"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/pkg/errors"
)

// DefaultMessageQueue implements a thread-safe message queue with priority support
type DefaultMessageQueue struct {
	mu            sync.RWMutex
	queue         []*pb.MeshPacket
	priorityQueue []*pb.MeshPacket
	capacity      int
	closed        bool

	// Flow control
	spaceAvailable chan struct{}
	itemAvailable  chan struct{}
}

// NewDefaultMessageQueue creates a new message queue with the given capacity
func NewDefaultMessageQueue(capacity int) *DefaultMessageQueue {
	return &DefaultMessageQueue{
		queue:          make([]*pb.MeshPacket, 0, capacity),
		priorityQueue:  make([]*pb.MeshPacket, 0, capacity/4), // 25% for priority
		capacity:       capacity,
		spaceAvailable: make(chan struct{}, 1),
		itemAvailable:  make(chan struct{}, 1),
	}
}

// Enqueue adds a packet to the regular queue
func (q *DefaultMessageQueue) Enqueue(packet *pb.MeshPacket) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return errors.New("queue is closed")
	}

	if len(q.queue)+len(q.priorityQueue) >= q.capacity {
		return errors.New("queue is full")
	}

	q.queue = append(q.queue, packet)
	q.notifyItemAvailable()

	return nil
}

// EnqueuePriority adds a packet to the priority queue
func (q *DefaultMessageQueue) EnqueuePriority(packet *pb.MeshPacket) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return errors.New("queue is closed")
	}

	if len(q.queue)+len(q.priorityQueue) >= q.capacity {
		return errors.New("queue is full")
	}

	q.priorityQueue = append(q.priorityQueue, packet)
	q.notifyItemAvailable()

	return nil
}

// Dequeue removes and returns the next packet (priority queue first)
func (q *DefaultMessageQueue) Dequeue() (*pb.MeshPacket, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil, errors.New("queue is closed")
	}

	// Check priority queue first
	if len(q.priorityQueue) > 0 {
		packet := q.priorityQueue[0]
		q.priorityQueue = q.priorityQueue[1:]
		q.notifySpaceAvailable()
		return packet, nil
	}

	// Check regular queue
	if len(q.queue) > 0 {
		packet := q.queue[0]
		q.queue = q.queue[1:]
		q.notifySpaceAvailable()
		return packet, nil
	}

	return nil, errors.New("queue is empty")
}

// Peek returns the next packet without removing it
func (q *DefaultMessageQueue) Peek() (*pb.MeshPacket, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return nil, errors.New("queue is closed")
	}

	// Check priority queue first
	if len(q.priorityQueue) > 0 {
		return q.priorityQueue[0], nil
	}

	// Check regular queue
	if len(q.queue) > 0 {
		return q.queue[0], nil
	}

	return nil, errors.New("queue is empty")
}

// Size returns the total number of items in the queue
func (q *DefaultMessageQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.queue) + len(q.priorityQueue)
}

// IsEmpty returns true if the queue is empty
func (q *DefaultMessageQueue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.queue) == 0 && len(q.priorityQueue) == 0
}

// Clear removes all items from the queue
func (q *DefaultMessageQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue = q.queue[:0]
	q.priorityQueue = q.priorityQueue[:0]

	// Notify that space is available
	q.notifySpaceAvailable()
}

// HasSpace returns true if there's space in the queue
func (q *DefaultMessageQueue) HasSpace() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.queue)+len(q.priorityQueue) < q.capacity
}

// WaitForSpace waits until there's space in the queue or context is cancelled
func (q *DefaultMessageQueue) WaitForSpace(ctx context.Context) error {
	if q.HasSpace() {
		return nil
	}

	select {
	case <-q.spaceAvailable:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WaitForItem waits until there's an item in the queue or context is cancelled
func (q *DefaultMessageQueue) WaitForItem(ctx context.Context) error {
	if !q.IsEmpty() {
		return nil
	}

	select {
	case <-q.itemAvailable:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close closes the queue
func (q *DefaultMessageQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil
	}

	q.closed = true
	close(q.spaceAvailable)
	close(q.itemAvailable)

	return nil
}

// GetCapacity returns the queue capacity
func (q *DefaultMessageQueue) GetCapacity() int {
	return q.capacity
}

// GetStatistics returns queue statistics
func (q *DefaultMessageQueue) GetStatistics() QueueStatistics {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return QueueStatistics{
		Size:         len(q.queue) + len(q.priorityQueue),
		Capacity:     q.capacity,
		RegularSize:  len(q.queue),
		PrioritySize: len(q.priorityQueue),
		IsClosed:     q.closed,
	}
}

// notifySpaceAvailable notifies that space is available (must be called with lock held)
func (q *DefaultMessageQueue) notifySpaceAvailable() {
	select {
	case q.spaceAvailable <- struct{}{}:
	default:
		// Channel full, notification already pending
	}
}

// notifyItemAvailable notifies that an item is available (must be called with lock held)
func (q *DefaultMessageQueue) notifyItemAvailable() {
	select {
	case q.itemAvailable <- struct{}{}:
	default:
		// Channel full, notification already pending
	}
}

// QueueStatistics holds statistics about the queue
type QueueStatistics struct {
	Size         int
	Capacity     int
	RegularSize  int
	PrioritySize int
	IsClosed     bool
}

// FlowControlledQueue implements a message queue with flow control
type FlowControlledQueue struct {
	*DefaultMessageQueue

	// Flow control settings
	maxWindowSize int
	currentWindow int
	ackRequired   bool
	pendingAcks   map[uint32]time.Time
	ackTimeout    time.Duration

	// Acknowledgment tracking
	ackMu sync.RWMutex
}

// NewFlowControlledQueue creates a new flow-controlled message queue
func NewFlowControlledQueue(capacity, maxWindowSize int, ackTimeout time.Duration) *FlowControlledQueue {
	return &FlowControlledQueue{
		DefaultMessageQueue: NewDefaultMessageQueue(capacity),
		maxWindowSize:       maxWindowSize,
		ackRequired:         true,
		pendingAcks:         make(map[uint32]time.Time),
		ackTimeout:          ackTimeout,
	}
}

// CanSend returns true if a message can be sent (flow control allows it)
func (q *FlowControlledQueue) CanSend() bool {
	q.ackMu.RLock()
	defer q.ackMu.RUnlock()

	if !q.ackRequired {
		return true
	}

	return q.currentWindow < q.maxWindowSize
}

// WaitForSendWindow waits until the send window has space
func (q *FlowControlledQueue) WaitForSendWindow(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if q.CanSend() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check for expired acknowledgments
			q.cleanupExpiredAcks()
		}
	}
}

// MessageSent records that a message was sent and needs acknowledgment
func (q *FlowControlledQueue) MessageSent(packetID uint32) {
	if !q.ackRequired {
		return
	}

	q.ackMu.Lock()
	defer q.ackMu.Unlock()

	q.pendingAcks[packetID] = time.Now()
	q.currentWindow++
}

// MessageAcked records that a message was acknowledged
func (q *FlowControlledQueue) MessageAcked(packetID uint32) {
	if !q.ackRequired {
		return
	}

	q.ackMu.Lock()
	defer q.ackMu.Unlock()

	if _, exists := q.pendingAcks[packetID]; exists {
		delete(q.pendingAcks, packetID)
		q.currentWindow--
	}
}

// SetAckRequired enables or disables acknowledgment requirements
func (q *FlowControlledQueue) SetAckRequired(required bool) {
	q.ackMu.Lock()
	defer q.ackMu.Unlock()

	q.ackRequired = required
	if !required {
		// Clear pending acknowledgments
		q.pendingAcks = make(map[uint32]time.Time)
		q.currentWindow = 0
	}
}

// GetFlowControlStatistics returns flow control statistics
func (q *FlowControlledQueue) GetFlowControlStatistics() FlowControlStatistics {
	q.ackMu.RLock()
	defer q.ackMu.RUnlock()

	return FlowControlStatistics{
		MaxWindowSize: q.maxWindowSize,
		CurrentWindow: q.currentWindow,
		PendingAcks:   len(q.pendingAcks),
		AckRequired:   q.ackRequired,
		AckTimeout:    q.ackTimeout,
	}
}

// cleanupExpiredAcks removes expired acknowledgments
func (q *FlowControlledQueue) cleanupExpiredAcks() {
	q.ackMu.Lock()
	defer q.ackMu.Unlock()

	now := time.Now()
	for packetID, timestamp := range q.pendingAcks {
		if now.Sub(timestamp) > q.ackTimeout {
			delete(q.pendingAcks, packetID)
			q.currentWindow--
		}
	}
}

// FlowControlStatistics holds flow control statistics
type FlowControlStatistics struct {
	MaxWindowSize int
	CurrentWindow int
	PendingAcks   int
	AckRequired   bool
	AckTimeout    time.Duration
}
