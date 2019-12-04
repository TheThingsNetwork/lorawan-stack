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
  asTitle: 'Application Layer',
  asDescription: 'Application-layer behavior and session',
  asDescriptionMissing: 'Application Server is not available',
  asDescriptionOTAA: 'Application Server does not store any OTAA specific device fields',
  jsTitle: 'Join',
  jsDescription: 'Root keys and network settings for device activation',
  jsDescriptionMissing: 'Join Server is not available',
  jsDescriptionOTAA: 'ABP/multicast devices are not stored on the Join Server',
  nsTitle: 'Network Layer',
  nsDescription: 'LoRaWAN network-layer settings, behavior and session',
  nsDescriptionMissing: 'Network Server is not available',
  deleteSuccess: 'The end device has been deleted successfully',
  deleteFailure: 'End device deletion failed',
  activationModeUnknown: 'Activation mode unknown because Network Server is not available',
  notInCluster: 'Not registered in this cluster',
  updateSuccess: 'The end device has been updated successfully',
})

export default messages
