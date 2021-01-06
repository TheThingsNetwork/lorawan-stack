// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package io

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"

	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
)

// Write output to Stdout.
// Uses either JSON or formats using the configured template.
func Write(w io.Writer, format string, data interface{}) (err error) {
	defer func() {
		fmt.Fprintln(w)
	}()
	rv := reflect.Indirect(reflect.ValueOf(data))
	switch rv.Type().Kind() {
	case reflect.Slice:
	case reflect.Struct:
	default:
		panic(fmt.Sprintf("unsupported value: %T", data))
	}
	var prefix, sep, suffix []byte
	var writeItem func(interface{}) error
	switch format {
	case "json":
		jsonpb := jsonpb.TTN()
		jsonpb.Indent = "  "
		encoder := jsonpb.NewEncoder(w)
		prefix, sep, suffix = []byte("["), []byte(", "), []byte("]")
		writeItem = func(v interface{}) error {
			return encoder.Encode(v)
		}
	default:
		format = strings.TrimSpace(format)
		tmpl, err := template.New("").Parse(format)
		if err != nil {
			return err
		}
		sep = []byte("\n")
		writeItem = func(v interface{}) error {
			return tmpl.Execute(w, v)
		}
	}
	if rv.Type().Kind() == reflect.Struct {
		return writeItem(data)
	}
	if prefix != nil {
		_, err = w.Write(prefix)
		if err != nil {
			return err
		}
	}
	n := rv.Len()
	for i := 0; i < n; i++ {
		if err = writeItem(rv.Index(i).Interface()); err != nil {
			return err
		}
		if sep != nil && i != n-1 {
			_, err = w.Write(sep)
			if err != nil {
				return err
			}
		}
	}
	if suffix != nil {
		_, err = w.Write(suffix)
		if err != nil {
			return err
		}
	}
	return nil
}

// BufferedPipe returns a buffered reader if the reader is a pipe that can be read.
func BufferedPipe(r io.Reader) (*bufio.Reader, bool) {
	if f, ok := r.(*os.File); ok {
		if stat, err := f.Stat(); err == nil && stat.Mode()&os.ModeCharDevice == 0 {
			rd := bufio.NewReader(r)
			if n, err := rd.Peek(1); err == nil && len(n) == 1 {
				return rd, true
			}
		}
	}
	return nil, false
}
