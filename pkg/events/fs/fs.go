// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Package fs implements watching files for changes.
package fs

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
)

// Watcher interface for watching the filesystem for changes.
type Watcher interface {
	// Watch a file for changes, the handler may be nil.
	Watch(name string, handler events.Handler) error
	Close() error
}

type watcher struct {
	*fsnotify.Watcher
	mu          sync.RWMutex
	subscribers map[string][]events.Handler
}

func (w *watcher) Watch(name string, handler events.Handler) error {
	name = filepath.Clean(name)
	err := w.Watcher.Add(name)
	if err != nil {
		return err
	}
	if handler != nil {
		w.mu.Lock()
		w.subscribers[name] = append(w.subscribers[name], handler)
		w.mu.Unlock()
	}
	return nil
}

func (w *watcher) Close() error {
	events.Unsubscribe("fs.*", w)
	return w.Watcher.Close()
}

func (w *watcher) Notify(evt events.Event) {
	name, ok := evt.Data().(string)
	if !ok {
		return
	}
	w.mu.RLock()
	subscribers := w.subscribers[name]
	w.mu.RUnlock()
	for _, subscriber := range subscribers {
		subscriber.Notify(evt)
	}
}

// NewWatcher returns a new filesystem Watcher that publishes events to the given PubSub.
// Event names will follow the pattern `fs.<type>` where the type can be create,
// write, remove, rename or chmod.
// The event identifiers will be nil, and the event payload will be the filename.
func NewWatcher(pubsub events.PubSub) (Watcher, error) {
	var err error
	w := &watcher{
		subscribers: make(map[string][]events.Handler),
	}
	w.Watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = pubsub.Subscribe("fs.*", w)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					pubsub.Publish(events.New(context.Background(), "fs.create", nil, event.Name))
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					pubsub.Publish(events.New(context.Background(), "fs.write", nil, event.Name))
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					pubsub.Publish(events.New(context.Background(), "fs.remove", nil, event.Name))
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					pubsub.Publish(events.New(context.Background(), "fs.rename", nil, event.Name))
				}
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					pubsub.Publish(events.New(context.Background(), "fs.chmod", nil, event.Name))
				}
			case err := <-w.Errors:
				log.WithError(err).Warn("Error in file watcher")
			}
		}
	}()
	return w, nil
}

// DefaultWatcher is the default filesystem Watcher.
// This watcher works on top of an isolated events PubSub.
var DefaultWatcher Watcher

func init() {
	var err error
	DefaultWatcher, err = NewWatcher(events.NewPubSub(events.DefaultBufferSize))
	if err != nil {
		log.WithError(err).Warn("Could not initialize filesystem watcher")
	}
}

// Watch a file on the default filesystem Watcher.
// The handler may be nil.
func Watch(name string, handler events.Handler) error {
	if DefaultWatcher == nil {
		return nil
	}
	return DefaultWatcher.Watch(name, handler)
}
