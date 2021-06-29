package fluentd

import (
	"strings"
)

const (
	BufferTypeFile   = "file"
	BufferTypeMemory = "memory"
)

type Buffer struct {
	outPrefix string
	Configuration
	Type string
}

func NewBuffer(bufferType string) *Buffer {
	return &Buffer{
		Configuration: Configuration{
			Type: bufferType,
			AllowedKeys: NewSet(
				"path",
				"chunk_limit_size",
				"chunk_limit_records",
				"total_limit_size",
				"chunk_full_threshold",
				"queued_chunks_limit_size",
				"compress",
				"timekey",
				"timekey_use_utc",
				"timekey_wait",
			),
			Config: map[string]interface{}{},
		},
		outPrefix: "\t",
	}
}

func (b *Buffer) Set(key string, value interface{}) {
	b.Config[key] = value
}

func (b *Buffer) AsList() []string {
	buf := []string{"<buffer>"}
	buf = append(buf, BuildBlock(b.Configuration)...)
	return append(buf, "</buffer>")
}

func (b *Buffer) String() string {
	return strings.Join(b.AsList(), "\n")
}

func (b *Buffer) WithPath(value string) *Buffer {
	b.Config["path"] = value
	return b
}
func (b *Buffer) WithTimeKey(value string) *Buffer {
	b.Config["timekey"] = value
	return b
}
func (b *Buffer) WithTimeKeyUseUTC(value bool) *Buffer {
	b.Config["timekey_use_utc"] = value
	return b
}
func (b *Buffer) WithTimeKeyWait(value string) *Buffer {
	b.Config["timekey_wait"] = value
	return b
}
