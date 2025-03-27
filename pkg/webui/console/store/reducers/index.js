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

import { combineReducers } from 'redux'

import {
  APPLICATION,
  END_DEVICE,
  GATEWAY,
  USER,
  ORGANIZATION,
  CLIENT,
} from '@console/constants/entities'

import {
  getUserId,
  getApplicationId,
  getGatewayId,
  getOrganizationId,
  getCombinedDeviceId,
  getApiKeyId,
  getCollaboratorId,
  getPacketBrokerNetworkId,
  getClientId,
} from '@ttn-lw/lib/selectors/id'
import { createNamedPaginationReducer } from '@ttn-lw/lib/store/reducers/pagination'
import fetching from '@ttn-lw/lib/store/reducers/ui/fetching'
import error from '@ttn-lw/lib/store/reducers/ui/error'
import status from '@ttn-lw/lib/store/reducers/status'
import init from '@ttn-lw/lib/store/reducers/init'
import collaborators from '@ttn-lw/lib/store/reducers/collaborators'
import searchAccounts from '@ttn-lw/lib/store/reducers/search-accounts'
import { SHARED_NAME as COLLABORATORS_SHARED_NAME } from '@ttn-lw/lib/store/actions/collaborators'

import { SHARED_NAME as API_KEYS_SHARED_NAME } from '@console/store/actions/api-keys'
import { SHARED_NAME as PACKET_BROKER_NETWORKS_SHARED_NAME } from '@console/store/actions/packet-broker'

import user from './user'
import logout from './logout'
import users from './users'
import applications from './applications'
import devices from './devices'
import gateways from './gateways'
import configuration from './configuration'
import apiKeys from './api-keys'
import createNamedRightsReducer from './rights'
import createNamedEventsReducer from './events'
import link from './link'
import webhooks from './webhooks'
import webhookFormats from './webhook-formats'
import webhookTemplates from './webhook-templates'
import pubsubs from './pubsubs'
import pubsubFormats from './pubsub-formats'
import applicationPackages from './application-packages'
import deviceTemplateFormats from './device-template-formats'
import organizations from './organizations'
import js from './join-server'
import gatewayStatus from './gateway-status'
import is from './identity-server'
import as from './application-server'
import deviceRepository from './device-repository'
import packetBroker from './packet-broker'
import ns from './network-server'
import notifications from './notifications'
import userPreferences from './user-preferences'
import recencyFrequencyItems from './recency-frequency-items'
import topEntities from './top-entities'
import search from './search'
import sessions from './sessions'
import authorizations from './authorizations'
import clients from './clients'
import connectionProfiles from './connection-profiles'

export default combineReducers({
  user,
  logout,
  users,
  init,
  status,
  collaborators,
  applications,
  link,
  devices,
  gateways,
  webhooks,
  webhookFormats,
  webhookTemplates,
  deviceTemplateFormats,
  pubsubs,
  pubsubFormats,
  applicationPackages,
  configuration,
  organizations,
  apiKeys,
  rights: combineReducers({
    applications: createNamedRightsReducer(APPLICATION),
    gateways: createNamedRightsReducer(GATEWAY),
    organizations: createNamedRightsReducer(ORGANIZATION),
    users: createNamedRightsReducer(USER),
    clients: createNamedRightsReducer(CLIENT),
  }),
  events: combineReducers({
    applications: createNamedEventsReducer(APPLICATION),
    devices: createNamedEventsReducer(END_DEVICE),
    gateways: createNamedEventsReducer(GATEWAY),
    organizations: createNamedEventsReducer(ORGANIZATION),
  }),
  ui: combineReducers({
    fetching,
    error,
  }),
  pagination: combineReducers({
    applications: createNamedPaginationReducer(APPLICATION, getApplicationId),
    apiKeys: createNamedPaginationReducer(API_KEYS_SHARED_NAME, getApiKeyId),
    collaborators: createNamedPaginationReducer(COLLABORATORS_SHARED_NAME, getCollaboratorId),
    devices: createNamedPaginationReducer(END_DEVICE, getCombinedDeviceId),
    gateways: createNamedPaginationReducer(GATEWAY, getGatewayId),
    organizations: createNamedPaginationReducer(ORGANIZATION, getOrganizationId),
    users: createNamedPaginationReducer(USER, getUserId),
    packetBrokerNetworks: createNamedPaginationReducer(
      PACKET_BROKER_NETWORKS_SHARED_NAME,
      getPacketBrokerNetworkId,
    ),
    clients: createNamedPaginationReducer('CLIENTS', getClientId),
    pubsubs: createNamedPaginationReducer('PUBSUB', getApplicationId),
    webhooks: createNamedPaginationReducer('WEBHOOKS', getApplicationId),
  }),
  js,
  gatewayStatus,
  is,
  as,
  deviceRepository,
  packetBroker,
  ns,
  searchAccounts,
  notifications,
  userPreferences,
  search,
  recencyFrequencyItems,
  topEntities,
  sessions,
  authorizations,
  clients,
  connectionProfiles,
})
