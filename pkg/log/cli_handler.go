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

package log

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	isatty "github.com/mattn/go-isatty"
)

const (
	red    = 31
	yellow = 33
	blue   = 34
	gray   = 90
)

// Colors mapping.
var Colors = [...]int{
	DebugLevel: gray,
	InfoLevel:  blue,
	WarnLevel:  yellow,
	ErrorLevel: red,
	FatalLevel: red,
}

// CLIHandler implements Handler.
type CLIHandler struct {
	mu       sync.Mutex
	Writer   io.Writer
	UseColor bool
}

// CLIHandlerOption is the type of options for the CLIHandler.
type CLIHandlerOption func(*CLIHandler)

// UseColor is a functional option with which you can force the usage of colors on or off.
func UseColor(arg bool) CLIHandlerOption {
	return func(handler *CLIHandler) {
		handler.UseColor = arg
	}
}

// colorTerms contains a list of substrings that indicate support for terminal colors.
var colorTerms = []string{
	"color",
	"xterm",
}

// ColorFromTerm determines from the TERM and COLORTERM environment variables wether or not to use colors.
// If set, colors will be enabled in these cases:
// - COLORTERM is set and has a value different from 0
// - TERM contains the substring "xterm" or "color" and COLORTERM is not 0
var ColorFromTerm = func(handler *CLIHandler) {
	COLORTERM := os.Getenv("COLORTERM")
	TERM := os.Getenv("TERM")

	// use colors if COLORTERM is set
	color := COLORTERM != ""

	// check all color term possibilities in TERM
	for _, substring := range colorTerms {
		color = color || strings.Contains(TERM, substring)
	}

	// COLORTERM=0 forces colors off
	color = color && COLORTERM != "0"

	if out, ok := handler.Writer.(*os.File); ok {
		// Only use color on terminals
		color = color && (isatty.IsTerminal(out.Fd()) || isatty.IsCygwinTerminal(out.Fd()))
	}

	handler.UseColor = color
}

// defaultCLIOptions are the default options for the handler.
var defaultCLIOptions = []CLIHandlerOption{
	ColorFromTerm,
}

// NewCLI returns a new CLIHandler.
func NewCLI(w io.Writer, opts ...CLIHandlerOption) *CLIHandler {
	handler := &CLIHandler{
		Writer:   w,
		UseColor: false,
	}

	for _, opt := range append(defaultCLIOptions, opts...) {
		opt(handler)
	}

	return handler
}

// HandleLog implements Handler.
func (h *CLIHandler) HandleLog(e Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	color := Colors[e.Level()]
	level := strings.ToUpper(e.Level().String())

	var fields []field
	for k, v := range e.Fields().Fields() {
		fields = append(fields, field{k, v})
	}

	// sort the fields by name
	sort.Sort(byName(fields))

	if h.UseColor {
		fmt.Fprintf(h.Writer, "\033[%dm%6s\033[0m %-40s", color, level, e.Message())
	} else {
		fmt.Fprintf(h.Writer, "%6s %-40s", level, e.Message())
	}

	for _, f := range fields {
		var value interface{}
		switch t := f.Value.(type) {
		case []byte:
			value = fmt.Sprintf("%X", t)
		default:
			value = f.Value
		}

		if h.UseColor {
			fmt.Fprintf(h.Writer, " \033[%dm%s\033[0m=%v", color, f.Name, value)
		} else {
			fmt.Fprintf(h.Writer, " %s=%v", f.Name, value)
		}
	}

	fmt.Fprintln(h.Writer)

	return nil
}

// field used for sorting.
type field struct {
	Name  string
	Value interface{}
}

// byName sorts fields by name.
type byName []field

// Len implments sort.Sort.
func (a byName) Len() int { return len(a) }

// Swap implments sort.Sort.
func (a byName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less implments sort.Sort.
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
