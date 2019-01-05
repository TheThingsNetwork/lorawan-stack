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

import fakeData from './fake-data'

export default {
  applications: {
    list (params) {
      const start = (params.page - 1) * params.pageSize
      const end = start + params.pageSize

      const res = fakeData.applications.filter(app => app.application_id)
      const total = res.length

      return new Promise(resolve => setTimeout(() => resolve(
        { applications: res.slice(start, end), totalCount: total }
      ), 1000))
    },
    search (params) {
      const start = (params.page - 1) * params.pageSize
      const end = start + params.pageSize
      const query = params.query || ''

      const res = fakeData.applications.filter(app => app.application_id.includes(query))
      const total = res.length

      return new Promise(resolve => setTimeout(() => resolve(
        { applications: res.slice(start, end), totalCount: total }
      ), 1000))
    },
  },
  application: {
    get (id) {
      const app = fakeData.applications.find(a => a.application_id === id)

      return new Promise((resolve, reject) => setTimeout(function () {
        if (app) {
          resolve(app)
        } else {
          reject(new Error())
        }
      }, 750))
    },
  },
  devices: {
    list (appId, params) {
      const start = (params.page - 1) * params.pageSize
      const end = start + params.pageSize

      const res = fakeData.devices.filter(d => d.application_id === appId)
      const total = res.length

      return new Promise(resolve => setTimeout(() => resolve(
        { devices: res.slice(start, end), totalCount: total }
      ), 1000))
    },
    search (appId, params) {
      const start = (params.page - 1) * params.pageSize
      const end = start + params.pageSize
      const query = params.query || ''

      const res = fakeData.devices
        .filter(d => d.application_id === appId)
        .filter(d => d.device_id.includes(query))
      const total = res.length

      return new Promise(resolve => setTimeout(() => resolve(
        { devices: res.slice(start, end), totalCount: total }
      ), 1000))
    },
  },
}
