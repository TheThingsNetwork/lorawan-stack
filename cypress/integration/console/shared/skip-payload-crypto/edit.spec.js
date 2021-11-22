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
    const applicationId = 'test-application-spc'
    const application = { ids: { application_id: applicationId } }
    const userId = 'edit-app-spc-test-user'
    const user = {
      ids: { user_id: userId },
      primary_email_address: 'edit-spc-test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
  
    const endDeviceId = 'end-device-spc-test'
    const endDevice = {
      application_server_address: 'localhost',
      ids: {
        device_id: endDeviceId,
        dev_eui: '0000000000000001',
        join_eui: '0000000000000000',
      },
      name: 'End Device Test Name',
      description: 'End Device Test Description',
      join_server_address: 'localhost',
      network_server_address: 'localhost',
    }
    const endDeviceFieldMask = {
      paths: [
        'join_server_address',
        'network_server_address',
        'application_server_address',
        'ids.dev_eui',
        'ids.join_eui',
        'name',
        'description',
        'skip_payload_crypto_override',
      ],
    }
    const endDeviceRequestBody = {
      end_device: endDevice,
      field_mask: endDeviceFieldMask,
    }
  
    before(() => {
      cy.dropAndSeedDatabase()
      cy.createUser(user)
      cy.createApplication(application, userId)
      cy.createEndDevice(applicationId, endDeviceRequestBody)
    })
  
    describe('Uplink', () => {
      beforeEach(() => {
        cy.loginConsole({ user_id: userId, password: user.password })
      })
  
      describe('application link skips payload crypto', () => {
        const adminApiKey = Cypress.config('adminApiKey')
        const linkRequestBody = {
          link: {
            default_formatters: {},
            skip_payload_crypto: false,
          },
        }
        cy.request({
          method: 'PUT',
          url: `api/v3/as/applications/${applicationId}/link?field_mask=skip_payload_crypto`,
          body: linkRequestBody,
          headers: {
            Authorization: `Bearer ${adminApiKey}`,
          },
        })
  
        it('disables messaging when not using a SPC overwrite', () => {
          const response = {
            skip_payload_crypto_override: null,
            session: {},
          }
  
          cy.request({
            method: 'POST',
            url: `api/v3/as/applications/${applicationId}/devices/${endDeviceId}?field_mask=version_ids,formatters,skip_payload_crypto_override,session,pending_session`,
            body: response,
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
  
          cy.findByLabelText('FPort').should('be.disabled')
          cy.findByLabelText('Payload').should('be.disabled')
  
          cy.findByRole('button', { name: 'Simulate uplink' }).should('be.disabled')
        })
  
        it('allows messaging when device disabled SPC via overwrite', () => {
          const response = {
            skip_payload_crypto_override: false,
            session: {},
          }
  
          cy.request({
            method: 'POST',
            url: `api/v3/as/applications/${applicationId}/devices/${endDeviceId}?field_mask=version_ids,formatters,skip_payload_crypto_override,session,pending_session`,
            body: response,
          })
  
          cy.visit(
            `${Cypress.config(
              'consoleRootPath',
            )}/applications/${applicationId}/devices/${endDeviceId}/messaging/uplink`,
          )
  
          cy.findByLabelText('FPort').should('be.enabled')
          cy.findByLabelText('Payload').should('be.enabled')
  
          cy.findByRole('button', { name: 'Simulate uplink' }).should('be.enabled')
        })
      })
  
      describe('application link does not skip payload crypto', () => {
        const adminApiKey = Cypress.config('adminApiKey')
        const linkRequestBody = {
          link: {
            default_formatters: {},
            skip_payload_crypto: false,
          },
        }
        cy.request({
          method: 'PUT',
          url: `api/v3/as/applications/${applicationId}/link?field_mask=skip_payload_crypto`,
          body: linkRequestBody,
          headers: {
            Authorization: `Bearer ${adminApiKey}`,
          },
        })
  
        it('allows messaging when not using a SPC overwrite', () => {
          const response = {
            skip_payload_crypto_override: null,
            session: {},
          }
  
          cy.request({
            method: 'POST',
            url: `api/v3/as/applications/${applicationId}/devices/${endDeviceId}?field_mask=version_ids,formatters,skip_payload_crypto_override,session,pending_session`,
            body: response,
          })
  
          cy.visit(
            `${Cypress.config(
              'consoleRootPath',
            )}/applications/${applicationId}/devices/${endDeviceId}/messaging/uplink`,
          )
  
          cy.findByLabelText('FPort').should('be.enabled')
          cy.findByLabelText('Payload').should('be.enabled')
  
          cy.findByRole('button', { name: 'Simulate uplink' }).should('be.enabled')
        })
  
        it('disables messaging when device disabled SPC via overwrite', () => {
          const response = {
            skip_payload_crypto_override: true,
            session: {},
          }
  
          cy.request({
            method: 'POST',
            url: `api/v3/as/applications/${applicationId}/devices/${endDeviceId}?field_mask=version_ids,formatters,skip_payload_crypto_override,session,pending_session`,
            body: response,
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
  
          cy.findByLabelText('FPort').should('be.disabled')
          cy.findByLabelText('Payload').should('be.disabled')
  
          cy.findByRole('button', { name: 'Simulate uplink' }).should('be.disabled')
        })
      })
    })
  })
  