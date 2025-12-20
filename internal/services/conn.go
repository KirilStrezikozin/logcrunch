// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package services

import (
	"github.com/KirilStrezikozin/logcrunch/internal/types"
	"github.com/rs/zerolog"
)

type IConnection interface {
	GetURL() (string, error)
	SetURL(url string) error
	GetStatus() (types.ConnectionStatus, error)
}

type Connection struct {
	logger zerolog.Logger
}
