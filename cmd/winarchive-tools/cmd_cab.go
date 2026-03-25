// Copyright (c) 2022 Lorenz Brun
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// This license does not apply to the cab subdirectory which is licensed under
// Apache 2.0. See the package comment for the license text.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ArchangelX360/winarchive-tools/cab"
	"github.com/ArchangelX360/winarchive-tools/util"
)

type stringListFlag []string

func (f *stringListFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *stringListFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func runCABExtract(args []string) {
	fs := flag.NewFlagSet("cab-extract", flag.ExitOnError)
	layoutPath := fs.String("layout", "", "Path to an extract layout JSON file")
	outDir := fs.String("out-dir", "", "Directory to extract files into")
	var cabPaths stringListFlag
	fs.Var(&cabPaths, "cab", "Path to a CAB file. May be repeated.")
	parseFlagsOrExit(fs, args)

	if *layoutPath == "" || *outDir == "" || len(cabPaths) == 0 {
		fs.Usage()
		log.Fatal("cab-extract requires --layout, --out-dir, and at least one --cab")
	}

	layout, err := util.LoadExtractLayout(*layoutPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := extractCABFiles(cabPaths, layout, *outDir); err != nil {
		log.Fatal(err)
	}
}

func extractCABFiles(cabPaths []string, layout map[string]string, outDir string) error {
	remaining := util.CopyEntryMap(layout)

	for _, cabPath := range cabPaths {
		cabFile, err := os.Open(cabPath)
		if err != nil {
			return fmt.Errorf("failed to open CAB %s: %w", cabPath, err)
		}

		cabReader, err := cab.New(cabFile)
		if err != nil {
			cabFile.Close()
			return fmt.Errorf("failed to parse CAB %s: %w", cabPath, err)
		}

		for {
			header, err := cabReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				cabFile.Close()
				return fmt.Errorf("failed to read CAB %s: %w", cabPath, err)
			}

			outputPath, ok := remaining[header.Name]
			if !ok {
				continue
			}
			if err := util.WriteOutputFile(outDir, outputPath, header.CreateTime, cabReader); err != nil {
				cabFile.Close()
				return fmt.Errorf("failed to extract %s from %s: %w", header.Name, cabPath, err)
			}
			delete(remaining, header.Name)
		}

		if err := cabFile.Close(); err != nil {
			return fmt.Errorf("failed to close CAB %s: %w", cabPath, err)
		}
	}

	if len(remaining) > 0 {
		return fmt.Errorf("failed to find %d requested CAB entries: %s", len(remaining), util.FormatMissingEntries(remaining))
	}
	return nil
}
