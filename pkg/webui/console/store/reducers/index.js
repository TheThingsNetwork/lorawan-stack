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
import {
  SHARED_NAME_SINGLE as APPLICATION_SHARED_NAME,
  SHARED_NAME as APPLICATIONS_SHARED_NAME,
} from '../actions/applications'
import { SHARED_NAME as GATEWAY_SHARED_NAME } from '../actions/gateways'

import { SHARED_NAME as DEVICE_SHARED_NAME } from '../actions/device'
import { getApplicationId, getGatewayId } from '../../../lib/selectors/id'
import user from './user'
import init from './init'
import applications from './applications'
import devices from './devices'
import device from './device'
import gateways from './gateways'
import configuration from './configuration'
import createNamedApiKeysReducer from './api-keys'
import createNamedRightsReducer from './rights'
import createNamedCollaboratorsReducer from './collaborators'
import createNamedEventsReducer from './events'
import createNamedApiKeyReducer from './api-key'
import link from './link'
import fetching from './ui/fetching'
import error from './ui/error'
import webhook from './webhook'
import webhooks from './webhooks'
import webhookFormats from './webhook-formats'
import { createNamedPaginationReducer } from './pagination'

export default combineReducers({
  user,
  init,
  applications,
  link,
  devices,
  device,
  gateways,
  webhook,
  webhooks,
  webhookFormats,
  configuration,
  apiKeys: combineReducers({
    application: createNamedApiKeyReducer(APPLICATION_SHARED_NAME),
    applications: createNamedApiKeysReducer(APPLICATION_SHARED_NAME),
    gateway: createNamedApiKeyReducer(GATEWAY_SHARED_NAME),
    gateways: createNamedApiKeysReducer(GATEWAY_SHARED_NAME),
  }),
  rights: combineReducers({
    applications: createNamedRightsReducer(APPLICATIONS_SHARED_NAME),
    gateways: createNamedRightsReducer(GATEWAY_SHARED_NAME),
  }),
  collaborators: combineReducers({
    applications: createNamedCollaboratorsReducer(APPLICATION_SHARED_NAME),
    gateways: createNamedCollaboratorsReducer(GATEWAY_SHARED_NAME),
  }),
  events: combineReducers({
    applications: createNamedEventsReducer(APPLICATION_SHARED_NAME),
    devices: createNamedEventsReducer(DEVICE_SHARED_NAME),
    gateways: createNamedEventsReducer(GATEWAY_SHARED_NAME),
  }),
  ui: combineReducers({
    fetching,
    error,
  }),
  pagination: combineReducers({
    applications: createNamedPaginationReducer(
      APPLICATION_SHARED_NAME,
      getApplicationId
    ),
    gateways: createNamedPaginationReducer(
      GATEWAY_SHARED_NAME,
      getGatewayId
    ),
  }),
})
