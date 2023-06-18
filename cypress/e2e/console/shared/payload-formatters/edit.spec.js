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

  const ns = {
    end_device: {
      frequency_plan_id: 'EU_863_870_TTN',
      lorawan_phy_version: 'PHY_V1_0_2_REV_A',
      multicast: false,
      supports_join: true,
      lorawan_version: 'MAC_V1_0_2',
      version_ids: {
        brand_id: 'the-things-products',
        model_id: 'the-things-uno',
        hw_version: '1.0',
        fw_version: 'quickstart',
        band_id: 'EU_863_870',
      },
      ids: {
        device_id: 'device-all-components',
        dev_eui: '70B3D57ED8000013',
        join_eui: '0000000000000006',
      },
      supports_class_c: false,
      supports_class_b: false,
      mac_settings: {
        rx2_data_rate_index: 0,
        rx2_frequency: 869525000,
        rx1_delay: 1,
        rx1_data_rate_offset: 0,
        resets_f_cnt: false,
      },
    },
    field_mask: {
      paths: [
        'version_ids.brand_id',
        'version_ids.model_id',
        'version_ids.hardware_version',
        'version_ids.firmware_version',
        'version_ids.band_id',
        'frequency_plan_id',
        'lorawan_phy_version',
        'multicast',
        'supports_join',
        'lorawan_version',
        'ids.device_id',
        'ids.dev_eui',
        'ids.join_eui',
        'supports_class_c',
        'supports_class_b',
      ],
    },
  }

  const is = {
    end_device: {
      ids: {
        dev_eui: '70B3D57ED8000013',
        join_eui: '0000000000000006',
        device_id: 'device-all-components',
      },
      network_server_address: 'localhost',
      application_server_address: 'localhost',
      join_server_address: 'localhost',
    },
    field_mask: {
      paths: ['network_server_address', 'application_server_address', 'join_server_address'],
    },
  }

  let endDeviceId
  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
    cy.setApplicationPayloadFormatter(applicationId)
    cy.createMockDeviceAllComponents(applicationId, undefined, { ns, is }).then(body => {
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
        cy.findByLabelText('GRPC host').type('localhost')

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
        cy.findByLabelText('GRPC host').type('localhost')

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

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('provides formatter options for devices with a repository formatter', () => {
        const repositoryFormatter = {
          formatter_parameter: 'Test formatter parameter',
        }
        cy.intercept(
          'GET',
          `/api/v3/dr/applications/test-application-payload-formatters/**`,
          repositoryFormatter,
        )

        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByRole('button', { name: 'Paste repository formatter' }).should('be.visible')
        cy.findByLabelText('Formatter type').selectOption('repository')
        cy.findByTestId('code-editor-repository-formatter').should('be.visible')
      })

      it('provides formatter options for devices of applications with application payload formatter', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByRole('button', { name: 'Paste repository formatter' }).should('not.exist')
        cy.findByRole('button', { name: 'Paste application formatter' }).should('be.visible')
        cy.findByLabelText('Formatter type').selectOption('application')
        cy.findByTestId('code-editor-repository-formatter').should('not.exist')
      })

      it('succeeds changing formatter type to GRPC service', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('grpc')
        cy.findByLabelText('GRPC host').type('localhost')

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
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/uplink`,
        )

        cy.findByLabelText('Formatter type').selectOption('none')

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

        cy.findByRole('button', { name: 'Save changes' }).click()

        cy.findByTestId('error-notification').should('not.exist')
        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('Payload formatter updated')
          .should('be.visible')
      })

      it('provides formatter options for devices with a repository formatter', () => {
        const repositoryFormatter = {
          formatter_parameter: 'Test formatter parameter',
        }
        cy.intercept(
          'GET',
          `/api/v3/dr/applications/test-application-payload-formatters/**`,
          repositoryFormatter,
        )

        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByRole('button', { name: 'Paste repository formatter' }).should('be.visible')
        cy.findByLabelText('Formatter type').selectOption('repository')
        cy.findByTestId('code-editor-repository-formatter').should('be.visible')
      })

      it('provides formatter options for devices of applications with application payload formatter', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/downlink`,
        )

        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByLabelText('Formatter type').selectOption('javascript')
        cy.findByTestId('code-editor-javascript-formatter').should('be.visible')
        cy.findByRole('button', { name: 'Paste repository formatter' }).should('not.exist')
        cy.findByRole('button', { name: 'Paste application formatter' }).should('be.visible')
      })

      it('succeeds changing formatter type to GRPC service', () => {
        cy.visit(
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('grpc')
        cy.findByLabelText('GRPC host').type('localhost')

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
          )}/applications/${applicationId}/devices/${endDeviceId}/payload-formatters/downlink`,
        )

        cy.findByLabelText('Formatter type').selectOption('none')

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
