// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package internal

import (
	"fmt"

	"github.com/KirilStrezikozin/logcrunch/internal/types"
	"github.com/boltdb/bolt"
)

type DBError struct {
	Op  string
	Err error
}

func (e *DBError) Error() string {
	return fmt.Sprintf("db %s: %v", e.Op, e.Err)
}

type DBReader interface {
	Get(bucketName, key []byte, fn func([]byte) error) error
}

type DBWriter interface {
	Put(bucketName, key, value []byte) error
}

type DBReadWriter interface {
	DBReader
	DBWriter
}

type DB interface {
	DBReadWriter
	Open() error
	Close() error
}

type BoltDB struct {
	db *bolt.DB
}

func NewBoltDB() *BoltDB {
	return &BoltDB{}
}

func (db *BoltDB) Open() error {
	var err error
	db.db, err = bolt.Open(types.DBFilePath, types.DBFileMode, nil)
	if err != nil {
		return &DBError{Op: "open", Err: err}
	}
	return nil
}

func (db *BoltDB) Close() error {
	err := db.db.Close()
	if err != nil {
		return &DBError{Op: "close", Err: err}
	}
	return nil
}

func (db *BoltDB) Get(bucketName, key []byte, fn func([]byte) error) error {
	err := db.db.View(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return &DBError{Op: "create or get bucket", Err: err}
		}

		if b == nil {
			return nil
		}
		value := b.Get(key)
		return fn(value)
	})

	if err != nil {
		return &DBError{Op: "get", Err: err}
	}
	return nil
}

func (db *BoltDB) Put(bucketName, key, value []byte) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return &DBError{Op: "create or get bucket", Err: err}
		}

		if err := b.Put(key, value); err != nil {
			return &DBError{Op: "put", Err: err}
		}
		return nil
	})

	if err != nil {
		return &DBError{Op: "put", Err: err}
	}
	return nil
}
