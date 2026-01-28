package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/cedar-policy/cedar-go/x/exp/schema"
)

func main() {
	inPath := flag.String("in", "corpus-tests.tar.gz", "input corpus tar.gz")
	outPath := flag.String("out", "corpus-tests-json-schemas.tar.gz", "output tar.gz of JSON schemas")
	allowErrors := flag.Bool("allow-errors", true, "continue on schema translation errors")
	flag.Parse()

	if err := run(*inPath, *outPath, *allowErrors); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(inPath, outPath string, allowErrors bool) error {
	inFile, err := os.Open(inPath)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer func() { _ = inFile.Close() }()

	gzr, err := gzip.NewReader(inFile)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer func() { _ = gzr.Close() }()

	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	gzw := gzip.NewWriter(outFile)
	defer func() { _ = gzw.Close() }()

	tw := tar.NewWriter(gzw)
	defer func() { _ = tw.Close() }()

	tr := tar.NewReader(gzr)

	var failures int
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if !strings.HasSuffix(hdr.Name, ".cedarschema") {
			continue
		}

		cedarBytes, err := io.ReadAll(tr)
		if err != nil {
			return fmt.Errorf("read %s: %w", hdr.Name, err)
		}

		var s schema.Schema
		s.SetFilename(hdr.Name)
		err = s.UnmarshalCedar(cedarBytes)
		var outBytes []byte
		if err == nil {
			outBytes, err = s.MarshalJSON()
		}
		if err != nil {
			failures++
			if !allowErrors {
				return fmt.Errorf("translate %s: %w", hdr.Name, err)
			}
			outBytes = []byte(fmt.Sprintf("error translating %s: %v\n", hdr.Name, err))
		}

		base := strings.TrimSuffix(path.Base(hdr.Name), ".cedarschema")
		outName := path.Join("corpus-tests-json-schemas", base+".cedarschema.json")
		outHdr := &tar.Header{
			Name: outName,
			Mode: 0o644,
			Size: int64(len(outBytes)),
		}
		if err := tw.WriteHeader(outHdr); err != nil {
			return fmt.Errorf("write header for %s: %w", outName, err)
		}
		if _, err := tw.Write(outBytes); err != nil {
			return fmt.Errorf("write %s: %w", outName, err)
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("finalize tar: %w", err)
	}
	if err := gzw.Close(); err != nil {
		return fmt.Errorf("finalize gzip: %w", err)
	}
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("close output: %w", err)
	}

	if failures > 0 {
		fmt.Fprintf(os.Stderr, "warning: failed to translate %d schema(s) (see output files for details)\n", failures)
	}
	return nil
}
