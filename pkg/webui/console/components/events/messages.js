// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import { defineMessages } from 'react-intl'

const messages = defineMessages({
  // Field messages
  frmPayload: 'FRMPayload',
  devAddr: 'DevAddr',
  fPort: 'FPort',
  fCnt: 'FCnt',
  rawPayload: 'Raw payload',
  txPower: 'Tx Power',
  bandwidth: 'Bandwidth',
  metrics: 'Metrics',
  versions: 'Versions',
  snr: 'SNR',
  sessionKeyId: 'Session key ID',
  selectedMacVersion: 'Selected MAC version',
  // Generic messages
  eventDetails: 'Event details',
  errorOverviewEntry:
    'There was an error and the event cannot be displayed. The raw event can by viewed by clicking this row.',
  dataPreview: 'Data preview',
  seeAllActivity: 'See all activity',
  syntheticEvent: 'Note: This meta event did not originate from the event stream',
  eventsTruncated: 'Old events have been truncated to save memory',
})

export default messages
