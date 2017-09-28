// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

// recorder is log.Handler that records the entries
type recorder struct {
	entries []Entry
}

// HandleLog implements Handler
func (r *recorder) HandleLog(e Entry) error {
	r.entries = append(r.entries, e)
	return nil
}

func newRecorder() *recorder {
	return &recorder{
		entries: make([]Entry, 0, 10),
	}
}

func TestLogger(t *testing.T) {
	a := assertions.New(t)

	rec := newRecorder()
	logger := &Logger{
		Level:   InfoLevel,
		Handler: rec,
	}

	logger.Debug("Yo!")
	a.So(rec.entries, should.HaveLength, 0)

	logger.Warn("Hi!")
	a.So(rec.entries, should.HaveLength, 1)
	{
		entry := rec.entries[0]
		a.So(entry.Message(), should.Equal, "Hi!")
	}

	logger.Infof("Hey, %s!", "you")
	a.So(rec.entries, should.HaveLength, 2)
	{
		entry := rec.entries[1]
		a.So(entry.Message(), should.Equal, "Hey, you!")
	}

	other := logger.WithFields(Fields(
		"foo", 10,
		"bar", "baz",
	))

	logger.Info("Ok!")
	a.So(rec.entries, should.HaveLength, 3)
	{
		entry := rec.entries[2]
		a.So(entry.Message(), should.Equal, "Ok!")
		a.So(entry.Fields().Fields(), should.BeEmpty)
	}

	other.Info("Nice!")
	a.So(rec.entries, should.HaveLength, 4)
	{
		entry := rec.entries[3]
		a.So(entry.Message(), should.Equal, "Nice!")
		a.So(entry.Fields().Fields(), should.HaveLength, 2)
		fields := entry.Fields().Fields()
		a.So(fields["foo"], should.Equal, 10)
		a.So(fields["bar"], should.Equal, "baz")
	}
}

func TestMiddleware(t *testing.T) {
	a := assertions.New(t)

	rec := newRecorder()
	logger := &Logger{
		Level:   InfoLevel,
		Handler: rec,
	}

	before := []int{}
	after := []int{}

	logger.Use(MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(entry Entry) error {
			before = append(before, 1)
			err := next.HandleLog(entry)
			after = append(after, 1)

			return err
		})
	}))

	logger.Info("Hey!")

	a.So(before, should.Resemble, []int{1})
	a.So(after, should.Resemble, []int{1})

	a.So(rec.entries, should.HaveLength, 1)

	// reset
	before = []int{}
	after = []int{}

	logger.Use(MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(entry Entry) error {
			before = append(before, 2)
			err := next.HandleLog(entry)
			after = append(after, 2)

			return err
		})
	}))

	logger.Info("Hey!")

	a.So(before, should.Resemble, []int{1, 2})
	a.So(after, should.Resemble, []int{2, 1})

	a.So(rec.entries, should.HaveLength, 2)
}
