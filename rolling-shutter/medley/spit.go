package medley

import (
	"os"
)

// SecureSpit creates a new file with the given path and writes the given content to it. The file
// is created with with mode 0600. SecureSpit will not overwrite an existing file.
func SecureSpit(path string, content []byte) error {
	var err error
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	if err != nil {
		return err
	}
	return file.Sync()
}
