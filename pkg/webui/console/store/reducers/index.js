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

import { combineReducers } from 'redux'
import { connectRouter } from 'connected-react-router'

import {
  SHARED_NAME_SINGLE as APPLICATION_SHARED_NAME,
  SHARED_NAME as APPLICATIONS_SHARED_NAME,
} from '../actions/applications'
import { SHARED_NAME as GATEWAY_SHARED_NAME } from '../actions/gateways'
import { SHARED_NAME as ORGANIZATION_SHARED_NAME } from '../actions/organizations'
import { SHARED_NAME as DEVICE_SHARED_NAME } from '../actions/devices'
import { SHARED_NAME as USER_SHARED_NAME } from '../actions/users'
import { SHARED_NAME as API_KEYS_SHARED_NAME } from '../actions/api-keys'
import { SHARED_NAME as COLLABORATORS_SHARED_NAME } from '../actions/collaborators'
import {
  getUserId,
  getApplicationId,
  getGatewayId,
  getOrganizationId,
  getCombinedDeviceId,
  getApiKeyId,
  getCollaboratorId,
} from '../../../lib/selectors/id'
import user from './user'
import users from './users'
import init from './init'
import applications from './applications'
import devices from './devices'
import gateways from './gateways'
import configuration from './configuration'
import apiKeys from './api-keys'
import createNamedRightsReducer from './rights'
import collaborators from './collaborators'
import createNamedEventsReducer from './events'
import link from './link'
import fetching from './ui/fetching'
import error from './ui/error'
import webhooks from './webhooks'
import webhookFormats from './webhook-formats'
import pubsubs from './pubsubs'
import pubsubFormats from './pubsub-formats'
import deviceTemplateFormats from './device-template-formats'
import organizations from './organizations'
import { createNamedPaginationReducer } from './pagination'
import js from './join-server'
import gatewayStatus from './gateway-status'

export default history =>
  combineReducers({
    user,
    users,
    init,
    applications,
    link,
    devices,
    gateways,
    webhooks,
    webhookFormats,
    deviceTemplateFormats,
    pubsubs,
    pubsubFormats,
    configuration,
    organizations,
    apiKeys,
    collaborators,
    rights: combineReducers({
      applications: createNamedRightsReducer(APPLICATIONS_SHARED_NAME),
      gateways: createNamedRightsReducer(GATEWAY_SHARED_NAME),
      organizations: createNamedRightsReducer(ORGANIZATION_SHARED_NAME),
    }),
    events: combineReducers({
      applications: createNamedEventsReducer(APPLICATION_SHARED_NAME),
      devices: createNamedEventsReducer(DEVICE_SHARED_NAME),
      gateways: createNamedEventsReducer(GATEWAY_SHARED_NAME),
      organizations: createNamedEventsReducer(ORGANIZATION_SHARED_NAME),
    }),
    ui: combineReducers({
      fetching,
      error,
    }),
    pagination: combineReducers({
      applications: createNamedPaginationReducer(APPLICATION_SHARED_NAME, getApplicationId),
      apiKeys: createNamedPaginationReducer(API_KEYS_SHARED_NAME, getApiKeyId),
      collaborators: createNamedPaginationReducer(COLLABORATORS_SHARED_NAME, getCollaboratorId),
      devices: createNamedPaginationReducer(DEVICE_SHARED_NAME, getCombinedDeviceId),
      gateways: createNamedPaginationReducer(GATEWAY_SHARED_NAME, getGatewayId),
      organizations: createNamedPaginationReducer(ORGANIZATION_SHARED_NAME, getOrganizationId),
      users: createNamedPaginationReducer(USER_SHARED_NAME, getUserId),
    }),
    router: connectRouter(history),
    js,
    gatewayStatus,
  })
