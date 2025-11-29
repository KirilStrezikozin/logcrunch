// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package internal

import "errors"

var ErrNilConnection = errors.New("nil connection")
var ErrConnectionAlreadyEstablished = errors.New("connection already established")
