// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs

import "errors"

var (
	ErrInvalidKey = errors.New("key contains invalid characters")
	ErrNilPtr     = errors.New("pointer is nil")
	ErrNotExists  = errors.New("karg does not exist")
)
