package entity

// Node represents a single node in the linked list
type Node[T any] struct {
	Value T
	Next  *Node[T]
}

// LinkedList represents a generic linked list
type LinkedList[T any] struct {
	head *Node[T]
	tail *Node[T]
	size int
}

// NewLinkedList creates a new empty linked list
func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

// PushBack adds an element to the end of the list
func (ll *LinkedList[T]) PushBack(value T) {
	newNode := &Node[T]{Value: value}
	
	if ll.head == nil {
		ll.head = newNode
		ll.tail = newNode
	} else {
		ll.tail.Next = newNode
		ll.tail = newNode
	}
	
	ll.size++
}

// Pop removes and returns the first element from the list
// Returns the zero value of T and false if the list is empty
func (ll *LinkedList[T]) Pop() (T, bool) {
	var zero T
	
	if ll.head == nil {
		return zero, false
	}
	
	value := ll.head.Value
	ll.head = ll.head.Next
	
	if ll.head == nil {
		ll.tail = nil
	}
	
	ll.size--
	return value, true
}

// Len returns the number of elements in the list
func (ll *LinkedList[T]) Len() int {
	return ll.size
}