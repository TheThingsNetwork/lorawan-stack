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
import { SHARED_NAME as APPLICATION_SHARED_NAME } from '../actions/application'
import { SHARED_NAME as APPLICATIONS_SHARED_NAME } from '../actions/applications'
import user from './user'
import client from './client'
import init from './init'
import applications from './applications'
import application from './application'
import devices from './devices'
import device from './device'
import gateways from './gateways'
import configuration from './configuration'
import createNamedApiKeysReducer from './api-keys'
import createNamedRightsReducer from './rights'
import createNamedCollaboratorsReducer from './collaborators'

export default combineReducers({
  user,
  client,
  init,
  applications,
  application,
  devices,
  device,
  gateways,
  configuration,
  apiKeys: combineReducers({
    applications: createNamedApiKeysReducer(APPLICATION_SHARED_NAME),
  }),
  rights: combineReducers({
    applications: createNamedRightsReducer(APPLICATIONS_SHARED_NAME),
  }),
  collaborators: combineReducers({
    applications: createNamedCollaboratorsReducer(APPLICATION_SHARED_NAME),
  }),
})
