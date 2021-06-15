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

import axios from 'axios'

import TTS, { STACK_COMPONENTS_MAP, AUTHORIZATION_MODES } from 'ttn-lw'

import toast from '@ttn-lw/components/toast'

import {
  selectStackConfig,
  selectApplicationRootPath,
  selectCSRFToken,
} from '@ttn-lw/lib/selectors/env'
import tokenCreator from '@ttn-lw/lib/access-token'

const stackConfig = selectStackConfig()
const appRoot = selectApplicationRootPath()

const stack = {
  [STACK_COMPONENTS_MAP.is]: stackConfig.is.enabled ? stackConfig.is.base_url : undefined,
  [STACK_COMPONENTS_MAP.gs]: stackConfig.gs.enabled ? stackConfig.gs.base_url : undefined,
  [STACK_COMPONENTS_MAP.ns]: stackConfig.ns.enabled ? stackConfig.ns.base_url : undefined,
  [STACK_COMPONENTS_MAP.as]: stackConfig.as.enabled ? stackConfig.as.base_url : undefined,
  [STACK_COMPONENTS_MAP.js]: stackConfig.js.enabled ? stackConfig.js.base_url : undefined,
  [STACK_COMPONENTS_MAP.edtc]: stackConfig.edtc.enabled ? stackConfig.edtc.base_url : undefined,
  [STACK_COMPONENTS_MAP.qrg]: stackConfig.qrg.enabled ? stackConfig.qrg.base_url : undefined,
  [STACK_COMPONENTS_MAP.gcs]: stackConfig.gcs.enabled ? stackConfig.gcs.base_url : undefined,
}

const isBaseUrl = stackConfig.is.base_url

const csrf = selectCSRFToken()
const instance = axios.create()

instance.interceptors.response.use(
  response => response,
  error => {
    if ('response' in error && error.response && 'data' in error.response) {
      throw error.response.data
    }
    throw error
  },
)

const token = tokenCreator(() => instance.get(`${appRoot}/api/auth/token`))

const tts = new TTS({
  authorization: {
    mode: AUTHORIZATION_MODES.KEY,
    key: token,
  },
  stackConfig: stack,
  connectionType: 'http',
  proxy: false,
  axiosConfig: {
    timeout: 10000,
  },
})

// Forward header warnings to the toast message queue.
tts.subscribe('warning', payload => {
  toast({
    title: 'Warning',
    type: toast.types.WARNING,
    message: payload,
    preventConsecutive: true,
  })
})

