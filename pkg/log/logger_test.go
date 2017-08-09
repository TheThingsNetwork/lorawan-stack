// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"testing"

	. "github.com/smartystreets/assertions"
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
	a := New(t)

	rec := newRecorder()
	logger := &Logger{
		Level:   Info,
		Handler: rec,
	}

	logger.Debug("Yo!")
	a.So(rec.entries, ShouldHaveLength, 0)

	logger.Warn("Hi!")
	a.So(rec.entries, ShouldHaveLength, 1)
	{
		entry := rec.entries[0]
		a.So(entry.Message(), ShouldEqual, "Hi!")
	}

	logger.Infof("Hey, %s!", "you")
	a.So(rec.entries, ShouldHaveLength, 2)
	{
		entry := rec.entries[1]
		a.So(entry.Message(), ShouldEqual, "Hey, you!")
	}

	other := logger.WithFields(Fields(
		"foo", 10,
		"bar", "baz",
	))

	logger.Info("Ok!")
	a.So(rec.entries, ShouldHaveLength, 3)
	{
		entry := rec.entries[2]
		a.So(entry.Message(), ShouldEqual, "Ok!")
		a.So(entry.Fields().Fields(), ShouldBeEmpty)
	}

	other.Info("Nice!")
	a.So(rec.entries, ShouldHaveLength, 4)
	{
		entry := rec.entries[3]
		a.So(entry.Message(), ShouldEqual, "Nice!")
		a.So(entry.Fields().Fields(), ShouldHaveLength, 2)
		fields := entry.Fields().Fields()
		a.So(fields["foo"], ShouldEqual, 10)
		a.So(fields["bar"], ShouldEqual, "baz")
	}
}
