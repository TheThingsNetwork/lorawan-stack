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

export function get (key) {
  const value = localStorage.getItem(key)
  try {
    return JSON.parse(value)
  } catch (e) {
    return value
  }
}

export function set (key, val) {
  const value = JSON.stringify(val)
  localStorage.setItem(key, value)
}

export function remove (key) {
  return localStorage.removeItem(key)
}

export function clearAll () {
  return localStorage.clear()
}
