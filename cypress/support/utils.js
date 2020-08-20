// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

/**
 * Disables Identity Server in the stack configuration object.
 *
 * @param {object} config - The stack configuration object.
 */
const disableIdentityServer = config => {
  Cypress._.merge(config, {
    APP_CONFIG: { stack_config: { is: { enabled: false, base_url: '' } } },
  })
}

/**
 * Disables Network Server in the stack configuration object.
 *
 * @param {object} config - The stack configuration object.
 */
const disableNetworkServer = config => {
  Cypress._.merge(config, {
    APP_CONFIG: { stack_config: { ns: { enabled: false, base_url: '' } } },
  })
}

/**
 * Disables Application Server in the stack configuration object.
 *
 * @param {object} config - The stack configuration object.
 */
const disableApplicationServer = config => {
  Cypress._.merge(config, {
    APP_CONFIG: { stack_config: { as: { enabled: false, base_url: '' } } },
  })
}

/**
 * Disables Join Server in the stack configuration object.
 *
 * @param {object} config - The stack configuration object.
 */
const disableJoinServer = config => {
  Cypress._.merge(config, {
    APP_CONFIG: { stack_config: { js: { enabled: false, base_url: '' } } },
  })
}

/**
 * Disables Gateway Server in the stack configuration object.
 *
 * @param {object} config - The stack configuration object.
 */
const disableGatewayServer = config => {
  Cypress._.merge(config, {
    APP_CONFIG: { stack_config: { gs: { enabled: false, base_url: '' } } },
  })
}

export {
  disableIdentityServer,
  disableNetworkServer,
  disableApplicationServer,
  disableJoinServer,
  disableGatewayServer,
}
