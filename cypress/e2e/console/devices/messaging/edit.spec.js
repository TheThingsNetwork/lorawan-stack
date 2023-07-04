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

describe('End device messaging', () => {
  const userId = 'main-overview-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'view-overview-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const applicationId = 'app-end-devices-overview-test'
  const application = {
    ids: { application_id: applicationId },
    name: 'Application End Devices Test Name',
    description: 'Application End Devices Test Description',
  }
  let endDeviceId

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
    cy.createMockDeviceAllComponents(applicationId).then(body => {
      endDeviceId = body.end_device.ids.device_id
    })
  })

  describe('Uplink', () => {
    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    })

    it('succeeds sending uplink message', () => {
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/applications/${applicationId}/devices/${endDeviceId}/messaging/uplink`,
      )

      cy.findByLabelText('FPort').type('1')
      cy.findByLabelText('Payload').type('0000')

      cy.findByRole('button', { name: 'Simulate uplink' }).click()

      cy.findByTestId('toast-notification').should('be.visible').and('contain', 'Uplink sent')

      cy.findByTestId('error-notification').should('not.exist')
    })

    it('disables uplink simulation when skip payload crypto is enabled', () => {
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/applications/${applicationId}/devices/${endDeviceId}/messaging/uplink`,
      )

      const response = {
        skip_payload_crypto_override: true,
        session: {},
      }

      cy.intercept(
        'GET',
        `/api/v3/as/applications/${applicationId}/devices/${endDeviceId}*`,
        response,
      )

      cy.findByTestId('notification')
        .should('be.visible')
        .findByText('Simulation is disabled for devices that skip payload crypto')
        .should('be.visible')

      cy.findByLabelText('FPort').should('be.disabled')
      cy.findByLabelText('Payload').should('be.disabled')

      cy.findByRole('button', { name: 'Simulate uplink' }).should('be.disabled')
    })
  })

  describe('Downlink', () => {
    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    })

    it('fails sending downlink message without valid session', () => {
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/applications/${applicationId}/devices/${endDeviceId}/messaging/downlink`,
      )

      cy.intercept('GET', `/api/v3/as/applications/${applicationId}/devices/${endDeviceId}*`, {})

      cy.findByTestId('notification')
        .should('be.visible')
        .findByText(
          `Downlinks can only be scheduled for end devices with a valid session. Please make sure your end device is properly connected to the network.`,
        )
        .should('be.visible')

      cy.findByLabelText('Replace downlink queue').should('be.disabled')
      cy.findByLabelText('Push to downlink queue (append)').should('be.disabled')
      cy.findByLabelText('FPort').should('be.disabled')
      cy.findByLabelText('Payload').should('be.disabled')
      cy.findByLabelText('Confirmed downlink').should('be.disabled')

      cy.findByRole('button', { name: 'Schedule downlink' }).should('be.disabled')
    })

    it('disables downlink messaging when skip payload crypto is enabled', () => {
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/applications/${applicationId}/devices/${endDeviceId}/messaging/downlink`,
      )

      const response = {
        skip_payload_crypto_override: true,
        session: {},
      }

      cy.intercept(
        'GET',
        `/api/v3/as/applications/${applicationId}/devices/${endDeviceId}*`,
        response,
      )

      cy.findByTestId('notification')
        .should('be.visible')
        .findByText('Simulation is disabled for devices that skip payload crypto')
        .should('be.visible')

      cy.findByLabelText('Replace downlink queue').should('be.disabled')
      cy.findByLabelText('Push to downlink queue (append)').should('be.disabled')
      cy.findByLabelText('FPort').should('be.disabled')
      cy.findByLabelText('Payload').should('be.disabled')
      cy.findByLabelText('Confirmed downlink').should('be.disabled')

      cy.findByRole('button', { name: 'Schedule downlink' }).should('be.disabled')
    })
  })
})
