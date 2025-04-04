// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs

import "fmt"

type kargItem struct {
	karg Karg
	next *kargItem
	prev *kargItem
}

// remove deletes k from the list
func remove(k *kargItem) error {
	if k == nil {
		return fmt.Errorf("remove: %w", ErrNilPtr)
	}
	if k.prev != nil {
		k.prev.next = k.next
	}
	if k.next != nil {
		k.next.prev = k.prev
	}

	return nil
}

// replace replaces oldK with newK in the list
func replace(oldK, newK *kargItem) error {
	if oldK == nil {
		return fmt.Errorf("replace: old item: %w", ErrNilPtr)
	}
	if newK == nil {
		return fmt.Errorf("replace: new item: %w", ErrNilPtr)
	}
	newK.prev = oldK.prev
	newK.next = oldK.next
	if oldK.prev != nil {
		oldK.prev.next = newK
	}
	if oldK.next != nil {
		oldK.next.prev = newK
	}
	oldK.prev = nil
	oldK.next = nil

	return nil
}
