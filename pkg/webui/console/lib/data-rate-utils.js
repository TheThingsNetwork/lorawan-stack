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

export default (data, selector = 'settings') => {
  if (!data) {
    return undefined
  }
  const { [selector]: container } = data
  if (!container) {
    return undefined
  }
  const { data_rate } = container
  if (!data_rate) {
    return undefined
  }
  const { lora, fsk, lrfhss } = data_rate
  // The encoding below mimics the encoding of the `modu` field of the UDP packet forwarder.
  if (lora) {
    const { bandwidth, spreading_factor } = lora
    return `SF${spreading_factor}BW${bandwidth / 1000}`
  } else if (fsk) {
    const { bit_rate } = fsk
    return `${bit_rate}`
  } else if (lrfhss) {
    const { modulation_type, operating_channel_width } = lrfhss
    return `M${modulation_type ?? 0}CW${operating_channel_width / 1000}`
  }
  return undefined
}
