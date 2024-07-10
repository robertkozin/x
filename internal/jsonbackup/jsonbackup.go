package jsonbackup

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Must backs up stuff
//
//	defer jsonbackup.Must(&data, "./data.json")()
func Must[T any](data *T, path string) func() {
	// open
	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(path)
		if err != nil {
			panic(fmt.Errorf("jsonbackup.open: creating file %q: %w", path, err))
		}

		enc := json.NewEncoder(f)
		enc.SetEscapeHTML(false)
		err = enc.Encode(data)
	} else if err != nil {
		panic(fmt.Errorf("jsonbackup.open: opening file %q: %w", path, err))
	} else {
		dec := json.NewDecoder(f)
		err = dec.Decode(data)
		if err != nil {
			panic(fmt.Errorf("jsonbackup.open: decoding file %q: %w", path, err))
		}
	}

	// TODO: Periodic backup

	// save
	return func() {
		var err1 error
		if err1 = f.Truncate(0); err1 != nil {
			fmt.Printf("jsonback.save: truncate %q: %v\n", f.Name(), err1)
		}
		if _, err1 = f.Seek(0, 0); err1 != nil {
			fmt.Printf("jsonback.save: seek %q: %v\n", f.Name(), err1)
		}

		enc := json.NewEncoder(f)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		if err1 = enc.Encode(data); err1 != nil {
			fmt.Printf("jsonback.save: encode %q: %v\n", f.Name(), err1)
		}

		if err1 = f.Sync(); err1 != nil {
			fmt.Printf("jsonback.save: sync %q: %v\n", f.Name(), err1)
		}

		if err1 = f.Close(); err1 != nil {
			fmt.Printf("jsonback.save: close %q: %v\n", f.Name(), err1)
		}
	}
}
