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
	"log"
	"os"

	"github.com/ArchangelX360/winarchive-tools/msi"
	"github.com/ArchangelX360/winarchive-tools/util"
)

type MSIInfo struct {
	CABFiles []string          `json:"cab_files"`
	FileMap  map[string]string `json:"file_map"`
}

func runMSIInfo(args []string) {
	fs := flag.NewFlagSet("msi-info", flag.ExitOnError)
	inputPath := fs.String("input", "", "Path to an MSI file")
	outputPath := fs.String("out", "", "Optional JSON output path. Prints to stdout when omitted.")
	parseFlagsOrExit(fs, args)

	if *inputPath == "" {
		fs.Usage()
		log.Fatal("msi-info requires --input")
	}
	if *outputPath == "" {
		fs.Usage()
		log.Fatal("msi-info requires --out")
	}

	info, err := loadMSIInfo(*inputPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := util.WriteJSONFile(*outputPath, info); err != nil {
		log.Fatal(err)
	}
}

func loadMSIInfo(msiPath string) (*MSIInfo, error) {
	file, err := os.Open(msiPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open MSI %s: %w", msiPath, err)
	}
	defer file.Close()

	parsed, err := msi.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MSI %s: %w", msiPath, err)
	}

	fileMap := make(map[string]string, len(parsed.FileMap))
	for archivePath, outputPath := range parsed.FileMap {
		fileMap[archivePath] = outputPath
	}

	return &MSIInfo{
		CABFiles: append([]string(nil), parsed.CABFiles...),
		FileMap:  fileMap,
	}, nil
}
