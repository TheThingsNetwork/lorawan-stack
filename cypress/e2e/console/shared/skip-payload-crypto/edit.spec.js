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

describe('Skip payload crypto', () => {
  const applicationId = 'spc-test-application'
  const application = { ids: { application_id: applicationId } }
  const userId = 'edit-app-spc-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'edit-spc-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
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
    describe('application link skips payload crypto', () => {
      beforeEach(() => {
        cy.loginConsole({ user_id: userId, password: user.password })
      })

      it('disables messaging when not using a SPC overwrite', () => {
        const adminApiKey = Cypress.config('adminApiKey')
        const linkRequestBody = {
          field_mask: { paths: ['skip_payload_crypto'] },
          link: {
            default_formatters: {},
            skip_payload_crypto: true,
          },
        }
        cy.request({
          method: 'PUT',
          url: `api/v3/as/applications/${applicationId}/link`,
          body: linkRequestBody,
          headers: {
            Authorization: `Bearer ${adminApiKey}`,
          },
        })

        const endDeviceRequestBody = {
          field_mask: { paths: ['skip_payload_crypto_override'] },
          end_device: { skip_payload_crypto_override: null },
        }
        cy.request({
          method: 'PUT',
          url: `api/v3/as/applications/${applicationId}/devices/${endDeviceId}`,
          body: endDeviceRequestBody,
          headers: {
            Authorization: `Bearer ${adminApiKey}`,
          },
        })

        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/messaging/uplink`,
        )

        cy.findByTestId('notification')
          .should('be.visible')
          .findByText(`Simulation is disabled for devices that skip payload crypto`)
          .should('be.visible')

        cy.findByRole('button', { name: 'Simulate uplink' }).should('be.disabled')
      })

      it('allows messaging when device disabled SPC via overwrite', () => {
        const adminApiKey = Cypress.config('adminApiKey')
        const linkRequestBody = {
          field_mask: { paths: ['skip_payload_crypto'] },
          link: {
            default_formatters: {},
            skip_payload_crypto: true,
          },
        }
        cy.request({
          method: 'PUT',
          url: `api/v3/as/applications/${applicationId}/link`,
          body: linkRequestBody,
          headers: {
            Authorization: `Bearer ${adminApiKey}`,
          },
        })

        const endDeviceRequestBody = {
          field_mask: { paths: ['skip_payload_crypto_override'] },
          end_device: { skip_payload_crypto_override: false },
        }
        cy.request({
          method: 'PUT',
          url: `api/v3/as/applications/${applicationId}/devices/${endDeviceId}`,
          body: endDeviceRequestBody,
          headers: {
            Authorization: `Bearer ${adminApiKey}`,
          },
        })

        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/messaging/uplink`,
        )

        cy.findByRole('button', { name: 'Simulate uplink' }).should('be.enabled')
      })
    })

    describe('application link does not skip payload crypto', () => {
      beforeEach(() => {
        cy.loginConsole({ user_id: userId, password: user.password })
      })

      it('allows messaging when not using a SPC overwrite', () => {
        const adminApiKey = Cypress.config('adminApiKey')
        const linkRequestBody = {
          field_mask: { paths: ['skip_payload_crypto'] },
          link: {
            default_formatters: {},
            skip_payload_crypto: false,
          },
        }
        cy.request({
          method: 'PUT',
          url: `api/v3/as/applications/${applicationId}/link`,
          body: linkRequestBody,
          headers: {
            Authorization: `Bearer ${adminApiKey}`,
          },
        })

        const endDeviceRequestBody = {
          field_mask: { paths: ['skip_payload_crypto_override'] },
          end_device: { skip_payload_crypto_override: null },
        }
        cy.request({
          method: 'PUT',
          url: `api/v3/as/applications/${applicationId}/devices/${endDeviceId}`,
          body: endDeviceRequestBody,
          headers: {
            Authorization: `Bearer ${adminApiKey}`,
          },
        })

        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/messaging/uplink`,
        )

        cy.findByRole('button', { name: 'Simulate uplink' }).should('be.enabled')
      })

      it('disables messaging when device disabled SPC via overwrite', () => {
        const adminApiKey = Cypress.config('adminApiKey')
        const linkRequestBody = {
          field_mask: { paths: ['skip_payload_crypto'] },
          link: {
            default_formatters: {},
            skip_payload_crypto: false,
          },
        }
        cy.request({
          method: 'PUT',
          url: `api/v3/as/applications/${applicationId}/link`,
          body: linkRequestBody,
          headers: {
            Authorization: `Bearer ${adminApiKey}`,
          },
        })

        const endDeviceRequestBody = {
          field_mask: { paths: ['skip_payload_crypto_override'] },
          end_device: { skip_payload_crypto_override: true },
        }
        cy.request({
          method: 'PUT',
          url: `api/v3/as/applications/${applicationId}/devices/${endDeviceId}`,
          body: endDeviceRequestBody,
          headers: {
            Authorization: `Bearer ${adminApiKey}`,
          },
        })

        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/messaging/uplink`,
        )

        cy.findByTestId('notification')
          .should('be.visible')
          .findByText(`Simulation is disabled for devices that skip payload crypto`)
          .should('be.visible')

        cy.findByRole('button', { name: 'Simulate uplink' }).should('be.disabled')
      })
    })
  })
})
