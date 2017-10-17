package dumper

import (
	"bufio"
	"io"
)

type FileWriter struct {
	wr *bufio.Writer
}

func (s *FileWriter) Write(data string) error {
	_, err := s.wr.WriteString(data)

	return err
}

func (s *FileWriter) Flush() error {
	return s.wr.Flush()
}

func NewFileWriter(w io.Writer) *FileWriter {
	return &FileWriter{wr: bufio.NewWriter(w)}
}
