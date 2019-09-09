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

const isGsEvent = eventName => Boolean(eventName) && eventName.startsWith('gs.')
const isGsUplinkEvent = eventName => isGsEvent(eventName) && eventName.includes('.up.')
const isGsDownlinkEvent = eventName => isGsEvent(eventName) && eventName.includes('.down.')
export const isGsStatusReceiveEvent = eventName =>
  isGsEvent(eventName) && eventName.includes('.status.receive')
export const isGsUplinkReceiveEvent = eventName =>
  isGsEvent(eventName) && isGsUplinkEvent && eventName.endsWith('.receive')
export const isGsDownlinkSendEvent = eventName =>
  isGsEvent(eventName) && isGsDownlinkEvent(eventName) && eventName.includes('.send')
