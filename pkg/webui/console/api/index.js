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
import { selectApplicationConfig, selectApplicationRootPath } from '../../lib/selectors/env'

const config = selectApplicationConfig()
const appRoot = selectApplicationRootPath()

const stack = {
  is: config.is.enabled ? config.is.base_url : undefined,
  gs: config.gs.enabled ? config.gs.base_url : undefined,
  ns: config.ns.enabled ? config.ns.base_url : undefined,
  as: config.as.enabled ? config.as.base_url : undefined,
  js: config.js.enabled ? config.js.base_url : undefined,
}

const isBaseUrl = config.is.base_url

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
    token() {
      return instance.get(`${appRoot}/api/auth/token`)
    },
    logout() {
      return instance.post(`${appRoot}/api/auth/logout`)
    },
  },
  clients: {
    get(client_id) {
      return instance.get(`${isBaseUrl}/is/clients/${client_id}`)
    },
  },
  users: {
    async get(userId) {
      return instance.get(`${isBaseUrl}/users/${userId}`, {
        headers: {
          Authorization: `Bearer ${(await token()).access_token}`,
        },
      })
    },
    async authInfo() {
      return instance.get(`${isBaseUrl}/auth_info`, {
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
    delete: ttnClient.Applications.deleteById.bind(ttnClient.Applications),
    create: ttnClient.Applications.create.bind(ttnClient.Applications),
    update: ttnClient.Applications.updateById.bind(ttnClient.Applications),
    eventsSubscribe: ttnClient.Applications.openStream.bind(ttnClient.Applications),
    apiKeys: {
      get: ttnClient.Applications.ApiKeys.getById.bind(ttnClient.Applications.ApiKeys),
      list: ttnClient.Applications.ApiKeys.getAll.bind(ttnClient.Applications.ApiKeys),
      update: ttnClient.Applications.ApiKeys.updateById.bind(ttnClient.Applications.ApiKeys),
      delete: ttnClient.Applications.ApiKeys.deleteById.bind(ttnClient.Applications.ApiKeys),
      create: ttnClient.Applications.ApiKeys.create.bind(ttnClient.Applications.ApiKeys),
    },
    link: {
      get: ttnClient.Applications.Link.get.bind(ttnClient.Applications.Link),
      set: ttnClient.Applications.Link.set.bind(ttnClient.Applications.Link),
      delete: ttnClient.Applications.Link.delete.bind(ttnClient.Applications.Link),
      stats: ttnClient.Applications.Link.getStats.bind(ttnClient.Applications.Link),
    },
    collaborators: {
      getOrganization: ttnClient.Applications.Collaborators.getByOrganizationId.bind(
        ttnClient.Applications.Collaborators,
      ),
      getUser: ttnClient.Applications.Collaborators.getByUserId.bind(
        ttnClient.Applications.Collaborators,
      ),
      list: ttnClient.Applications.Collaborators.getAll.bind(ttnClient.Applications.Collaborators),
      add: ttnClient.Applications.Collaborators.add.bind(ttnClient.Applications.Collaborators),
      update: ttnClient.Applications.Collaborators.update.bind(
        ttnClient.Applications.Collaborators,
      ),
      remove: ttnClient.Applications.Collaborators.remove.bind(
        ttnClient.Applications.Collaborators,
      ),
    },
    webhooks: {
      list: ttnClient.Applications.Webhooks.getAll.bind(ttnClient.Applications.Webhooks),
      get: ttnClient.Applications.Webhooks.getById.bind(ttnClient.Applications.Webhooks),
      create: ttnClient.Applications.Webhooks.create.bind(ttnClient.Applications.Webhooks),
      update: ttnClient.Applications.Webhooks.updateById.bind(ttnClient.Applications.Webhooks),
      delete: ttnClient.Applications.Webhooks.deleteById.bind(ttnClient.Applications.Webhooks),
      getFormats: ttnClient.Applications.Webhooks.getFormats.bind(ttnClient.Applications.Webhooks),
    },
  },
  devices: {
    list: ttnClient.Applications.Devices.getAll.bind(ttnClient.Applications.Devices),
  },
  device: {
    get: ttnClient.Applications.Devices.getById.bind(ttnClient.Applications.Devices),
    create: ttnClient.Applications.Devices.create.bind(ttnClient.Applications.Devices),
    update: ttnClient.Applications.Devices.updateById.bind(ttnClient.Applications.Devices),
    eventsSubscribe: ttnClient.Applications.Devices.openStream.bind(ttnClient.Applications.Devices),
    delete: ttnClient.Applications.Devices.deleteById.bind(ttnClient.Applications.Devices),
  },
  gateways: {
    list: ttnClient.Gateways.getAll.bind(ttnClient.Gateways),
  },
  gateway: {
    get: ttnClient.Gateways.getById.bind(ttnClient.Gateways),
    delete: ttnClient.Gateways.deleteById.bind(ttnClient.Gateways),
    create: ttnClient.Gateways.create.bind(ttnClient.Gateways),
    update: ttnClient.Gateways.updateById.bind(ttnClient.Gateways),
    stats: ttnClient.Gateways.getStatisticsById.bind(ttnClient.Gateways),
    eventsSubscribe: ttnClient.Gateways.openStream.bind(ttnClient.Gateways),
    collaborators: {
      getOrganization: ttnClient.Gateways.Collaborators.getByOrganizationId.bind(
        ttnClient.Gateways.Collaborators,
      ),
      getUser: ttnClient.Gateways.Collaborators.getByUserId.bind(ttnClient.Gateways.Collaborators),
      list: ttnClient.Gateways.Collaborators.getAll.bind(ttnClient.Gateways.Collaborators),
      add: ttnClient.Gateways.Collaborators.add.bind(ttnClient.Gateways.Collaborators),
      update: ttnClient.Gateways.Collaborators.update.bind(ttnClient.Gateways.Collaborators),
      remove: ttnClient.Gateways.Collaborators.remove.bind(ttnClient.Gateways.Collaborators),
    },
    apiKeys: {
      get: ttnClient.Gateways.ApiKeys.getById.bind(ttnClient.Gateways.ApiKeys),
      list: ttnClient.Gateways.ApiKeys.getAll.bind(ttnClient.Gateways.ApiKeys),
      update: ttnClient.Gateways.ApiKeys.updateById.bind(ttnClient.Gateways.ApiKeys),
      delete: ttnClient.Gateways.ApiKeys.deleteById.bind(ttnClient.Gateways.ApiKeys),
      create: ttnClient.Gateways.ApiKeys.create.bind(ttnClient.Gateways.ApiKeys),
    },
  },
  rights: {
    applications: ttnClient.Applications.getRightsById.bind(ttnClient.Applications),
    gateways: ttnClient.Gateways.getRightsById.bind(ttnClient.Gateways),
  },
  configuration: {
    listNsFrequencyPlans: ttnClient.Configuration.listNsFrequencyPlans.bind(
      ttnClient.Configuration,
    ),
    listGsFrequencyPlans: ttnClient.Configuration.listGsFrequencyPlans.bind(
      ttnClient.Configuration,
    ),
  },
  js: {
    joinEUIPRefixes: {
      list: ttnClient.Js.listJoinEUIPrefixes.bind(ttnClient.Js),
    },
  },
  ns: {
    generateDevAddress: ttnClient.Ns.generateDevAddress.bind(ttnClient.Ns),
  },
  organizations: {
    list: ttnClient.Organizations.getAll.bind(ttnClient.Organizations),
    create: ttnClient.Organizations.create.bind(ttnClient.Organizations),
  },
  organization: {
    get: ttnClient.Organizations.getById.bind(ttnClient.Organizations),
    eventsSubscribe: ttnClient.Organizations.openStream.bind(ttnClient.Organizations),
    delete: ttnClient.Organizations.deleteById.bind(ttnClient.Organizations),
    update: ttnClient.Organizations.updateById.bind(ttnClient.Organizations),
  },
}
