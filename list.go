// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs

import "fmt"

type kargItem struct {
	karg Karg
	next *kargItem
	prev *kargItem
}

// unlink detaches k from the linked list owned by the passed Kargs, fixing the
// prev/next pointers of its neighbors as well as the list head (k.list) and
// tail (k.last) pointers when k is at either end. It does not touch the keyMap
// or numParams; callers are responsible for keeping those in sync.
func (k *Kargs) unlink(item *kargItem) error {
	if item == nil {
		return fmt.Errorf("unlink: %w", ErrNilPtr)
	}
	if item.prev != nil {
		item.prev.next = item.next
	} else {
		// item was the head
		k.list = item.next
	}
	if item.next != nil {
		item.next.prev = item.prev
	} else {
		// item was the tail
		k.last = item.prev
	}
	item.prev = nil
	item.next = nil

	return nil
}

// substitute replaces oldItem with newItem in the linked list owned by the
// passed Kargs, preserving list order and fixing the head (k.list) and tail
// (k.last) pointers when oldItem is at either end. It does not touch the keyMap
// or numParams; callers are responsible for keeping those in sync.
func (k *Kargs) substitute(oldItem, newItem *kargItem) error {
	if oldItem == nil {
		return fmt.Errorf("substitute: old item: %w", ErrNilPtr)
	}
	if newItem == nil {
		return fmt.Errorf("substitute: new item: %w", ErrNilPtr)
	}
	newItem.prev = oldItem.prev
	newItem.next = oldItem.next
	if oldItem.prev != nil {
		oldItem.prev.next = newItem
	} else {
		// oldItem was the head
		k.list = newItem
	}
	if oldItem.next != nil {
		oldItem.next.prev = newItem
	} else {
		// oldItem was the tail
		k.last = newItem
	}
	oldItem.prev = nil
	oldItem.next = nil

	return nil
}
