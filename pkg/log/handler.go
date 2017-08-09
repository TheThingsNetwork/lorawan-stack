// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

// Handler is the interface of things that can handle log entries
type Handler interface {
	HandleLog(Entry) error
}
