//    HipparchiaGoDBHelper: search and vector helper app and functions for HipparchiaServer
//    Copyright: E Gunderson 2021
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package main

import "sync"

type DbWorkline struct {
	WkUID       string
	TbIndex     int64
	Lvl5Value   string
	Lvl4Value   string
	Lvl3Value   string
	Lvl2Value   string
	Lvl1Value   string
	Lvl0Value   string
	MarkedUp    string
	Accented    string
	Stripped    string
	Hypenated   string
	Annotations string
}

func (dbw DbWorkline) FindLocus() []string {
	loc := [6]string{
		dbw.Lvl5Value,
		dbw.Lvl4Value,
		dbw.Lvl3Value,
		dbw.Lvl2Value,
		dbw.Lvl1Value,
		dbw.Lvl0Value,
	}

	var trim []string
	for _, l := range loc {
		if l != "-1" {
			trim = append(trim, l)
		}
	}
	return trim
}

func (dbw DbWorkline) FindAuthor() string {
	return dbw.WkUID[:6]
}

// WorklineStack - stack of []DbWorklines.
type WorklineStack struct {
	// Slice of type ItemType, it holds items in stack.
	items []DbWorkline
	// rwLock for handling concurrent operations on the stack.
	rwLock sync.RWMutex
}

// Append - Adds Items to the top of the stack
func (stack *WorklineStack) Append(t []DbWorkline) {
	//Initialize items slice if not initialized
	if stack.items == nil {
		stack.items = []DbWorkline{}
	}
	// Acquire read, write lock before inserting a new item in the stack.
	stack.rwLock.Lock()
	// Performs append operation.
	stack.items = append(stack.items, t...)
	// This will release read, write lock
	stack.rwLock.Unlock()
}

// Push - Adds an Item to the top of the stack
func (stack *WorklineStack) Push(t DbWorkline) {
	//Initialize items slice if not initialized
	if stack.items == nil {
		stack.items = []DbWorkline{}
	}
	// Acquire read, write lock before inserting a new item in the stack.
	stack.rwLock.Lock()
	// Performs append operation.
	stack.items = append(stack.items, t)
	// This will release read, write lock
	stack.rwLock.Unlock()
}

// Pop removes an Item from the top of the stack
func (stack *WorklineStack) Pop() *DbWorkline {
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
func (stack *WorklineStack) Size() int {
	// Acquire read lock
	stack.rwLock.RLock()
	// defer operation of unlock.
	defer stack.rwLock.RUnlock()
	// Return length of items slice.
	return len(stack.items)
}

// All - return all items present in stack
func (stack WorklineStack) All() []DbWorkline {
	// Acquire read lock
	stack.rwLock.RLock()
	// defer operation of unlock.
	defer stack.rwLock.RUnlock()
	// Return items slice to caller.
	return stack.items
}

// IsEmpty - Check is stack is empty or not.
func (stack *WorklineStack) IsEmpty() bool {
	// Acquire read lock
	stack.rwLock.RLock()
	// defer operation of unlock.
	defer stack.rwLock.RUnlock()
	return len(stack.items) == 0
}

type DbWordCount struct {
	Word  string
	Total int64
	Gr    int64
	Lt    int64
	Dp    int64
	In    int64
	Ch    int64
}

type DbLexicon struct {
	// skipping 'unaccented_entry' from greek_dictionary
	// skipping 'entry_key' from latin_dictionary
	Word     string
	Metrical string
	ID       int64
	POS      string
	Transl   string
	Entry    string
}

// https://golangbyexample.com/sort-custom-struct-collection-golang/
type WeightedHeadword struct {
	Word  string
	Count int
}

type WHWList []WeightedHeadword

func (w WHWList) Len() int {
	return len(w)
}

func (w WHWList) Less(i, j int) bool {
	return w[i].Count > w[j].Count
}

func (w WHWList) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

type BagWithLocus struct {
	Loc string
	Bag string
}

type DbMorphology struct {
	Observed    string
	Xrefs       string
	PrefixXrefs string
	RawPossib   string
	RelatedHW   string
}

type MorphPossib struct {
	Transl   string `json:"transl"`
	Anal     string `json:"analysis"`
	Headwd   string `json:"headword"`
	Scansion string `json:"scansion"`
	Xrefkind string `json:"xref_kind"`
	Xrefval  string `json:"xref_value"`
}

type CompositePollingData struct {
	// this has to be kept in sync with rediskeys[8] and HipparchiaServer's interface
	Launchtime    float64
	Active        string // redis polls store 'yes' or 'no'; but the value is converted to T/F by .getactivity()
	Statusmessage string
	Remaining     int64
	Poolofwork    int64
	Hitcount      int64
	Portnumber    int64
	Notes         string
	ID            string // this is not stored in redis; it is asserted here
}

type BrowsedPassage struct {
	// marshal will not do lc names
	Browseforwards    string `json:"browseforwards"`
	Browseback        string `json:"browseback"`
	Authornumber      string `json:"authornumber"`
	Workid            string `json:"workid"`
	Worknumber        string `json:"worknumber"`
	Authorboxcontents string `json:"authorboxcontents"`
	Workboxcontents   string `json:"workboxcontents"`
	Browserhtml       string `json:"browserhtml"`
}
