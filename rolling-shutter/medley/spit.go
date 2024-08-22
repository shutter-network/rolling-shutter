package medley

import (
	"os"

	"github.com/spf13/afero"
)

// SecureSpit creates a new file with the given path and writes the given content to it. The file
// is created with with mode 0600. SecureSpit will not overwrite an existing file unless asked.
func SecureSpit(fs afero.Fs, path string, content []byte, overwrite bool) error {
	var err error
	flags := os.O_RDWR | os.O_CREATE
	if !overwrite {
		flags |= os.O_EXCL
	}
	file, err := fs.OpenFile(path, flags, 0o600)
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
