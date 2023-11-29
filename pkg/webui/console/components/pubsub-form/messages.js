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

export default defineMessages({
  idPlaceholder: 'my-new-pubsub',
  messageInfo: 'For each enabled message type an optional sub-topic can be defined',
  deletePubsub: 'Delete Pub/Sub',
  modalWarning:
    'Are you sure you want to delete Pub/Sub "{pubsubId}"? Deleting a Pub/Sub cannot be undone.',
  headers: 'Headers',
  headersValidateRequired: 'All header entry values are required. Please remove empty entries.',
  usernamePlaceholder: 'my-username',
  passwordPlaceholder: 'my-password',
  natsAddressPlaceholder: 'nats.example.com',
  natsPortPlaceholder: '4222',
  natsConfig: 'NATS configuration',
  mqttConfig: 'MQTT configuration',
  mqttClientIdPlaceholder: 'my-client-id',
  mqttServerUrlPlaceholder: 'mqtts://example.com',
  subscribeQos: 'Subscribe QoS',
  providerDescription: 'Changing the Pub/Sub provider has been disabled by an administrator',
  publishQos: 'Publish QoS',
  tlsCa: 'Root CA certificate',
  tlsClientCert: 'Client certificate',
  tlsClientKey: 'Client private key',
  selectPemFile: 'Select .pem file…',
  pemFileProvided: '.pem file has been provided',
  useCredentials: 'Use credentials',
  alreadyExistsModalMessage:
    'A Pub/Sub with the ID "{id}" already exists. Do you wish to replace this Pub/Sub?',
  replacePubsub: 'Replace Pub/Sub',
  useSecureConnection: 'Use secure connection',
  pubsubsDescription:
    'The Pub/Sub integration allows the Application Server to publish and subscribe to topics, using The Things Stack built-in MQTT client or NATS client. Learn more in our <Link>Pub/Sub guide</Link>.',
})
