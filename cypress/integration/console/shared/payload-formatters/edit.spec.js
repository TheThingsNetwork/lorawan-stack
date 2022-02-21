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

describe('Payload formatters', () => {
  const applicationId = 'test-application-payload-formatters'
  const application = { ids: { application_id: applicationId } }
  const userId = 'edit-app-payload-formatter-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'edit-application-formatters-test-user@example.com',
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

  describe('Application', () => {
    describe('Uplink', () => {
      beforeEach(() => {
        cy.loginConsole({ user_id: userId, password: user.password })
      })

      it('succeeds changing formatter type to GRPC service', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('grpc')
        cy.findByLabelText('Formatter parameter').type('localhost')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to CayenneLPP', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('cayennelpp')
        cy.findByLabelText('Formatter parameter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to Repository', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('repository')
        cy.findByLabelText('Formatter parameter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to None', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('none')
        cy.findByLabelText('Formatter parameter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to Javascript', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })
    })

    describe('Downlink', () => {
      beforeEach(() => {
        cy.loginConsole({ user_id: userId, password: user.password })
      })

      it('succeeds changing formatter type to GRPC service', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('grpc')
        cy.findByLabelText('Formatter parameter').type('localhost')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to CayenneLPP', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('cayennelpp')
        cy.findByLabelText('Formatter parameter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to Repository', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('repository')
        cy.findByLabelText('Formatter parameter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to None', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('none')
        cy.findByLabelText('Formatter parameter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to Javascript', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })
    })
  })

  describe('End Devices', () => {
    describe('Uplink', () => {
      beforeEach(() => {
        cy.loginConsole({ user_id: userId, password: user.password })
      })

      it('succeeds changing formatter type to Javascript', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByRole('button', { name: 'Paste application formatter' }).should('be.visible')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds showing repository formatter button for formatter type Javascript', () => {
        const versionIdsResponseBody = {
          ids: {
            device_id: 'device-all-components',
            application_ids: {
              application_id: 'test-application-payload-formatters',
            },
            dev_eui: '70B3D57ED8000010',
            join_eui: '0000000000000000',
          },
          created_at: '2022-02-14T14:34:33.233Z',
          updated_at: '2022-02-14T14:34:33.233Z',
          version_ids: {
            brand_id: 'the-things-products',
            model_id: 'the-things-uno',
            hardware_version: '1.0',
            firmware_version: 'quickstart',
            band_id: 'US_902_928',
          },
          network_server_address: 'localhost',
          application_server_address: 'localhost',
          join_server_address: 'localhost',
        }
        cy.intercept(
          'GET',
          `/api/v3/applications/${applicationId}/devices/${endDeviceId}?field_mask=name,description,version_ids,network_server_address,application_server_address,join_server_address,locations,attributes`,
          versionIdsResponseBody,
        )
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByRole('button', { name: 'Paste repository formatter' }).should('be.visible')
      })

      it('succeeds not showing repository formatter button for formatter type Javascript', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByRole('button', { name: 'Paste repository formatter' }).should('not.exist')
      })

      it('succeeds changing formatter type to GRPC service', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('grpc')
        cy.findByLabelText('Formatter parameter').type('localhost')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to CayenneLPP', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('cayennelpp')
        cy.findByLabelText('Formatter parameter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to Repository', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('repository')
        cy.findByTestId('code-editor-repository-formatter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds showing code editor when there is a default formatter for type Repository', () => {
        const versionIdsResponseBody = {
          ids: {
            device_id: 'device-all-components',
            application_ids: {
              application_id: 'test-application-payload-formatters',
            },
            dev_eui: '70B3D57ED8000010',
            join_eui: '0000000000000000',
          },
          created_at: '2022-02-14T14:34:33.233Z',
          updated_at: '2022-02-14T14:34:33.233Z',
          version_ids: {
            brand_id: 'the-things-products',
            model_id: 'the-things-uno',
            hardware_version: '1.0',
            firmware_version: 'quickstart',
            band_id: 'US_902_928',
          },
          network_server_address: 'localhost',
          application_server_address: 'localhost',
          join_server_address: 'localhost',
        }
        cy.intercept(
          'GET',
          `/api/v3/applications/${applicationId}/devices/${endDeviceId}?field_mask=name,description,version_ids,network_server_address,application_server_address,join_server_address,locations,attributes`,
          versionIdsResponseBody,
        )
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('repository')
        cy.findByTestId('code-editor-repository-formatter').should('be.visible')
      })

      it('succeeds not showing code editor when there is no default formatter for type Repository', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('repository')
        cy.findByTestId('code-editor-repository-formatter').should('not.exist')
      })

      it('succeeds changing formatter type to None', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('none')
        cy.findByLabelText('Formatter parameter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to Application payload formatter', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('default')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })
    })

    describe('Downlink', () => {
      beforeEach(() => {
        cy.loginConsole({ user_id: userId, password: user.password })
      })

      it('succeeds changing formatter type to Javascript', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByRole('button', { name: 'Paste application formatter' }).should('be.visible')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds showing repository formatter button for formatter type Javascript', () => {
        const versionIdsResponseBody = {
          ids: {
            device_id: 'device-all-components',
            application_ids: {
              application_id: 'test-application-payload-formatters',
            },
            dev_eui: '70B3D57ED8000010',
            join_eui: '0000000000000000',
          },
          created_at: '2022-02-14T14:34:33.233Z',
          updated_at: '2022-02-14T14:34:33.233Z',
          version_ids: {
            brand_id: 'the-things-products',
            model_id: 'the-things-uno',
            hardware_version: '1.0',
            firmware_version: 'quickstart',
            band_id: 'US_902_928',
          },
          network_server_address: 'localhost',
          application_server_address: 'localhost',
          join_server_address: 'localhost',
        }
        cy.intercept(
          'GET',
          `/api/v3/applications/${applicationId}/devices/${endDeviceId}?field_mask=name,description,version_ids,network_server_address,application_server_address,join_server_address,locations,attributes`,
          versionIdsResponseBody,
        )
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByRole('button', { name: 'Paste repository formatter' }).should('be.visible')
      })

      it('succeeds not showing repository formatter button for formatter type Javascript', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByRole('button', { name: 'Paste repository formatter' }).should('not.exist')
      })

      it('succeeds changing formatter type to GRPC service', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('grpc')
        cy.findByLabelText('Formatter parameter').type('localhost')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to CayenneLPP', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('cayennelpp')
        cy.findByLabelText('Formatter parameter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to Repository', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('repository')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds showing code editor when there is a default formatter for type Repository', () => {
        const versionIdsResponseBody = {
          ids: {
            device_id: 'device-all-components',
            application_ids: {
              application_id: 'test-application-payload-formatters',
            },
            dev_eui: '70B3D57ED8000010',
            join_eui: '0000000000000000',
          },
          created_at: '2022-02-14T14:34:33.233Z',
          updated_at: '2022-02-14T14:34:33.233Z',
          version_ids: {
            brand_id: 'the-things-products',
            model_id: 'the-things-uno',
            hardware_version: '1.0',
            firmware_version: 'quickstart',
            band_id: 'US_902_928',
          },
          network_server_address: 'localhost',
          application_server_address: 'localhost',
          join_server_address: 'localhost',
        }
        cy.intercept(
          'GET',
          `/api/v3/applications/${applicationId}/devices/${endDeviceId}?field_mask=name,description,version_ids,network_server_address,application_server_address,join_server_address,locations,attributes`,
          versionIdsResponseBody,
        )
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('repository')
        cy.findByTestId('code-editor-repository-formatter').should('be.visible')
      })

      it('succeeds not showing code editor when there is no default formatter for type Repository', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('repository')
        cy.findByTestId('code-editor-repository-formatter').should('not.exist')
      })

      it('succeeds changing formatter type to None', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('none')
        cy.findByLabelText('Formatter parameter').should('not.exist')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('succeeds changing formatter type to Application payload formatter', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('default')

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })
    })
  })
})
