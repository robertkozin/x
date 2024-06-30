package htmgo

import (
	"errors"
	"fmt"
	"golang.org/x/tools/imports"
	"os"
	"path/filepath"
	"strings"
)

func Generate(file string, pkg string) error {
	dir := filepath.Dir(file)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var errs []error

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			f := filepath.Join(dir, entry.Name())
			errs = append(errs, Parse(f, pkg))
		}
	}

	return errors.Join(errs...)
}

func Parse(file string, pkg string) error {
	base := filepath.Base(file)

	dir := filepath.Dir(file)

	var outPath string
	if i := strings.Index(base, "."); i < 0 {
		outPath = filepath.Join(dir, fmt.Sprintf("%s_gen.go", base))
	} else {
		outPath = filepath.Join(dir, fmt.Sprintf("%s_gen%s.go", base[:i], base[i:]))
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}

	//(&MyParser2{z: NewTokenizer(f), w: LangBuf{w: out}}).Do()
	(&Trans{z: NewTokenizer(f), w: LangBuf{w: out}}).DoComponents(pkg)

	out.Close()

	//ffout, err := format.Source(ff)

	ffout, err := imports.Process(outPath, nil, nil)
	if err != nil {
		return fmt.Errorf("%w:\n=====\n%s\n=====\n", err, ffout)
	}

	return os.WriteFile(outPath, ffout, 0644)
}

func newName(path string) string {
	return ""
}
