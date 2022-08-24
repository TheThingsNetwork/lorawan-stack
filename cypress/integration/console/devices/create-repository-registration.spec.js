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

import {
  generateHexValue,
  disableNetworkServer,
  disableApplicationServer,
  disableJoinServer,
} from '../../../support/utils'

describe('End device repository manual registration', () => {
  const user = {
    ids: { user_id: 'create-dr-test-user' },
    primary_email_address: 'create-dr-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  describe('OTAA', () => {
    const appId = 'otaa-test-application'
    const application = {
      ids: { application_id: appId },
    }

    const selectDevice = ({ brand_id, model_id, hw_version, fw_version, band_id }) => {
      cy.findByLabelText('End device brand').selectOption(brand_id)
      cy.findByLabelText('Model').selectOption(model_id)
      cy.findByLabelText('Hardware Ver.').selectOption(hw_version)
      cy.findByLabelText('Firmware Ver.').selectOption(fw_version)
      cy.findByLabelText('Profile (Region)').selectOption(band_id)
    }

    before(() => {
      cy.createApplication(application, user.ids.user_id)
    })

    it('displays UI elements in place', () => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`)

      cy.findByText('Register end device', { selector: 'h1' }).should('be.visible')
      cy.findByLabelText('Select the end device in the LoRaWAN Device Repository').should(
        'be.checked',
      )
      cy.findByLabelText('End device brand').should('be.visible')
      cy.findByText(/Cannot find your exact end device?/).should('be.visible')
      cy.findByText(/enter end device specifics manually/).should('be.visible')
      cy.findByText('Please specify your device above to continue').should('be.visible')
      cy.findByRole('button', { name: 'Add end device' }).should('not.exist')

      cy.findByLabelText('End device brand').selectOption('_other_')
      cy.findByTestId('notification')
        .should('be.visible')
        .should('contain', 'Your end device will be added soon!')
        .within(() => {
          cy.findByText(
            /We're sorry, but your device is not yet part of The LoRaWAN Device Repository./,
          ).should('be.visible')
          cy.findByText(/enter end device specifics manually/).should('be.visible')
          cy.findByRole('link', { name: /Adding Devices/ })
            .should('be.visible')
            .and(
              'have.attr',
              'href',
              `https://thethingsindustries.com/docs/devices/adding-devices/`,
            )
        })
    })

    describe('Test Brand', () => {
      beforeEach(() => {
        cy.fixture('console/devices/repository/test-brand-otaa-model3.template.json').then(
          templateJson => {
            cy.intercept(
              'GET',
              `/api/v3/dr/applications/${appId}/brands/test-brand-otaa/models/test-model3/1.0.1/EU_863_870/template`,
              templateJson,
            )
          },
        )
        cy.fixture('console/devices/repository/test-brand-otaa-model2.template.json').then(
          templateJson => {
            cy.intercept(
              'GET',
              `/api/v3/dr/applications/${appId}/brands/test-brand-otaa/models/test-model2/1.0/EU_863_870/template`,
              templateJson,
            )
          },
        )
        cy.fixture('console/devices/repository/test-brand-otaa.models.json').then(modelsJson => {
          cy.intercept(
            'GET',
            `/api/v3/dr/applications/${appId}/brands/test-brand-otaa/models*`,
            modelsJson,
          )
        })
        cy.fixture('console/devices/repository/brands.json').then(brandsJson => {
          cy.intercept('GET', `/api/v3/dr/applications/${appId}/brands*`, brandsJson)
        })

        cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
        cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`)
      })

      it('succeeds handling incomplete model', () => {
        cy.findByLabelText('End device brand').selectOption('test-brand-otaa')
        cy.findByLabelText('Model').selectOption('test-model1')
        cy.findByLabelText('Hardware Ver.').selectOption('1.0')

        cy.findByTestId('notification')
          .should('be.visible')
          .should('contain', 'Your end device will be added soon!')
        cy.findByTestId('device-registration').should('not.exist')
      })

      it('succeeds shwoing modal on device type change', () => {
        cy.findByLabelText('End device brand').selectOption('test-brand-otaa')
        cy.findByLabelText('Model').selectOption('test-model2')

        cy.findByLabelText('Enter end device specifics manually').check()
        cy.findByTestId('modal-window')
          .should('be.visible')
          .within(() => {
            cy.findByText('Change input method', { selector: 'h1' }).should('be.visible')
            cy.findByText(
              'Are you sure you want to change the input method? Your current form progress will be lost.',
            ).should('be.visible')
            cy.findByRole('button', { name: /Change input method/ }).click()
          })
      })

      it('succeeds hiding provisioing information on JoinEUI reset', () => {
        cy.findByLabelText('End device brand').selectOption('test-brand-otaa')
        cy.findByLabelText('Model').selectOption('test-model2')

        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(generateHexValue(16))
        cy.findByLabelText('AppKey').type(generateHexValue(32))

        cy.findByRole('button', { name: 'Reset' }).click()
        cy.findByLabelText('DevEUI').should('not.exist')
        cy.findByLabelText('AppKey').should('not.exist')
      })

      it('succeeds registering device with single region, hardware and firmware versions', () => {
        const devEui = generateHexValue(16)

        // End device selection.
        cy.findByLabelText('End device brand').selectOption('test-brand-otaa')
        cy.findByLabelText('Model').selectOption('test-model2')

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(devEui)
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').should(
          'have.value',
          `eui-${devEui.toLocaleLowerCase()}`,
        )

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${devEui.toLocaleLowerCase()}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('validates before submitting an empty form', () => {
        cy.findByLabelText('End device brand').selectOption('test-brand-otaa')
        cy.findByLabelText('Model').selectOption('test-model2')
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`,
        )

        cy.findErrorByLabelText('DevEUI')
          .should('contain.text', 'DevEUI is required')
          .and('be.visible')
        cy.findErrorByLabelText('AppKey')
          .should('contain.text', 'AppKey is required')
          .and('be.visible')
        cy.findErrorByLabelText('End device ID')
          .should('contain.text', 'End device ID is required')
          .and('be.visible')
      })

      it('succeeds registering device', () => {
        const devEui = generateHexValue(16)

        // End device selection.
        cy.findByLabelText('End device brand').selectOption('test-brand-otaa')
        cy.findByLabelText('Model').selectOption('test-model3')
        cy.findByLabelText('Hardware Ver.').selectOption('2.0')
        cy.findByLabelText('Firmware Ver.').selectOption('1.0.1')
        cy.findByLabelText('Profile (Region)').selectOption('EU_863_870')

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(devEui)
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').should(
          'have.value',
          `eui-${devEui.toLocaleLowerCase()}`,
        )

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${devEui.toLocaleLowerCase()}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('The Things Products', () => {
      const selectUno = () => {
        selectDevice({
          brand_id: 'the-things-products',
          model_id: 'The Things Uno',
          hw_version: '1.0',
          fw_version: 'quickstart',
          band_id: 'EU_863_870',
        })
      }
      const selectNode = () => {
        selectDevice({
          brand_id: 'the-things-products',
          model_id: 'The Things Node',
          hw_version: '1.0',
          fw_version: '1.0',
          band_id: 'EU_863_870',
        })
      }

      beforeEach(() => {
        cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
        cy.visit(
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/repository`,
        )
      })

      it('displays UI elements in place', () => {
        selectUno()

        cy.findByLabelText('Frequency plan').should('be.visible')
        cy.findByLabelText('JoinEUI').should('be.visible')
        cy.findByRole('button', { name: 'Confirm' }).should('be.visible').and('be.disabled')
        cy.findByRole('button', { name: 'Add end device' }).should('not.exist')
        cy.findByText(
          'To continue, please enter the JoinEUI of the end device so we can determine onboarding options',
        ).should('be.visible')
      })

      it('succeeds registering the things uno', () => {
        const devEui = generateHexValue(16)

        selectUno()

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(devEui)
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').should(
          'have.value',
          `eui-${devEui.toLocaleLowerCase()}`,
        )

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/eui-${devEui}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering the things uno with custom ID', () => {
        const devId = 'uno-custom-id'

        selectUno()

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(generateHexValue(16))
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').clear().type(devId)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${devId}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering the things node', () => {
        const devEui = generateHexValue(16)

        selectNode()

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(devEui)
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').should(
          'have.value',
          `eui-${devEui.toLocaleLowerCase()}`,
        )

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${devEui.toLocaleLowerCase()}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering two the things uno', () => {
        const devEui1 = generateHexValue(16)

        // End device selection.
        selectUno()

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(devEui1)
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').should(
          'have.value',
          `eui-${devEui1.toLocaleLowerCase()}`,
        )
        cy.findByLabelText('Register another end device of this type').check()

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('End device registered')
          .should('be.visible')

        const devEui2 = generateHexValue(16)

        // End device registration.
        cy.findByLabelText('DevEUI').type(devEui2)
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').should(
          'have.value',
          `eui-${devEui2.toLocaleLowerCase()}`,
        )
        cy.findByLabelText('View registered end device').check()

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${devEui2.toLocaleLowerCase()}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('Only Join server enabled', () => {
      beforeEach(() => {
        cy.augmentStackConfig([disableNetworkServer, disableApplicationServer])
        cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
        cy.visit(
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/repository`,
        )
      })

      it('succeeds registering OTAA device', () => {
        const devEui = generateHexValue(16)

        // End device selection.
        selectDevice({
          brand_id: 'the-things-products',
          model_id: 'The Things Uno',
          hw_version: '1.0',
          fw_version: 'quickstart',
          band_id: 'EU_863_870',
        })

        // End device registration.
        cy.findByLabelText('Frequency plan').should('not.exist')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(devEui)
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').should(
          'have.value',
          `eui-${devEui.toLocaleLowerCase()}`,
        )

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${devEui.toLocaleLowerCase()}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('Join server disabled', () => {
      beforeEach(() => {
        cy.augmentStackConfig(disableJoinServer)
        cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
        cy.visit(
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/repository`,
        )
      })

      it('succeeds registering OTAA device', () => {
        const devEui = generateHexValue(16)

        // End device selection.
        selectDevice({
          brand_id: 'the-things-products',
          model_id: 'The Things Uno',
          hw_version: '1.0',
          fw_version: 'quickstart',
          band_id: 'EU_863_870',
        })

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(devEui).blur()
        cy.findByLabelText('AppKey').should('not.exist')
        cy.findByLabelText('End device ID').should(
          'have.value',
          `eui-${devEui.toLocaleLowerCase()}`,
        )

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${devEui.toLocaleLowerCase()}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })
  })

  describe('ABP', () => {
    const appId = 'abp-test-application'
    const application = {
      ids: { application_id: appId },
    }

    before(() => {
      cy.createApplication(application, user.ids.user_id)
    })

    describe('Test Brand', () => {
      beforeEach(() => {
        cy.fixture('console/devices/repository/test-brand-abp-model1.template.json').then(
          templateJson => {
            cy.intercept(
              'GET',
              `/api/v3/dr/applications/${appId}/brands/test-brand-abp/models/test-model1/1.0/EU_863_870/template`,
              templateJson,
            )
          },
        )
        cy.fixture('console/devices/repository/test-brand-abp.models.json').then(modelsJson => {
          cy.intercept(
            'GET',
            `/api/v3/dr/applications/${appId}/brands/test-brand-abp/models*`,
            modelsJson,
          )
        })
        cy.fixture('console/devices/repository/brands.json').then(brandsJson => {
          cy.intercept('GET', `/api/v3/dr/applications/${appId}/brands*`, brandsJson)
        })

        cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
        cy.visit(
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/repository`,
        )
      })

      it('validates before submitting an empty form', () => {
        cy.findByLabelText('End device brand').selectOption('test-brand-abp')
        cy.findByLabelText('Model').selectOption('test-model1')
        cy.findByLabelText('Profile (Region)').selectOption('EU_863_870')
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/repository`,
        )

        cy.findErrorByLabelText('Device address')
          .should('contain.text', 'Device address is required')
          .and('be.visible')
        cy.findErrorByLabelText('AppSKey')
          .should('contain.text', 'AppSKey is required')
          .and('be.visible')
        cy.findErrorByLabelText('NwkSKey')
          .should('contain.text', 'NwkSKey is required')
          .and('be.visible')
        cy.findErrorByLabelText('End device ID')
          .should('contain.text', 'End device ID is required')
          .and('be.visible')
      })

      it('succeeds registering device', () => {
        const devId = 'test-abp-dev'

        // End device selection.
        cy.findByLabelText('End device brand').selectOption('test-brand-abp')
        cy.findByLabelText('Model').selectOption('test-model1')
        cy.findByLabelText('Profile (Region)').selectOption('EU_863_870')

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('Device address').type(generateHexValue(8))
        cy.findByLabelText('AppSKey').type(generateHexValue(32))
        cy.findByLabelText('NwkSKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').type(devId)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${devId}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering two devices', () => {
        const devId1 = 'test-abp-dev-1'

        // End device selection.
        cy.findByLabelText('End device brand').selectOption('test-brand-abp')
        cy.findByLabelText('Model').selectOption('test-model1')
        cy.findByLabelText('Profile (Region)').selectOption('EU_863_870')

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('JoinEUI').type(generateHexValue(16))
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('Device address').type(generateHexValue(8))
        cy.findByLabelText('AppSKey').type(generateHexValue(32))
        cy.findByLabelText('NwkSKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').type(devId1)
        cy.findByLabelText('Register another end device of this type').check()

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('End device registered')
          .should('be.visible')

        const devId2 = 'test-abp-dev-2'

        // End device registration.
        cy.findByLabelText('Device address').type(generateHexValue(8))
        cy.findByLabelText('AppSKey').type(generateHexValue(32))
        cy.findByLabelText('NwkSKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').type(devId2)
        cy.findByLabelText('View registered end device').check()

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${devId2}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })
  })
})
