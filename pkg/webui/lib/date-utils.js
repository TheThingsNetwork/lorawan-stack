// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

export const subtractDays = (date, days) => {
  const result = new Date(date)
  result.setDate(result.getDate() - days)
  return result
}

export const generateDatesInInterval = (startDate, count) => {
  const ONE_DAY = 24 * 60 * 60 * 1000 // One day in milliseconds
  const interval = ONE_DAY / (count - 1)
  const dates = []

  for (let i = 0; i < count; i++) {
    const date = new Date(startDate.getTime() + i * interval)
    dates.push(date)
  }

  return dates
}
