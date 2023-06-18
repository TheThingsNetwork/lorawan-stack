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

describe('Overview', () => {
  const userId = 'main-overview-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'view-overview-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  describe('Application', () => {
    const applicationId = 'app-overview-test'
    const application = {
      ids: { application_id: applicationId },
      name: 'Application Test Name',
      description: `Application Test Description`,
    }

    before(() => {
      cy.createApplication(application, userId)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    })

    it('displays UI elements in place', () => {
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}`)

      cy.findByText(`${application.name}`, { selector: 'h1' }).should('be.visible')
      cy.findByRole('link', { name: /0 End device/ }).should('be.visible')
      cy.findByRole('link', { name: /1 Collaborator/ }).should('be.visible')
      cy.findByRole('link', { name: /0 API key/ }).should('be.visible')

      cy.findByRole('row', { name: new RegExp(applicationId) }).should('be.visible')

      cy.findByTestId('events-widget').should('be.visible')
      cy.findByTestId('devices-table').should('be.visible')

      cy.findByTestId('error-notification').should('not.exist')
    })
  })

  describe('Gateway', () => {
    const gatewayId = 'gateway-overview-test'
    const gateway = {
      ids: { gateway_id: gatewayId, eui: '0000000000000000' },
      name: 'Gateway Test Name',
      description: 'Gateway Test Description',
      gateway_server_address: 'test-address',
      frequency_plan_id: 'EU_863_870',
    }

    before(() => {
      cy.createGateway(gateway, userId)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    })

    it('displays UI elements in place', () => {
      cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}`)

      cy.findByText(`${gateway.name}`, { selector: 'h1' }).should('be.visible')
      cy.findByRole('link', { name: /1 Collaborator/ }).should('be.visible')
      cy.findByRole('link', { name: /0 API key/ }).should('be.visible')

      cy.findByRole('row', { name: new RegExp(gatewayId) }).should('be.visible')
      cy.findByRole('row', { name: new RegExp(gateway.ids.eui) }).should('be.visible')
      cy.findByText(new RegExp(gateway.description)).should('be.visible')
      cy.findByText(new RegExp(gateway.gateway_server_address)).should('be.visible')
      cy.findByText(new RegExp(gateway.frequency_plan_id)).should('be.visible')

      cy.findByTestId('events-widget').should('be.visible')
      cy.findByTestId('map-widget').should('be.visible')

      cy.findByTestId('error-notification').should('not.exist')
    })
  })

  describe('Organization', () => {
    const organizationId = 'org-overview-test'
    const organization = {
      ids: { organization_id: organizationId },
      name: 'Organization Test Name',
    }

    before(() => {
      cy.createOrganization(organization, userId)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    })

    it('displays UI elements in place', () => {
      cy.visit(`${Cypress.config('consoleRootPath')}/organizations/${organizationId}`)

      cy.findByText(`${organization.name}`, { selector: 'h1' }).should('be.visible')
      cy.findByRole('link', { name: /1 Collaborator/ }).should('be.visible')
      cy.findByRole('link', { name: /0 API key/ }).should('be.visible')

      cy.findByRole('row', { name: new RegExp(organizationId) }).should('be.visible')

      cy.findByTestId('events-widget').should('be.visible')

      cy.findByTestId('error-notification').should('not.exist')
    })
  })

  describe('End Devices', () => {
    const applicationId = 'app-end-devices-overview-test'
    const application = {
      ids: { application_id: applicationId },
      name: 'Application End Devices Test Name',
      description: `Application End Devices Test Description`,
    }
    let endDevice

    before(() => {
      cy.createApplication(application, userId)
      cy.createMockDeviceAllComponents(applicationId).then(body => {
        endDevice = body.end_device
      })
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    })

    it('displays UI elements in place', () => {
      cy.visit(
        `${Cypress.config('consoleRootPath')}/applications/${applicationId}/devices/${
          endDevice.ids.device_id
        }`,
      )

      cy.findByText(`${endDevice.name}`, { selector: 'h1' }).should('be.visible')

      cy.findByRole('row', { name: new RegExp(endDevice.ids.device_id) }).should('be.visible')
      cy.findByText(new RegExp(endDevice.description)).should('be.visible')
      cy.findByRole('row', { name: new RegExp(endDevice.ids.dev_eui) }).should('be.visible')
      cy.findByRole('row', { name: new RegExp(endDevice.ids.join_eui) }).should('be.visible')

      cy.findByTestId('events-widget').should('be.visible')
      cy.findByTestId('map-widget').should('be.visible')

      cy.findByTestId('error-notification').should('not.exist')
    })
  })
})
