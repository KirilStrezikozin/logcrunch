// Copyright 2025 The Logcrunch Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package types

const (
	DBFilePath = "logcrunch.db"
	DBFileMode = 0600
)

var (
	connectionBucketName = []byte("connection")
	connectionURLKey     = []byte("url")
)

func GetConnectionBucketName() []byte {
	return connectionBucketName
}

func GetConnectionURLKey() []byte {
	return connectionURLKey
}
