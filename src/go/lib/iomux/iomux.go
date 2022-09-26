package iomux

import (
	"io"
	"os"
)

// StdIOCapture multiplexes stdout and stderr to all provided writers; panics of io.Copy fails for any writer
func StdIOCapture(writers ...io.Writer) (io.Writer, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	oldStdErr := os.Stderr

	writers = append(writers, oldStdErr)

	mWriter := io.MultiWriter(writers...)

	os.Stderr = w
	os.Stdout = w

	go func() {
		if _, err = io.Copy(mWriter, r); err != nil {
			panic(err)
		}
	}()

	return w, nil
}
