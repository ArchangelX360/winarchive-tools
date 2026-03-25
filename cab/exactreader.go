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

package cab

import "io"

// ExactReader returns a Reader that reads from r
// but stops with EOF after n bytes. It returns ErrUnexpectedEOF if
// the underlying reader returns EOF before n bytes.
// The underlying implementation is a *ExactReader.
func ExactReader(r io.Reader, n int64) io.ReadCloser { return &ExactReaderImpl{r, n} }

// A ExactReaderImpl reads from R but limits the amount of
// data returned to just N bytes. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
type ExactReaderImpl struct {
	R io.Reader // underlying reader
	N int64     // max bytes remaining
}

func (e *ExactReaderImpl) Read(p []byte) (n int, err error) {
	if e.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > e.N {
		p = p[0:e.N]
	}
	n, err = e.R.Read(p)
	e.N -= int64(n)
	if err == io.EOF && e.N > 0 {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (e *ExactReaderImpl) Close() error {
	_, err := io.Copy(io.Discard, e)
	return err
}
