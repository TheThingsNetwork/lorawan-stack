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

package test

import (
	"fmt"
	"sync"
	"time"
)

// MockClock is used to mock time package functionality.
type MockClock struct {
	nowMu sync.RWMutex
	now   time.Time

	nowChsMu sync.RWMutex
	nowChs   map[chan<- time.Time]struct{}
}

// NewMockClock returns a new MockClock.
func NewMockClock(t time.Time) *MockClock {
	return &MockClock{
		now:    t,
		nowChs: map[chan<- time.Time]struct{}{},
	}
}

func (m *MockClock) notifyNow(t time.Time) {
	m.nowChsMu.RLock()
	defer m.nowChsMu.RUnlock()
	for ch := range m.nowChs {
		ch <- t
	}
}

// Set sets clock to time t and returns old clock time.
func (m *MockClock) Set(t time.Time) time.Time {
	m.nowMu.Lock()
	defer m.nowMu.Unlock()

	old := m.now
	if old.After(t) {
		panic(fmt.Sprintf("current time (`%s`) is after the one being set (`%s`)", old, t))
	}
	m.now = t
	m.notifyNow(t)
	return old
}

// Add adds d to clock and returns new clock time.
func (m *MockClock) Add(d time.Duration) time.Time {
	m.nowMu.Lock()
	defer m.nowMu.Unlock()

	m.now = m.now.Add(d)
	m.notifyNow(m.now)
	return m.now
}

// Now returns current clock time.
func (m *MockClock) Now() time.Time {
	m.nowMu.RLock()
	defer m.nowMu.RUnlock()

	return m.now
}

// After returns a channel, on which current time.Time will be sent once d passes.
func (m *MockClock) After(d time.Duration) <-chan time.Time {
	m.nowMu.RLock()
	defer m.nowMu.RUnlock()

	if d <= 0 {
		ch := make(chan time.Time, 1)
		ch <- m.now
		close(ch)
		return ch
	}

	m.nowChsMu.Lock()
	nowCh := make(chan time.Time)
	m.nowChs[nowCh] = struct{}{}
	m.nowChsMu.Unlock()

	at := m.now.Add(d)
	afterCh := make(chan time.Time, 1)
	go func() {
		for now := range nowCh {
			if now.Before(at) {
				continue
			}
			afterCh <- now
			close(afterCh)

			go func() {
				m.nowChsMu.Lock()
				defer m.nowChsMu.Unlock()
				delete(m.nowChs, nowCh)
				close(nowCh)
			}()
			for range nowCh {
			}
		}
	}()
	return afterCh
}
