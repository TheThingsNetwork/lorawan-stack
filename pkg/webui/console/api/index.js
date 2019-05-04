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

import axios from 'axios'
import TTN from 'ttn-lw'

import token from '../lib/access-token'
import getCookieValue from '../../lib/cookie'

const config = window.APP_CONFIG
const stack = {
  is: config.is.enabled ? config.is.base_url : undefined,
  gs: config.gs.enabled ? config.gs.base_url : undefined,
  ns: config.ns.enabled ? config.ns.base_url : undefined,
  as: config.as.enabled ? config.as.base_url : undefined,
  js: config.js.enabled ? config.js.base_url : undefined,
}

const ttnClient = new TTN(token, {
  stackConfig: stack,
  connectionType: 'http',
  proxy: false,
})

const csrf = getCookieValue('_csrf')
const instance = axios.create({
  headers: { 'X-CSRF-Token': csrf },
})

export default {
  console: {
    token () {
      return instance.get('/console/api/auth/token')
    },
    logout () {
      return instance.post('/console/api/auth/logout')
    },
  },
  clients: {
    get (client_id) {
      return instance.get(`/api/v3/is/clients/${client_id}`)
    },
  },
  users: {
    async get (userId) {
      return instance.get(`/api/v3/users/${userId}`, {
        headers: {
          Authorization: `Bearer ${(await token()).access_token}`,
        },
      })
    },
    async authInfo () {
      return instance.get('/api/v3/auth_info', {
        headers: {
          Authorization: `Bearer ${(await token()).access_token}`,
        },
      })
    },
  },
  applications: {
    list: ttnClient.Applications.getAll.bind(ttnClient.Applications),
    search: ttnClient.Applications.search.bind(ttnClient.Applications),
  },
  application: {
    get: ttnClient.Applications.getById.bind(ttnClient.Applications),
    'delete': ttnClient.Applications.deleteById.bind(ttnClient.Applications),
    create: ttnClient.Applications.create.bind(ttnClient.Applications),
    update: ttnClient.Applications.updateById.bind(ttnClient.Applications),
    eventsSubscribe: ttnClient.Applications.openStream.bind(ttnClient.Applications),
    apiKeys: {
      list: ttnClient.Applications.ApiKeys.getAll.bind(ttnClient.Applications.ApiKeys),
      update: ttnClient.Applications.ApiKeys.updateById.bind(ttnClient.Applications.ApiKeys),
      'delete': ttnClient.Applications.ApiKeys.deleteById.bind(ttnClient.Applications.ApiKeys),
      create: ttnClient.Applications.ApiKeys.create.bind(ttnClient.Applications.ApiKeys),
    },
    link: {
      get: ttnClient.Applications.Link.get.bind(ttnClient.Applications.Link),
      set: ttnClient.Applications.Link.set.bind(ttnClient.Applications.Link),
      'delete': ttnClient.Applications.Link.delete.bind(ttnClient.Applications.Link),
      stats: ttnClient.Applications.Link.getStats.bind(ttnClient.Applications.Link),
    },
    collaborators: {
      list: ttnClient.Applications.Collaborators.getAll.bind(ttnClient.Applications.Collaborators),
      add: ttnClient.Applications.Collaborators.add.bind(ttnClient.Applications.Collaborators),
      update: ttnClient.Applications.Collaborators.update.bind(ttnClient.Applications.Collaborators),
      remove: ttnClient.Applications.Collaborators.remove.bind(ttnClient.Applications.Collaborators),
    },
  },
  devices: {
    list: ttnClient.Applications.Devices.getAll.bind(ttnClient.Applications.Devices),
  },
  device: {
    get: ttnClient.Applications.Devices.getById.bind(ttnClient.Applications.Devices),
    create: ttnClient.Applications.Devices.create.bind(ttnClient.Applications.Devices),
    update: ttnClient.Applications.Devices.updateById.bind(ttnClient.Applications.Devices),
  },
  gateways: {
    list: ttnClient.Gateways.getAll.bind(ttnClient.Gateways),
  },
  gateway: {
    get: ttnClient.Gateways.getById.bind(ttnClient.Gateways),
    'delete': ttnClient.Gateways.deleteById.bind(ttnClient.Gateways),
    create: ttnClient.Gateways.create.bind(ttnClient.Gateways),
    update: ttnClient.Gateways.updateById.bind(ttnClient.Gateways),
    stats: ttnClient.Gateways.getStatisticsById.bind(ttnClient.Gateways),
  },
  rights: {
    applications: ttnClient.Applications.getRightsById.bind(ttnClient.Applications),
  },
  configuration: {
    listNsFrequencyPlans: ttnClient.Configuration.listNsFrequencyPlans.bind(ttnClient.Configuration),
    listGsFrequencyPlans: ttnClient.Configuration.listGsFrequencyPlans.bind(ttnClient.Configuration),
  },
}
