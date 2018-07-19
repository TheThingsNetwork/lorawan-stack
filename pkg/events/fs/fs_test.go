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

package fs_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/events/fs"
)

func TestWatcher(t *testing.T) {
	a := assertions.New(t)

	filename := filepath.Join(os.TempDir(), fmt.Sprintf("fs_%d", time.Now().Unix()))

	pubsub := events.NewPubSub()
	watcher, err := fs.NewWatcher(pubsub)
	a.So(err, should.BeNil)

	defer watcher.Close()

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
	a.So(err, should.BeNil)

	var i int
	var expected = []string{
		"fs.write",
		"fs.chmod",
		"fs.remove",
	}

	var wg sync.WaitGroup
	wg.Add(len(expected))

	err = watcher.Watch(filename, events.HandlerFunc(func(evt events.Event) {
		a.So(evt.Name(), should.Equal, expected[i])
		i++
		wg.Done()
	}))
	a.So(err, should.BeNil)

	file.WriteString("Hello, World!")
	file.Chmod(0640)

	file.Close()

	os.Remove(filename)

	wg.Wait()
}

func Example() {
	fs.Watch("config.yml", events.HandlerFunc(func(evt events.Event) {
		if evt.Name() != "fs.write" {
			return
		}
		fmt.Println("Detected a configuration update")
	}))
}
