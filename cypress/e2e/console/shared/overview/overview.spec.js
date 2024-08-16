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

      // Check header.
      cy.findByText(`${application.name}`, { selector: 'h5' }).should('be.visible')
      cy.findByText(/0 End devices/).should('be.visible')
      cy.findByText(/Created/).should('be.visible')

      cy.findByText(new RegExp(applicationId)).should('be.visible')

      // Check panels.
      cy.findByText('Top end devices').should('be.visible')
      cy.findByText('Network activity').should('be.visible')
      cy.findByText('Latest decoded payload').should('be.visible')
      cy.findByText('Device locations').should('be.visible')

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

      // Check header.
      cy.findByText(`${gateway.name}`, { selector: 'h5' }).should('be.visible')

      // Check general information panel.
      cy.get('#stage').within(() => {
        cy.findAllByText(gatewayId).should('be.visible').and('have.length', 2)
        cy.findByText('Europe 863-870 MHz (SF12 for RX2)').should('be.visible')
        cy.findByText(/No uplinks yet/) // Check for no uplinks.

        // Check panels.
        cy.findByText('General information').should('be.visible')
        cy.findByText('Gateway status').should('be.visible')
        cy.findByText('Network activity').should('exist').scrollIntoView()
        cy.findByText('Network activity').should('be.visible')
        cy.findByText('Location').should('be.visible')
      })

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

      cy.findByTestId('organization-header').within(() => {
        cy.findByText(`${organization.name}`, { selector: 'h5' }).should('be.visible')
        cy.findByText(/Members/).should('be.visible')
        cy.findByText(/API keys/).should('be.visible')
      })

      cy.findByText(new RegExp(organizationId)).should('be.visible')

      cy.findByRole('button', { name: 'Members' }).should('be.visible')
      cy.findByRole('button', { name: 'API keys' }).should('be.visible')
      cy.findByRole('button', { name: 'Settings' }).should('be.visible')

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

      cy.findByTestId('device-overview-header').within(() => {
        cy.findByText(`${endDevice.name}`, { selector: 'h5' }).should('be.visible')
        cy.findByText(new RegExp(endDevice.ids.device_id)).should('be.visible')
      })

      cy.findByText('End device info').should('be.visible')
      cy.findByText('General information').should('be.visible')
      cy.findByText('Latest decoded payload').should('be.visible')
      cy.findByText('Network activity').should('be.visible')
      cy.get('[data-test-id="overview-panel-Location"]').scrollIntoView()
      cy.findByTestId('overview-panel-Location').within(() => {
        cy.findByText('Location').should('be.visible')
      })

      cy.findByTestId('error-notification').should('not.exist')
    })
  })
})
