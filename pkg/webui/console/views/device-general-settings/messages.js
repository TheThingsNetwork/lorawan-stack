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

import { defineMessages } from 'react-intl'

const messages = defineMessages({
  isTitle: 'Basic',
  isDescription: 'Description, cluster information and metadata',
  isDescriptionMissing: 'Identity Server is not available',
  asTitle: 'Application layer',
  asDescription: 'Application-layer behavior and session',
  asDescriptionMissing: 'Application Server is not available',
  asDescriptionOTAA: 'Only keys of joined OTAA end devices are stored on the Application Server',
  jsTitle: 'Join settings',
  jsDescription: 'Root keys and network settings for end device activation',
  jsDescriptionMissing: 'Join Server is not available',
  jsDescriptionOTAA: 'ABP/multicast end devices are not stored on the Join Server',
  nsTitle: 'Network layer',
  nsDescription: 'LoRaWAN network-layer settings, behavior and session',
  nsDescriptionMissing: 'Network Server is not available',
  deleteSuccess: 'End device deleted',
  deleteFailure: 'An error occurred and the end device could not be deleted',
  activationModeUnknown: 'Activation mode unknown because Network Server is not available',
  notInCluster: 'Not registered in this cluster',
  updateSuccess: 'End device updated',
  keysResetWarning:
    'You do not have sufficient rights to view end device keys. Only overwriting is allowed.',
  validateSessionKey: '{field} must have non-zero value',
  macSettingsError:
    'There was an error and the default MAC settings for the <code>{freqPlan}</code> frequency plan could not be loaded',
  fpNotFoundError:
    'The LoRaWAN version <code>{lorawanVersion}</code> does not support the <code>{freqPlan}</code> frequency plan. Please choose a different MAC version or frequency plan.',
})

export default messages
