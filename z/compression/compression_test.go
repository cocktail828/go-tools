// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package compression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testProvider struct {
	name     string
	provider Provider

	// Compressed data for "hello"
	compressedHello []byte
}

var providers = []testProvider{
	{"zlib", NewZLibProvider(), []byte{0x78, 0x9c, 0xca, 0x48, 0xcd, 0xc9, 0xc9, 0x07, 0x00, 0x00, 0x00, 0xff, 0xff}},
	{"lz4", NewLz4Provider(), []byte{0x50, 0x68, 0x65, 0x6c, 0x6c, 0x6f}},
	{"zstd", NewZStdProvider(Default),
		[]byte{0x28, 0xb5, 0x2f, 0xfd, 0x20, 0x05, 0x29, 0x00, 0x00, 0x68, 0x65, 0x6c, 0x6c, 0x6f}},
}

func TestCompression(t *testing.T) {
	for _, provider := range providers {
		p := provider
		t.Run(p.name, func(t *testing.T) {
			hello := []byte("test compression data")
			compressed := make([]byte, 1024)
			compressed, _ = p.provider.Compress(compressed, hello)

			uncompressed := make([]byte, 1024)
			uncompressed, err := p.provider.Decompress(uncompressed, compressed, len(hello))
			assert.Nil(t, err)
			assert.ElementsMatch(t, hello, uncompressed)
		})
	}
}

func TestCompressionNoBuffers(t *testing.T) {
	for _, provider := range providers {
		p := provider
		t.Run(p.name, func(t *testing.T) {
			hello := []byte("test compression data")
			compressed, _ := p.provider.Compress(nil, hello)
			uncompressed, err := p.provider.Decompress(nil, compressed, len(hello))
			assert.Nil(t, err)
			assert.ElementsMatch(t, hello, uncompressed)
		})
	}
}

func TestJavaCompatibility(t *testing.T) {
	for _, provider := range providers {
		p := provider
		t.Run(p.name, func(t *testing.T) {
			hello := []byte("hello")
			uncompressed, err := p.provider.Decompress(nil, p.compressedHello, len(hello))
			assert.Nil(t, err)
			assert.ElementsMatch(t, hello, uncompressed)
		})
	}
}

func TestDecompressionError(t *testing.T) {
	for _, provider := range providers {
		p := provider
		t.Run(p.name, func(t *testing.T) {
			_, err := p.provider.Decompress(nil, []byte{0x05}, 10)
			assert.NotNil(t, err)
		})
	}
}