export default {
  console: {
    token: () => instance.get(`${appRoot}/api/auth/token`),
    logout: async () => {
      const headers = token => ({
        headers: { 'X-CSRF-Token': token },
      })
      try {
        return await axios.post(`${appRoot}/api/auth/logout`, undefined, headers(csrf))
      } catch (error) {
        if (
          error.response &&
          error.response.status === 403 &&
          typeof error.response.data === 'string' &&
          error.response.data.includes('CSRF')
        ) {
          // If the CSRF token is invalid, it likely means that the CSRF cookie
          // has been deleted or became outdated. Making a new request to the
          // current path can then retrieve a fresh CSRF cookie, with which
          // the logout can be retried.
          const csrfResult = await axios.get(window.location)
          const freshCsrf = csrfResult.headers['x-csrf-token']
          if (freshCsrf) {
            return axios.post(`${appRoot}/api/auth/logout`, undefined, headers(freshCsrf))
          }
        }

        throw error
      }
    },
  },
  clients: {
    get: client_id => instance.get(`${isBaseUrl}/is/clients/${client_id}`),
  },
  users: {
    create: tts.Users.create.bind(tts.Users),
    get: tts.Users.getById.bind(tts.Users),
    list: tts.Users.getAll.bind(tts.Users),
    update: tts.Users.updateById.bind(tts.Users),
    delete: tts.Users.deleteById.bind(tts.Users),
    search: tts.Users.search.bind(tts.Users),
    authInfo: tts.Auth.getAuthInfo.bind(tts.Auth),
    apiKeys: {
      get: tts.Users.ApiKeys.getById.bind(tts.Users.ApiKeys),
      list: tts.Users.ApiKeys.getAll.bind(tts.Users.ApiKeys),
      update: tts.Users.ApiKeys.updateById.bind(tts.Users.ApiKeys),
      delete: tts.Users.ApiKeys.deleteById.bind(tts.Users.ApiKeys),
      create: tts.Users.ApiKeys.create.bind(tts.Users.ApiKeys),
    },
  },
  applications: {
    list: tts.Applications.getAll.bind(tts.Applications),
    search: tts.Applications.search.bind(tts.Applications),
  },
  application: {
    get: tts.Applications.getById.bind(tts.Applications),
    delete: tts.Applications.deleteById.bind(tts.Applications),
    purge: tts.Applications.purgeById.bind(tts.Applications),
    create: tts.Applications.create.bind(tts.Applications),
    update: tts.Applications.updateById.bind(tts.Applications),
    eventsSubscribe: tts.Applications.openStream.bind(tts.Applications),
    getMqttConnectionInfo: tts.Applications.getMqttConnectionInfo.bind(tts.Applications),
    apiKeys: {
      get: tts.Applications.ApiKeys.getById.bind(tts.Applications.ApiKeys),
      list: tts.Applications.ApiKeys.getAll.bind(tts.Applications.ApiKeys),
      update: tts.Applications.ApiKeys.updateById.bind(tts.Applications.ApiKeys),
      delete: tts.Applications.ApiKeys.deleteById.bind(tts.Applications.ApiKeys),
      create: tts.Applications.ApiKeys.create.bind(tts.Applications.ApiKeys),
    },
    link: {
      get: tts.Applications.Link.get.bind(tts.Applications.Link),
      set: tts.Applications.Link.set.bind(tts.Applications.Link),
    },
    collaborators: {
      getOrganization: tts.Applications.Collaborators.getByOrganizationId.bind(
        tts.Applications.Collaborators,
      ),
      getUser: tts.Applications.Collaborators.getByUserId.bind(tts.Applications.Collaborators),
      list: tts.Applications.Collaborators.getAll.bind(tts.Applications.Collaborators),
      add: tts.Applications.Collaborators.add.bind(tts.Applications.Collaborators),
      update: tts.Applications.Collaborators.update.bind(tts.Applications.Collaborators),
      remove: tts.Applications.Collaborators.remove.bind(tts.Applications.Collaborators),
    },
    webhooks: {
      list: tts.Applications.Webhooks.getAll.bind(tts.Applications.Webhooks),
      get: tts.Applications.Webhooks.getById.bind(tts.Applications.Webhooks),
      create: tts.Applications.Webhooks.create.bind(tts.Applications.Webhooks),
      update: tts.Applications.Webhooks.updateById.bind(tts.Applications.Webhooks),
      delete: tts.Applications.Webhooks.deleteById.bind(tts.Applications.Webhooks),
      getFormats: tts.Applications.Webhooks.getFormats.bind(tts.Applications.Webhooks),
      listTemplates: tts.Applications.Webhooks.listTemplates.bind(tts.Applications.Webhooks),
      getTemplate: tts.Applications.Webhooks.getTemplate.bind(tts.Applications.Webhooks),
    },
    pubsubs: {
      list: tts.Applications.PubSubs.getAll.bind(tts.Applications.PubSubs),
      get: tts.Applications.PubSubs.getById.bind(tts.Applications.PubSubs),
      create: tts.Applications.PubSubs.create.bind(tts.Applications.PubSubs),
      update: tts.Applications.PubSubs.updateById.bind(tts.Applications.PubSubs),
      delete: tts.Applications.PubSubs.deleteById.bind(tts.Applications.PubSubs),
      getFormats: tts.Applications.PubSubs.getFormats.bind(tts.Applications.PubSubs),
    },
    packages: {
      getDefaultAssociation: tts.Applications.Packages.getDefaultAssociation.bind(
        tts.Applications.Packages,
      ),
      setDefaultAssociation: tts.Applications.Packages.setDefaultAssociation.bind(
        tts.Applications.Packages,
      ),
      deleteDefaultAssociation: tts.Applications.Packages.deleteDefaultAssociation.bind(
        tts.Applications.Packages,
      ),
    },
  },
  devices: {
    list: tts.Applications.Devices.getAll.bind(tts.Applications.Devices),
    search: tts.Applications.Devices.search.bind(tts.Applications.Devices),
  },
  device: {
    get: tts.Applications.Devices.getById.bind(tts.Applications.Devices),
    create: tts.Applications.Devices.create.bind(tts.Applications.Devices),
    bulkCreate: tts.Applications.Devices.bulkCreate.bind(tts.Applications.Devices),
    update: tts.Applications.Devices.updateById.bind(tts.Applications.Devices),
    eventsSubscribe: tts.Applications.Devices.openStream.bind(tts.Applications.Devices),
    delete: tts.Applications.Devices.deleteById.bind(tts.Applications.Devices),
    simulateUplink: tts.Applications.Devices.simulateUplink.bind(tts.Applications.Devices),
  },
  deviceTemplates: {
    listFormats: tts.Applications.Devices.listTemplateFormats.bind(tts.Applications.Devices),
    convert: tts.Applications.Devices.convertTemplate.bind(tts.Applications.Devices),
  },
  deviceRepository: {
    listBrands: tts.Applications.Devices.Repository.listBrands.bind(
      tts.Applications.Devices.Repository,
    ),
    getBrand: tts.Applications.Devices.Repository.getBrand.bind(
      tts.Applications.Devices.Repository,
    ),
    listModels: tts.Applications.Devices.Repository.listModels.bind(
      tts.Applications.Devices.Repository,
    ),
    getModel: tts.Applications.Devices.Repository.getModel.bind(
      tts.Applications.Devices.Repository,
    ),
    getTemplate: tts.Applications.Devices.Repository.getTemplate.bind(
      tts.Applications.Devices.Repository,
    ),
  },
  gateways: {
    list: tts.Gateways.getAll.bind(tts.Gateways),
    search: tts.Gateways.search.bind(tts.Gateways),
  },
  downlinkQueue: {
    list: tts.Applications.Devices.DownlinkQueue.list.bind(tts.Applications.Devices.DownlinkQueue),
    replace: tts.Applications.Devices.DownlinkQueue.replace.bind(
      tts.Applications.Devices.DownlinkQueue,
    ),
    push: tts.Applications.Devices.DownlinkQueue.push.bind(tts.Applications.Devices.DownlinkQueue),
  },
  gateway: {
    get: tts.Gateways.getById.bind(tts.Gateways),
    getGlobalConf: tts.Gateways.getGlobalConf.bind(tts.Gateways),
    delete: tts.Gateways.deleteById.bind(tts.Gateways),
    purge: tts.Gateways.purgeById.bind(tts.Gateways),
    create: tts.Gateways.create.bind(tts.Gateways),
    update: tts.Gateways.updateById.bind(tts.Gateways),
    stats: tts.Gateways.getStatisticsById.bind(tts.Gateways),
    eventsSubscribe: tts.Gateways.openStream.bind(tts.Gateways),
    collaborators: {
      getOrganization: tts.Gateways.Collaborators.getByOrganizationId.bind(
        tts.Gateways.Collaborators,
      ),
      getUser: tts.Gateways.Collaborators.getByUserId.bind(tts.Gateways.Collaborators),
      list: tts.Gateways.Collaborators.getAll.bind(tts.Gateways.Collaborators),
      add: tts.Gateways.Collaborators.add.bind(tts.Gateways.Collaborators),
      update: tts.Gateways.Collaborators.update.bind(tts.Gateways.Collaborators),
      remove: tts.Gateways.Collaborators.remove.bind(tts.Gateways.Collaborators),
    },
    apiKeys: {
      get: tts.Gateways.ApiKeys.getById.bind(tts.Gateways.ApiKeys),
      list: tts.Gateways.ApiKeys.getAll.bind(tts.Gateways.ApiKeys),
      update: tts.Gateways.ApiKeys.updateById.bind(tts.Gateways.ApiKeys),
      delete: tts.Gateways.ApiKeys.deleteById.bind(tts.Gateways.ApiKeys),
      create: tts.Gateways.ApiKeys.create.bind(tts.Gateways.ApiKeys),
    },
  },
  rights: {
    applications: tts.Applications.getRightsById.bind(tts.Applications),
    gateways: tts.Gateways.getRightsById.bind(tts.Gateways),
    organizations: tts.Organizations.getRightsById.bind(tts.Organizations),
    users: tts.Users.getRightsById.bind(tts.Users),
  },
  configuration: {
    listNsFrequencyPlans: tts.Configuration.listNsFrequencyPlans.bind(tts.Configuration),
    listGsFrequencyPlans: tts.Configuration.listGsFrequencyPlans.bind(tts.Configuration),
  },
  js: {
    joinEUIPrefixes: {
      list: tts.Js.listJoinEUIPrefixes.bind(tts.Js),
    },
  },
  ns: {
    generateDevAddress: tts.Ns.generateDevAddress.bind(tts.Ns),
  },
  is: {
    getConfiguration: tts.Is.getConfiguration.bind(tts.Is),
  },
  as: {
    decodeUplink: tts.As.decodeUplink.bind(tts.As),
    encodeDownlink: tts.As.encodeDownlink.bind(tts.As),
  },
  organizations: {
    list: tts.Organizations.getAll.bind(tts.Organizations),
    search: tts.Organizations.search.bind(tts.Organizations),
    create: tts.Organizations.create.bind(tts.Organizations),
  },
  organization: {
    get: tts.Organizations.getById.bind(tts.Organizations),
    eventsSubscribe: tts.Organizations.openStream.bind(tts.Organizations),
    delete: tts.Organizations.deleteById.bind(tts.Organizations),
    purge: tts.Organizations.purgeById.bind(tts.Organizations),
    update: tts.Organizations.updateById.bind(tts.Organizations),
    apiKeys: {
      get: tts.Organizations.ApiKeys.getById.bind(tts.Organizations.ApiKeys),
      list: tts.Organizations.ApiKeys.getAll.bind(tts.Organizations.ApiKeys),
      update: tts.Organizations.ApiKeys.updateById.bind(tts.Organizations.ApiKeys),
      delete: tts.Organizations.ApiKeys.deleteById.bind(tts.Organizations.ApiKeys),
      create: tts.Organizations.ApiKeys.create.bind(tts.Organizations.ApiKeys),
    },
    collaborators: {
      getOrganization: tts.Organizations.Collaborators.getByOrganizationId.bind(
        tts.Organizations.Collaborators,
      ),
      getUser: tts.Organizations.Collaborators.getByUserId.bind(tts.Organizations.Collaborators),
      list: tts.Organizations.Collaborators.getAll.bind(tts.Organizations.Collaborators),
      add: tts.Organizations.Collaborators.add.bind(tts.Organizations.Collaborators),
      update: tts.Organizations.Collaborators.update.bind(tts.Organizations.Collaborators),
      remove: tts.Organizations.Collaborators.remove.bind(tts.Organizations.Collaborators),
    },
  },
}
