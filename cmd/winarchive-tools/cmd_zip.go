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
	"archive/zip"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/ArchangelX360/winarchive-tools/util"
)

type ZIPListing struct {
	Entries []string `json:"entries"`
}

func runZIPList(args []string) {
	fs := flag.NewFlagSet("zip-list", flag.ExitOnError)
	inputPath := fs.String("input", "", "Path to a ZIP or VSIX archive")
	outputPath := fs.String("out", "", "Optional JSON output path. Prints to stdout when omitted.")
	parseFlagsOrExit(fs, args)

	if *inputPath == "" {
		fs.Usage()
		log.Fatal("zip-list requires --input")
	}
	if *outputPath == "" {
		fs.Usage()
		log.Fatal("zip-list requires --out")
	}

	entries, err := listZIPEntries(*inputPath)
	if err != nil {
		log.Fatal(err)
	}
	listing := ZIPListing{Entries: entries}
	if err := util.WriteJSONFile(*outputPath, listing); err != nil {
		log.Fatal(err)
	}
}

func runZIPExtract(args []string) {
	fs := flag.NewFlagSet("zip-extract", flag.ExitOnError)
	inputPath := fs.String("input", "", "Path to a ZIP or VSIX archive")
	layoutPath := fs.String("layout", "", "Path to an extract layout JSON file")
	outDir := fs.String("out-dir", "", "Directory to extract files into")
	parseFlagsOrExit(fs, args)

	if *inputPath == "" || *layoutPath == "" || *outDir == "" {
		fs.Usage()
		log.Fatal("zip-extract requires --input, --layout, and --out-dir")
	}

	layout, err := util.LoadExtractLayout(*layoutPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := extractZIPFiles(*inputPath, layout, *outDir); err != nil {
		log.Fatal(err)
	}
}

func listZIPEntries(zipPath string) ([]string, error) {
	file, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP %s: %w", zipPath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat ZIP %s: %w", zipPath, err)
	}
	archive, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ZIP %s: %w", zipPath, err)
	}

	entries := make([]string, 0, len(archive.File))
	for _, archiveFile := range archive.File {
		if archiveFile.FileInfo().IsDir() {
			continue
		}
		entries = append(entries, archiveFile.Name)
	}
	sort.Strings(entries)
	return entries, nil
}

func extractZIPFiles(zipPath string, layout map[string]string, outDir string) error {
	file, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP %s: %w", zipPath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat ZIP %s: %w", zipPath, err)
	}
	archive, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to parse ZIP %s: %w", zipPath, err)
	}

	remaining := util.CopyEntryMap(layout)
	for _, archiveFile := range archive.File {
		outputPath, ok := remaining[archiveFile.Name]
		if !ok || archiveFile.FileInfo().IsDir() {
			continue
		}

		reader, err := archiveFile.Open()
		if err != nil {
			return fmt.Errorf("failed to open %s inside %s: %w", archiveFile.Name, zipPath, err)
		}
		if err := util.WriteOutputFile(outDir, outputPath, archiveFile.FileInfo().ModTime(), reader); err != nil {
			reader.Close()
			return fmt.Errorf("failed to extract %s from %s: %w", archiveFile.Name, zipPath, err)
		}
		if err := reader.Close(); err != nil {
			return fmt.Errorf("failed to close %s inside %s: %w", archiveFile.Name, zipPath, err)
		}
		delete(remaining, archiveFile.Name)
	}

	if len(remaining) > 0 {
		return fmt.Errorf("failed to find %d requested ZIP entries: %s", len(remaining), util.FormatMissingEntries(remaining))
	}
	return nil
}
