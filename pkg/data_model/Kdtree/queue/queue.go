package queue

type Queue[T any] interface {
	Pop() (T, bool)
	Push(elem T)
	Empty() bool
	Size() int
}

type LLQueue[T any] struct {
	head *LLQueueNode[T]
	tail *LLQueueNode[T]
	size int
}

type LLQueueNode[T any] struct {
	value T
	next  *LLQueueNode[T]
}

func NewLLQueue[T any]() *LLQueue[T] {
	return &LLQueue[T]{}
}

func (llq *LLQueue[T]) Pop() (T, bool) {
	if llq.head == nil {
		return *new(T), false
	}
	llq.size--
	if llq.head == llq.tail {
		value := llq.head.value
		llq.head = nil
		llq.tail = nil
		return value, true
	}
	temp := llq.head
	llq.head = llq.head.next
	value := temp.value
	return value, true
}

func (llq *LLQueue[T]) Push(elem T) {
	llq.size++
	n := &LLQueueNode[T]{
		value: elem,
	}
	if llq.head == nil {
		llq.head = n
		llq.tail = llq.head
		return
	}
	llq.tail.next = n
	llq.tail = n
}

func (llq *LLQueue[T]) Size() int {
	return llq.size
}

func (llq *LLQueue[T]) Empty() bool {
	return llq.size == 0
}
