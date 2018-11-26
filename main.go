package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var (
	inputFileFlag = flag.String("input", "./data/client.jar", "input file to read")
	outputDirFlag = flag.String("output", "./dump/client", "directory to output files")
	prefixFlag    = flag.String("prefix", "assets/", "file prefix to filter")
	dirOnlyFlag   = flag.Bool("dir", false, "only output directories")
	verboseFlag   = flag.Bool("v", false, "output every file")
)

func main() {
	flag.Parse()

	exitIf(os.MkdirAll(*outputDirFlag, 0666))

	reader, err := zip.OpenReader(*inputFileFlag)
	exitIf(errors.Wrapf(err, "failed to open file %q", *inputFileFlag))

	for _, file := range reader.File {
		if !strings.HasPrefix(file.Name, *prefixFlag) {
			continue
		}

		if file.Mode().IsDir() {
			// only output directories, this can be noisy
			logf("Reading directory %q", file.Name)
		}

		if *dirOnlyFlag {
			continue
		}

		if *verboseFlag {
			logf("Saving file %q", filepath.Join(*outputDirFlag, file.Name))
		}

		if err := saveToDisk(*outputDirFlag, file); err != nil {
			errf(err)
		}
	}
}

func saveToDisk(path string, file *zip.File) error {
	filename := filepath.Join(path, file.Name)

	if file.Mode().IsDir() {
		return os.MkdirAll(filename, 0666)
	}

	if err := os.MkdirAll(filepath.Dir(filename), 0666); err != nil {
		return errors.Wrapf(err, "failed to create directory for file %q", file.Name)
	}

	f, err := file.Open()
	if err != nil {
		return errors.Wrapf(err, "failed to open file %q", file.Name)
	}
	defer f.Close()

	output, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %q", filename)
	}
	defer output.Close()

	_, err = io.Copy(output, f)

	return errors.Wrapf(err, "failed to copy file %q", filename)
}

func logf(m string, v ...interface{}) {
	fmt.Fprintf(os.Stdout, fmt.Sprintf("%s\n", m), v...)
}

func errf(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
}

func exitIf(err error) {
	if err == nil {
		return
	}

	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
