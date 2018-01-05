// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"os"
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var Marshal = marshal

func TestUnflattened(t *testing.T) {
	for _, tc := range []struct {
		in  map[string]interface{}
		out map[string]interface{}
	}{
		{
			map[string]interface{}{
				"foo.bar":             os.Stdout,
				"foo.baz":             map[string]string{"test": "foo"},
				"foo.recursive.hello": struct{ hi string }{"hello"},
				"42.foo":              42,
				"42.baz":              "baz",
			},
			map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": os.Stdout,
					"baz": map[string]string{"test": "foo"},
					"recursive": map[string]interface{}{
						"hello": struct{ hi string }{"hello"},
					},
				},
				"42": map[string]interface{}{
					"foo": 42,
					"baz": "baz",
				},
			},
		},
	} {
		assertions.New(t).So(unflattened(tc.in), should.Resemble, tc.out)
	}
}

func TestSlicify(t *testing.T) {
	for i, tc := range []struct {
		in  map[string]interface{}
		out map[string]interface{}
	}{
		{
			map[string]interface{}{
				"foo": map[string]interface{}{
					"2": "two",
					"5": "five",
					"1": "one",
				},
				"bar": map[string]interface{}{
					"11": "eleven",
					"3":  "three",
					"0":  "zero",
				},
				"baz": map[string]interface{}{
					"3":  "three",
					"0":  "zero",
					"hi": "there",
				},
			},
			map[string]interface{}{
				"foo": []interface{}{
					nil,
					"one",
					"two",
					nil,
					nil,
					"five",
				},
				"bar": []interface{}{
					"zero",
					nil,
					nil,
					"three",
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					"eleven",
				},
				"baz": map[string]interface{}{
					"3":  "three",
					"0":  "zero",
					"hi": "there",
				},
			},
		},
		{
			map[string]interface{}{
				"foo": map[string]interface{}{
					"0": map[string]interface{}{
						"1": map[string]interface{}{
							"hello": "hi",
						},
					},
					"2": "two",
				},
			},
			map[string]interface{}{
				"foo": []interface{}{
					[]interface{}{
						nil,
						map[string]interface{}{
							"hello": "hi",
						},
					},
					nil,
					"two",
				},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assertions.New(t).So(slicify(tc.in), should.Resemble, tc.out)
		})
	}
}
