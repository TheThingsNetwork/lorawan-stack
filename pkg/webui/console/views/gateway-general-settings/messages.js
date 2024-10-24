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
  lorawanDescription: 'LoRaWAN network-layer settings',
  updateSuccess: 'Gateway updated',
  modalWarning:
    'Are you sure you want to delete "{gtwName}"? This action cannot be undone and it will not be possible to reuse the gateway ID.',
  disablePacketBrokerForwarding:
    'Disable forwarding uplink messages received from this gateway to the Packet Broker',
  adminContactDescription:
    'Administrative contact information for this gateway. Typically used to indicate who to contact with administrative questions about the gateway.',
  techContactDescription:
    'Technical contact information for this gateway. Typically used to indicate who to contact with technical/security questions about the gateway.',
  deleteGatewayDefaultMessage:
    'This will <strong>PERMANENTLY DELETE THE ENTITY ITSELF AND ALL ASSOCIATED ENTITIES</strong>, including collaborator associations. It will also <strong>NOT BE POSSIBLE TO REUSE THE ENTITY ID</strong> until purged by an admin but the EUI can be reregistered later with a different ID.',
})

export default messages
