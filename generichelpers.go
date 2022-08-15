package main

// https://rksurwase.medium.com/slice-based-stack-implementation-in-golang-8140603a1dc2
// will just stack everything as json...

import (
	"sync"
)

// Stack - Stack of items.
type Stack struct {
	// Slice of type ItemType, it holds items in stack.
	items [][]byte
	// rwLock for handling concurrent operations on the stack.
	rwLock sync.RWMutex
}

// Push - Adds an Item to the top of the stack
func (stack *Stack) Push(t []byte) {
	//Initialize items slice if not initialized
	if stack.items == nil {
		stack.items = [][]byte{}
	}
	// Acquire read, write lock before inserting a new item in the stack.
	stack.rwLock.Lock()
	// Performs append operation.
	stack.items = append(stack.items, t)
	// This will release read, write lock
	stack.rwLock.Unlock()
}

// Pop removes an Item from the top of the stack
func (stack *Stack) Pop() *[]byte {
	// Checking if stack is empty before performing pop operation
	if len(stack.items) == 0 {
		return nil
	}
	// Acquire read, write lock as items are going to modify.
	stack.rwLock.Lock()
	// Popping item from items slice.
	item := stack.items[len(stack.items)-1]
	//Adjusting the item's length accordingly
	stack.items = stack.items[0 : len(stack.items)-1]
	// Release read write lock.
	stack.rwLock.Unlock()
	// Return last popped item
	return &item
}

// Size return size i.e. number of items present in stack.
func (stack *Stack) Size() int {
	// Acquire read lock
	stack.rwLock.RLock()
	// defer operation of unlock.
	defer stack.rwLock.RUnlock()
	// Return length of items slice.
	return len(stack.items)
}

// All - return all items present in stack
func (stack *Stack) All() [][]byte {
	// Acquire read lock
	stack.rwLock.RLock()
	// defer operation of unlock.
	defer stack.rwLock.RUnlock()
	// Return items slice to caller.
	return stack.items
}

// IsEmpty - Check is stack is empty or not.
func (stack *Stack) IsEmpty() bool {
	// Acquire read lock
	stack.rwLock.RLock()
	// defer operation of unlock.
	defer stack.rwLock.RUnlock()
	return len(stack.items) == 0
}

//
// misc generic functions
//

// unique - return only the unique items from a slice
func unique[T comparable](s []T) []T {
	// https://gosamples.dev/generics-remove-duplicates-slice/
	inResult := make(map[T]bool)
	var result []T
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			result = append(result, str)
		}
	}
	return result
}
