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

import crypto from 'crypto'

/** Stack configuration utitlities. */

/**
 * Disables Identity Server in the stack configuration object.
 *
 * @param {object} config - The stack configuration object.
 */
const disableIdentityServer = config => {
  Cypress._.merge(config, {
    APP_CONFIG: { stack_config: { is: { enabled: false, base_url: '' } } },
  })
  Cypress.config({
    isBaseUrl: '',
    isEnabled: false,
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
  Cypress.config({
    nsBaseUrl: '',
    nsEnabled: false,
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
  Cypress.config({
    asBaseUrl: '',
    asEnabled: false,
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
  Cypress.config({
    jsBaseUrl: '',
    jsEnabled: false,
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
  Cypress.config({
    gsBaseUrl: '',
    gsEnabled: false,
  })
}

/**
 * Generates a Join Server Only configuration object.
 *
 * @param {object} config - The stack configuration object.
 */
const generateJoinServerOnlyConfig = config => {
  Cypress._.merge(config, {
    APP_CONFIG: {
      stack_config: {
        gs: { enabled: false, base_url: '' },
        ns: { enabled: false, base_url: '' },
        as: { enabled: false, base_url: '' },
      },
    },
  })
  Cypress.config({
    asBaseUrl: '',
    asEnabled: false,
    gsBaseUrl: '',
    gsEnabled: false,
    nsBaseUrl: '',
    nsEnabled: false,
  })
}

/** General utitlies. */

/**
 * Generates pseudorandom hex value.
 *
 * @param {number} length -  The length of the resulting hex value.
 * @returns {string} - Pseudorandom hex value of length `length`.
 */
const generateHexValue = length => crypto.randomBytes(Math.floor(length / 2)).toString('hex')

const generateCollaborator = (entity, type) => {
  const collabUserId = 'test-collab-user'
  const organizationId = 'test-collab-org'

  const applicationRights = ['RIGHT_APPLICATION_ALL']
  const gatewayRights = ['RIGHT_GATEWAY_ALL']
  const organizationRights = [
    'RIGHT_APPLICATION_ALL',
    'RIGHT_CLIENT_ALL',
    'RIGHT_GATEWAY_ALL',
    'RIGHT_ORGANIZATION_ALL',
  ]

  const userIds = {
    user_ids: {
      user_id: collabUserId,
    },
  }

  const orgIds = {
    organization_ids: {
      organization_id: organizationId,
    },
  }

  return {
    collaborator: {
      ids: type === 'user' ? userIds : orgIds,
      rights:
        entity === 'applications'
          ? applicationRights
          : entity === 'gateways'
            ? gatewayRights
            : organizationRights,
    },
  }
}

export {
  disableIdentityServer,
  disableNetworkServer,
  disableApplicationServer,
  disableJoinServer,
  disableGatewayServer,
  generateHexValue,
  generateJoinServerOnlyConfig,
  generateCollaborator,
}
