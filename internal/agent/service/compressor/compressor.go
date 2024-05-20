package compressor

import (
	"bytes"
	"compress/gzip"
)

type compressor struct {
}

func NewCompressorService() *compressor {
	return &compressor{}
}

func (c *compressor) Compress(b []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(buf)

	_, err := gz.Write(b)
	if err != nil {
		return nil, err
	}
	gz.Close()

	return buf.Bytes(), nil
}
