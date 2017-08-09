// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/option"
)

const (
	none   = 0
	red    = 31
	green  = 32
	yellow = 33
	blue   = 34
	gray   = 90
)

// Colors mapping
var Colors = [...]int{
	log.Debug: gray,
	log.Info:  blue,
	log.Warn:  yellow,
	log.Error: red,
	log.Fatal: red,
}

// Handler implementation.
type Handler struct {
	mu       sync.Mutex
	Writer   io.Writer
	UseColor bool
}

// UseColor is a functional option with which you can force the usage of colors on or off
func UseColor(arg bool) option.Option {
	return option.Fn(func(to interface{}) error {
		if h, ok := to.(*Handler); ok {
			h.UseColor = arg
		}

		return nil
	})
}

// colorTerms contains a list of substrings that indicate support for terminal colors
var colorTerms = []string{
	"color",
	"xterm",
}

// ColorFromTerm determines from the TERM and COLORTERM environment variables wether or not
// to use colors
var ColorFromTerm = option.Fn(func(to interface{}) error {
	h, ok := to.(*Handler)
	if !ok {
		return nil
	}

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

	h.UseColor = color
	return nil
})

// defaultOptions are the default options for the handler
var defaultOptions = option.Options{
	ColorFromTerm,
}

// New returns a new handler
func New(w io.Writer, opts ...option.Option) *Handler {
	h := &Handler{
		Writer:   w,
		UseColor: false,
	}

	defaultOptions.Apply(h)
	option.Options(opts).Apply(h)

	return h
}

// HandleLog implements log.Handler
func (h *Handler) HandleLog(e log.Entry) error {
	color := Colors[e.Level()]
	level := strings.ToUpper(e.Level().String())

	var fields []field
	for k, v := range e.Fields().Fields() {
		fields = append(fields, field{k, v})
	}

	// sort the fields by name
	sort.Sort(byName(fields))

	h.mu.Lock()
	defer h.mu.Unlock()

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

// field used for sorting
type field struct {
	Name  string
	Value interface{}
}

// byName sorts fields by name
type byName []field

// Len implments sort.Sort
func (a byName) Len() int { return len(a) }

// Swap implments sort.Sort
func (a byName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less implments sort.Sort
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
