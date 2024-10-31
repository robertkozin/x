package htmgo

import (
	"errors"
	"fmt"
	"github.com/samber/oops"
	"golang.org/x/tools/imports"
	"os"
	"path/filepath"
	"strings"
)

// input dir, output file, output package

func Generate2(inputDir string, outputFilepath string, outputPackage string) error {

	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		oops.Wrap(err)
	}
	defer outputFile.Close()

	langBuf := LangBuf{w: outputFile}

	langBuf.Gof("package %s\n\n", outputPackage)
	langBuf.Gof(`import (
	"io"
	"fmt"
	"context"
	"github.com/robertkozin/x/htmgo"
	)
	
	`)
	langBuf.Flush()

	filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(info.Name(), ".html") {
			return nil
		}

		inputFilepath := filepath.Join(inputDir, info.Name())
		inputFile, err := os.Open(inputFilepath)
		if err != nil {
			return err
		}
		defer inputFile.Close()

		// parse and append

		(&Trans{z: NewTokenizer(inputFile), w: LangBuf{w: outputFile}}).DoComponents2()

		return nil
	})

	outputFile.Close()

	formattedFileBytes, err := imports.Process(outputFilepath, nil, nil)
	if err != nil {
		return oops.Wrap(err)
	}

	return os.WriteFile(outputFilepath, formattedFileBytes, 0644)
}

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
