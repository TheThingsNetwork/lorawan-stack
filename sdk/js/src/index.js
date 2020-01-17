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

import Applications from './service/applications'
import Configuration from './service/configuration'
import Application from './entity/application'
import Api from './api'
import Token from './util/token'
import Gateways from './service/gateways'
import Js from './service/join-server'
import Ns from './service/network-server'
import Organizations from './service/organizations'
import Users from './service/users'
import Auth from './service/auth'
import EventHandler from './util/events'
import StackConfiguration from './util/stack-components'

class TtnLw {
  constructor(token, { stackConfig, connectionType, defaultUserId, proxy, axiosConfig }) {
    const tokenInstance = new Token(token)
    const stackConfiguration = new StackConfiguration(stackConfig)

    this.config = arguments.config
    this.api = new Api(connectionType, stackConfiguration, axiosConfig, tokenInstance.get())

    this.Applications = new Applications(this.api, {
      defaultUserId,
      proxy,
      stackConfig: stackConfiguration,
    })
    this.Application = Application.bind(null, this.Applications)
    this.Configuration = new Configuration(this.api.Configuration)
    this.Gateways = new Gateways(this.api, {
      defaultUserId,
      proxy,
      stackConfig: stackConfiguration,
    })
    this.Js = new Js(this.api.Js)
    this.Ns = new Ns(this.api.Ns)
    this.Organizations = new Organizations(this.api)
    this.Users = new Users(this.api)
    this.Auth = new Auth(this.api.EntityAccess)

    this.subscribe = EventHandler.subscribe
    this.unsubscribe = EventHandler.unsubscribe
  }
}

export default TtnLw
