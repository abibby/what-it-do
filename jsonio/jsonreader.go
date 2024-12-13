package jsonio

import (
	"encoding/json"
	"io"
)

type JSONReader struct {
	reader *io.PipeReader
}

var _ io.ReadCloser = (*JSONReader)(nil)

func NewReader(v any) *JSONReader {
	r, w := io.Pipe()
	go func() {
		err := json.NewEncoder(w).Encode(v)
		w.CloseWithError(err)
	}()
	return &JSONReader{
		reader: r,
	}
}

// Read implements io.Reader.
func (j *JSONReader) Read(p []byte) (n int, err error) {
	return j.reader.Read(p)
}

// Close implements io.ReadCloser.
func (j *JSONReader) Close() error {
	return j.reader.Close()
}
