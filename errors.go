// SPDX-FileCopyrightText: © 2025 synack.d
//
// SPDX-License-Identifier: BSD-3-Clause

package kargs

import "errors"

var (
	ErrInvalidKey = errors.New("key contains invalid characters")
	ErrNilPtr     = errors.New("pointer is nil")
	ErrNotExists  = errors.New("karg does not exist")
)
