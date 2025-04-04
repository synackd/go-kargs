// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs

import "errors"

var (
	ErrNilPtr     = errors.New("pointer is nil")
	ErrInvalidKey = errors.New("key contains invalid characters")
	ErrNotExists  = errors.New("karg does not exist")
)
