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

import { generateHexValue } from '../../../support/utils'

describe('End device manual create', () => {
  const user = {
    ids: { user_id: 'create-manual-test-user' },
    primary_email_address: 'create-manual-test-user@example.com',
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

    before(() => {
      cy.createApplication(application, user.ids.user_id)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/manual`)
    })

    it('validates before submitting an empty form', () => {
      cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
      cy.findByLabelText('Over the air activation (OTAA)').check()

      cy.findByRole('button', { name: 'Register end device' }).click()
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`,
      )
      cy.findByTestId('full-error-view').should('not.exist')
      cy.findByTestId('error-notification').should('not.exist')

      cy.findErrorByLabelText('LoRaWAN version')
        .should('contain.text', 'LoRaWAN version is required')
        .and('be.visible')
      cy.findErrorByLabelText('Regional Parameters version')
        .should('contain.text', 'Regional Parameters version is required')
        .and('be.visible')
      cy.findErrorByLabelText('Frequency plan')
        .should('contain.text', 'Frequency plan is required')
        .and('be.visible')
      cy.findErrorByLabelText('DevEUI')
        .should('contain.text', 'DevEUI is required')
        .and('be.visible')
      cy.findErrorByLabelText('AppEUI')
        .should('contain.text', 'AppEUI is required')
        .and('be.visible')
      cy.findErrorByLabelText('AppKey')
        .should('contain.text', 'AppKey is required')
        .and('be.visible')
      cy.findErrorByLabelText('End device ID')
        .should('contain.text', 'End device ID is required')
        .and('be.visible')
    })

    describe('LoRaWAN V1.0', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppEUI').type(device.app_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/eui-${
            device.dev_eui
          }`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new class B end device', () => {
        const device = {
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
          class_b_timeout: 10,
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-b')
        cy.findByLabelText('Network defaults').uncheck()
        cy.findByLabelText('Class B timeout').type(device.class_b_timeout)
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppEUI').type(device.app_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/eui-${
            device.dev_eui
          }`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new class C end device', () => {
        const device = {
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-c')
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppEUI').type(device.app_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/eui-${
            device.dev_eui
          }`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new end device without join server address', () => {
        const device = {
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-c')
        cy.findByLabelText('Cluster settings').check()
        cy.findByLabelText('Join Server address').clear()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppEUI').type(device.app_eui)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/eui-${
            device.dev_eui
          }`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.0.2', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0_2',
          frequency_plan_id: '863-870 MHz',
          phy_version: 'PHY_V1_0_2_REV_A',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Regional Parameters version').selectOption(device.phy_version)
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppEUI').type(device.app_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/eui-${
            device.dev_eui
          }`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.1', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_1',
          phy_version: 'PHY_V1_1_REV_A',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
          nwk_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Regional Parameters version').selectOption(device.phy_version)
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('JoinEUI').type(device.app_eui)
        cy.findByLabelText('AppKey').type(device.app_key)
        cy.findByLabelText('NwkKey').type(device.nwk_key)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/eui-${
            device.dev_eui
          }`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('Uses external join server address', () => {
      it('succeeds registering a new end device', () => {
        const device = {
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          join_server_address: 'external-js-address',
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Use external LoRaWAN backend servers').check()
        cy.findByLabelText('Join Server address').clear().type(device.join_server_address)
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppEUI').type(device.app_eui)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/eui-${
            device.dev_eui
          }`,
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

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/manual`)
    })

    it('validates before submitting an empty form', () => {
      cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
      cy.findByLabelText('Activation by personalization (ABP)').check()

      cy.findByRole('button', { name: 'Register end device' }).click()
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/manual`,
      )
      cy.findByTestId('full-error-view').should('not.exist')
      cy.findByTestId('error-notification').should('not.exist')

      cy.findErrorByLabelText('LoRaWAN version')
        .should('contain.text', 'LoRaWAN version is required')
        .and('be.visible')
      cy.findErrorByLabelText('Regional Parameters version')
        .should('contain.text', 'Regional Parameters version is required')
        .and('be.visible')
      cy.findErrorByLabelText('Frequency plan')
        .should('contain.text', 'Frequency plan is required')
        .and('be.visible')
      cy.findErrorByLabelText('AppSKey')
        .should('contain.text', 'AppSKey is required')
        .and('be.visible')
      cy.findErrorByLabelText('Device address')
        .should('contain.text', 'Device address is required')
        .and('be.visible')
      cy.findErrorByLabelText('NwkSKey')
        .should('contain.text', 'NwkSKey is required')
        .and('be.visible')
      cy.findErrorByLabelText('End device ID')
        .should('contain.text', 'End device ID is required')
        .and('be.visible')
    })

    describe('LoRaWAN V1.0', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          id: 'abp-test-1-0-class-a',
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          nwk_s_key: generateHexValue(32),
          app_s_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Activation by personalization (ABP)').check()
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new class B end device', () => {
        const device = {
          id: 'abp-test-1-0-class-b',
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          nwk_s_key: generateHexValue(32),
          app_s_key: generateHexValue(32),
          class_b_timeout: 10,
          ping_slot_periodicity: 'EVERY_2S',
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Activation by personalization (ABP)').check()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-b')
        cy.findByLabelText('Network defaults').uncheck()
        cy.findByLabelText('Class B timeout').type(device.class_b_timeout)
        cy.findByLabelText('Ping slot periodicity').selectOption(device.ping_slot_periodicity)
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new class C end device', () => {
        const device = {
          id: 'abp-test-1-0-class-c',
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          nwk_s_key: generateHexValue(32),
          app_s_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Activation by personalization (ABP)').check()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-c')
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.1', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          id: 'abp-test-1-1-class-a',
          dev_eui: generateHexValue(16),
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_1',
          phy_version: 'PHY_V1_1_REV_A',
          frequency_plan_id: '863-870 MHz',
          app_s_key: generateHexValue(32),
          f_nwk_s_int_key: generateHexValue(32),
          s_nwk_s_int_key: generateHexValue(32),
          nwk_s_enc_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Regional Parameters version').selectOption(device.phy_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Activation by personalization (ABP)').check()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('FNwkSIntKey').type(device.f_nwk_s_int_key)
        cy.findByLabelText('SNwkSIntKey').type(device.f_nwk_s_int_key)
        cy.findByLabelText('NwkSEncKey').type(device.f_nwk_s_int_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })
  })

  describe('Multicast', () => {
    const appId = 'multicast-test-application'
    const application = {
      ids: { application_id: appId },
    }

    before(() => {
      cy.createApplication(application, user.ids.user_id)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/manual`)
    })

    it('validates before submitting an empty form', () => {
      cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
      cy.findByLabelText('Define multicast group (ABP & Multicast)').check()

      cy.findByRole('button', { name: 'Register end device' }).click()
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/manual`,
      )
      cy.findByTestId('full-error-view').should('not.exist')
      cy.findByTestId('error-notification').should('not.exist')

      cy.findErrorByLabelText('LoRaWAN version')
        .should('contain.text', 'LoRaWAN version is required')
        .and('be.visible')
      cy.findErrorByLabelText('Regional Parameters version')
        .should('contain.text', 'Regional Parameters version is required')
        .and('be.visible')
      cy.findErrorByLabelText('Frequency plan')
        .should('contain.text', 'Frequency plan is required')
        .and('be.visible')
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

    describe('LoRaWAN V1.0', () => {
      it('succeeds registering a new class B end device', () => {
        const device = {
          id: 'multicast-test-1-0-class-b',
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_s_key: generateHexValue(32),
          nwk_s_key: generateHexValue(32),
          ping_slot_periodicity: 'EVERY_4S',
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Define multicast group (ABP & Multicast)').check()
        cy.findByLabelText('LoRaWAN class for multicast downlinks').selectOption('class-b')
        cy.findByLabelText('Ping slot periodicity').selectOption(device.ping_slot_periodicity)
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new class C end device', () => {
        const device = {
          id: 'multicast-test-1-0-class-c',
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_s_key: generateHexValue(32),
          nwk_s_key: generateHexValue(32),
          ping_slot_periodicity: 'EVERY_4S',
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Define multicast group (ABP & Multicast)').check()
        cy.findByLabelText('LoRaWAN class for multicast downlinks').selectOption('class-c')
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })
  })
})
