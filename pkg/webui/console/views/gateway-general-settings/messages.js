// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
  basicTitle: 'Basic settings',
  basicDescription: 'General settings, gateway updates and metadata',
  lorawanTitle: 'LoRaWAN options',
  lorawanDescription: 'LoRaWAN network-layer settings',
  updateSuccess: 'Gateway updated',
  deleteSuccess: 'Gateway deleted',
  deleteFailure: 'Gateway delete error',
  deleteGateway: 'Delete gateway',
  modalWarning:
    'Are you sure you want to delete "{gtwName}"? This action cannot be undone and it will not be possible to reuse the gateway ID.',
})

export default messages
