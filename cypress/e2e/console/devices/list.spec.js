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

describe('End Device list', () => {
  const appId = 'device-list-test-application'
  const application = { ids: { application_id: appId } }
  const userId = 'device-list-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'end-device-edit-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const endDevices = [
    {
      application_server_address: window.location.hostname,
      ids: {
        device_id: 'some-test-end-device',
        dev_eui: '0000000000000001',
        join_eui: '0000000000000000',
      },
      name: 'End Device Test Name',
      description: 'End Device Test Description',
      join_server_address: window.location.hostname,
      network_server_address: window.location.hostname,
    },
    {
      application_server_address: window.location.hostname,
      ids: {
        device_id: 'other-test-end-device',
        dev_eui: '0000000000000002',
        join_eui: '0000000000000000',
      },
      name: 'End Device Test Name',
      description: 'End Device Test Description',
      join_server_address: window.location.hostname,
      network_server_address: window.location.hostname,
    },
    {
      application_server_address: window.location.hostname,
      ids: {
        device_id: 'third-test-end-device',
        dev_eui: '0000000000000003',
        join_eui: '0000000000000000',
      },
      name: 'End Device Test Name',
      description: 'End Device Test Description',
      join_server_address: window.location.hostname,
      network_server_address: window.location.hostname,
    },
  ]

  const endDeviceFieldMask = {
    paths: [
      'join_server_address',
      'network_server_address',
      'application_server_address',
      'ids.dev_eui',
      'ids.join_eui',
      'name',
      'description',
    ],
  }

  const endDeviceRequestBody = [
    {
      end_device: endDevices[0],
      field_mask: endDeviceFieldMask,
    },
    {
      end_device: endDevices[1],
      field_mask: endDeviceFieldMask,
    },
    {
      end_device: endDevices[2],
      field_mask: endDeviceFieldMask,
    },
  ]

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
    cy.createEndDeviceIsOnly(appId, endDeviceRequestBody[0])
    cy.createEndDeviceIsOnly(appId, endDeviceRequestBody[1])
    cy.createEndDeviceIsOnly(appId, endDeviceRequestBody[2])
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices`)
  })

  it('succeeds searching by end device id', () => {
    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 3)
    })
    cy.findByRole('cell', { name: endDevices[0].ids.device_id }).should('be.visible')
    cy.findByRole('cell', { name: endDevices[1].ids.device_id }).should('be.visible')
    cy.findByRole('cell', { name: endDevices[2].ids.device_id }).should('be.visible')

    cy.findByTestId('search-input').as('searchInput')
    cy.get('@searchInput').type('some')

    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 1)
    })
    cy.findByRole('cell', { name: endDevices[0].ids.device_id }).should('be.visible')
    cy.findByRole('cell', { name: endDevices[1].ids.device_id }).should('not.exist')
    cy.findByRole('cell', { name: endDevices[2].ids.device_id }).should('not.exist')

    cy.get('@searchInput').clear()
    cy.get('@searchInput').type('other')

    cy.findByRole('rowgroup').within(() => {
      cy.findByRole('row').should('have.length', 1)
    })
    cy.findByRole('cell', { name: endDevices[0].ids.device_id }).should('not.exist')
    cy.findByRole('cell', { name: endDevices[1].ids.device_id }).should('be.visible')
    cy.findByRole('cell', { name: endDevices[2].ids.device_id }).should('not.exist')

    cy.get('@searchInput').clear()
    cy.get('@searchInput').type('third')

    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 1)
    })
    cy.findByRole('cell', { name: endDevices[0].ids.device_id }).should('not.exist')
    cy.findByRole('cell', { name: endDevices[1].ids.device_id }).should('not.exist')
    cy.findByRole('cell', { name: endDevices[2].ids.device_id }).should('be.visible')

    cy.get('@searchInput').clear()
    cy.get('@searchInput').type('test-end-device')

    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 3)
    })
    cy.findByRole('cell', { name: endDevices[0].ids.device_id }).should('be.visible')
    cy.findByRole('cell', { name: endDevices[1].ids.device_id }).should('be.visible')
    cy.findByRole('cell', { name: endDevices[2].ids.device_id }).should('be.visible')
  })
})
