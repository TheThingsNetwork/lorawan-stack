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

import { generateHexValue } from '../../../../support/utils'

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
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`)
      cy.findByLabelText('Enter end device specifics manually').check()
    })

    it('displays UI elements in place', () => {
      cy.findByTestId('full-error-view').should('not.exist')
      cy.findByTestId('error-notification').should('not.exist')

      cy.findByLabelText('Frequency plan').should('be.visible')
      cy.findByLabelText('LoRaWAN version').should('be.visible')
      cy.findByLabelText('Regional Parameters version').should('be.visible')

      cy.findByText(
        'Please enter versions and frequency plan information above to continue',
      ).should('be.visible')
    })

    it('validates empty form before submitting', () => {
      const device = {
        lorawan_version: 'MAC_V1_0',
        frequency_plan_id: '863-870 MHz',
        join_eui: generateHexValue(16),
      }

      cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
      cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
      cy.findByLabelText('JoinEUI').type(device.join_eui)
      cy.findByRole('button', { name: 'Confirm' }).click()
      cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
      cy.findByLabelText('Network defaults').uncheck()
      cy.findByLabelText('Rx2 data rate').clear()
      cy.findByRole('button', { name: 'Add end device' }).click()

      cy.findErrorByLabelText('Rx2 data rate')
        .should('contain.text', 'Rx2 data rate is required')
        .and('be.visible')
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

    it('succeeds resetting provisioning fields on device type information change', () => {
      const device = {
        lorawan_version: 'MAC_V1_0',
        frequency_plan_id: '863-870 MHz',
        frequency_plan_id_2: '433 MHz',
        join_eui: generateHexValue(16),
        dev_eui: generateHexValue(16),
        app_key: generateHexValue(32),
      }

      cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
      cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
      cy.findByLabelText('JoinEUI').type(device.join_eui)
      cy.findByRole('button', { name: 'Confirm' }).click()
      cy.findByLabelText('DevEUI').type(device.dev_eui)
      cy.findByLabelText('AppKey').type(device.app_key)

      cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id_2)

      cy.findByLabelText('JoinEUI').should('not.exist')
      cy.findByLabelText('DevEUI').should('not.exist')
      cy.findByLabelText('AppKey').should('not.exist')
    })

    it('succeeds generating keys by clicking on `Generate` button', () => {
      const device = {
        lorawan_version: 'MAC_V1_0',
        frequency_plan_id: '863-870 MHz',
        join_eui: generateHexValue(16),
      }

      cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
      cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
      cy.findByLabelText('JoinEUI').type(device.join_eui)
      cy.findByRole('button', { name: 'Confirm' }).click()
      cy.findByRole('button', { name: 'Generate' }).click()
      cy.findByLabelText('AppKey').should('not.equal', '')
    })

    describe('LoRaWAN V1.0', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          join_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${device.dev_eui.toLocaleLowerCase()}`,
        )
        cy.findByRole('heading', { name: `eui-${device.dev_eui.toLocaleLowerCase()}` }).should(
          'be.visible',
        )

        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new class B end device', () => {
        const device = {
          join_eui: generateHexValue(16),
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
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${device.dev_eui.toLocaleLowerCase()}`,
        )
        cy.findByRole('heading', { name: `eui-${device.dev_eui.toLocaleLowerCase()}` }).should(
          'be.visible',
        )

        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new class C end device', () => {
        const device = {
          join_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-c')
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${device.dev_eui.toLocaleLowerCase()}`,
        )
        cy.findByRole('heading', { name: `eui-${device.dev_eui.toLocaleLowerCase()}` }).should(
          'be.visible',
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new end device skipping registration on join server', () => {
        const device = {
          join_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-c')
        cy.findByLabelText('Skip registration on Join Server').check()
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${device.dev_eui.toLocaleLowerCase()}`,
        )
        cy.findByRole('heading', { name: `eui-${device.dev_eui.toLocaleLowerCase()}` }).should(
          'be.visible',
        )

        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.0.1', () => {
      it('succeeds registering a new end device', () => {
        const device = {
          join_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0_1',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${device.dev_eui.toLocaleLowerCase()}`,
        )
        cy.findByRole('heading', { name: `eui-${device.dev_eui.toLocaleLowerCase()}` }).should(
          'be.visible',
        )

        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.0.2', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          join_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0_2',
          frequency_plan_id: '863-870 MHz',
          phy_version: 'PHY_V1_0_2_REV_A',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Regional Parameters version').selectOption(device.phy_version)
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${device.dev_eui.toLocaleLowerCase()}`,
        )
        cy.findByRole('heading', { name: `eui-${device.dev_eui.toLocaleLowerCase()}` }).should(
          'be.visible',
        )

        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.0.3', () => {
      it('succeeds registering a new end device', () => {
        const device = {
          join_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0_3',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${device.dev_eui.toLocaleLowerCase()}`,
        )
        cy.findByRole('heading', { name: `eui-${device.dev_eui.toLocaleLowerCase()}` }).should(
          'be.visible',
        )

        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.0.4', () => {
      it('succeeds registering a new end device', () => {
        const device = {
          join_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0_4',
          frequency_plan_id: '863-870 MHz',
          phy_version: 'PHY_V1_0',
          app_key: generateHexValue(32),
        }
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Regional Parameters version').selectOption(device.phy_version)
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppKey').type(device.app_key)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${device.dev_eui.toLocaleLowerCase()}`,
        )
        cy.findByRole('heading', { name: `eui-${device.dev_eui.toLocaleLowerCase()}` }).should(
          'be.visible',
        )

        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.1', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          join_eui: generateHexValue(16),
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
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppKey').type(device.app_key)
        cy.findByLabelText('NwkKey').type(device.nwk_key)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${device.dev_eui.toLocaleLowerCase()}`,
        )
        cy.findByRole('heading', { name: `eui-${device.dev_eui.toLocaleLowerCase()}` }).should(
          'be.visible',
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
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`)
      cy.findByLabelText('Enter end device specifics manually').check()
    })

    it('validates before submitting an empty form', () => {
      const device = {
        lorawan_version: 'MAC_V1_0',
        frequency_plan_id: '863-870 MHz',
        join_eui: generateHexValue(16),
      }

      cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
      cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
      cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
      cy.findByLabelText('Activation by personalization (ABP)').check()
      cy.findByLabelText('JoinEUI').type(device.join_eui)
      cy.findByRole('button', { name: 'Confirm' }).click()
      cy.findByRole('button', { name: 'Add end device' }).click()

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
      it('succeeds registering a new class A end device', () => {
        const device = {
          id: 'abp-test-1-0-class-a',
          join_eui: generateHexValue(16),
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
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByRole('heading', { name: `${device.id}` }).should('be.visible')

        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new class B end device', () => {
        const device = {
          id: 'abp-test-1-0-class-b',
          join_eui: generateHexValue(16),
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
        cy.findByLabelText('Network defaults').should('not.be.checked').and('have.attr', 'disabled')
        cy.findByLabelText('Class B timeout').type(device.class_b_timeout)
        cy.findByLabelText('Ping slot periodicity').selectOption(device.ping_slot_periodicity)
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByRole('heading', { name: `${device.id}` }).should('be.visible')

        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new class C end device', () => {
        const device = {
          id: 'abp-test-1-0-class-c',
          join_eui: generateHexValue(16),
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
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByRole('heading', { name: `${device.id}` }).should('be.visible')

        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.1', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          id: 'abp-test-1-1-class-a',
          join_eui: generateHexValue(16),
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
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('FNwkSIntKey').type(device.f_nwk_s_int_key)
        cy.findByLabelText('SNwkSIntKey').type(device.f_nwk_s_int_key)
        cy.findByLabelText('NwkSEncKey').type(device.f_nwk_s_int_key)
        cy.findByLabelText('End device ID').should(
          'have.value',
          `eui-${device.dev_eui.toLocaleLowerCase()}`,
        )

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config(
            'consoleRootPath',
          )}/applications/${appId}/devices/eui-${device.dev_eui.toLocaleLowerCase()}`,
        )
        cy.findByRole('heading', { name: `eui-${device.dev_eui.toLocaleLowerCase()}` }).should(
          'be.visible',
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
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`)
      cy.findByLabelText('Enter end device specifics manually').check()
    })

    it('validates before submitting an empty form', () => {
      const device = {
        lorawan_version: 'MAC_V1_0',
        frequency_plan_id: '863-870 MHz',
        join_eui: generateHexValue(16),
      }

      cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
      cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
      cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
      cy.findByLabelText('Define multicast group (ABP & Multicast)').check()
      cy.findByLabelText('JoinEUI').type(device.join_eui)
      cy.findByRole('button', { name: 'Confirm' }).click()
      cy.findByRole('button', { name: 'Add end device' }).click()

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
          join_eui: generateHexValue(16),
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
        cy.findByLabelText('Network defaults').should('not.be.checked').and('have.attr', 'disabled')
        cy.findByLabelText('Ping slot periodicity').selectOption(device.ping_slot_periodicity)
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByRole('heading', { name: `${device.id}` }).should('be.visible')

        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering a new class C end device', () => {
        const device = {
          id: 'multicast-test-1-0-class-c',
          join_eui: generateHexValue(16),
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_s_key: generateHexValue(32),
          nwk_s_key: generateHexValue(32),
        }

        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Define multicast group (ABP & Multicast)').check()
        cy.findByLabelText('LoRaWAN class for multicast downlinks').selectOption('class-c')
        cy.findByLabelText('Network defaults').uncheck()
        cy.findByLabelText('JoinEUI').type(device.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByRole('heading', { name: `${device.id}` }).should('be.visible')

        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds regitering multiple devices', () => {
        const device1 = {
          id: 'multicast-test-1',
          join_eui: generateHexValue(16),
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_s_key: generateHexValue(32),
          nwk_s_key: generateHexValue(32),
        }

        const device2 = {
          id: 'multicast-test-2',
          join_eui: generateHexValue(16),
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_s_key: generateHexValue(32),
          nwk_s_key: generateHexValue(32),
        }

        const device3 = {
          id: 'multicast-test-3',
          join_eui: generateHexValue(16),
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_s_key: generateHexValue(32),
          nwk_s_key: generateHexValue(32),
        }

        // Device 1
        cy.findByLabelText('Frequency plan').selectOption(device1.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').selectOption(device1.lorawan_version)
        cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
        cy.findByLabelText('Define multicast group (ABP & Multicast)').check()
        cy.findByLabelText('LoRaWAN class for multicast downlinks').selectOption('class-c')
        cy.findByLabelText('Network defaults').uncheck()
        cy.findByLabelText('JoinEUI').type(device1.join_eui)
        cy.findByRole('button', { name: 'Confirm' }).click()
        cy.findByLabelText('Device address').type(device1.dev_addr)
        cy.findByLabelText('AppSKey').type(device1.app_s_key)
        cy.findByLabelText('NwkSKey').type(device1.nwk_s_key)
        cy.findByLabelText('End device ID').type(device1.id)
        cy.findByLabelText('Register another end device of this type').check()

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('End device registered')
          .should('be.visible')

        // Device 2
        cy.findByLabelText('Device address').type(device2.dev_addr)
        cy.findByLabelText('AppSKey').clear().type(device2.app_s_key)
        cy.findByLabelText('NwkSKey').clear().type(device2.nwk_s_key)
        cy.findByLabelText('End device ID').type(device2.id)
        cy.findByLabelText('Register another end device of this type').check()

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('End device registered')
          .should('be.visible')

        // Device 3
        cy.findByLabelText('Device address').clear().type(device3.dev_addr)
        cy.findByLabelText('AppSKey').clear().type(device3.app_s_key)
        cy.findByLabelText('NwkSKey').clear().type(device3.nwk_s_key)
        cy.findByLabelText('End device ID').type(device3.id)
        cy.findByLabelText('View registered end device').check()

        cy.findByRole('button', { name: 'Add end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device3.id}`,
        )
        cy.findByRole('heading', { name: `${device3.id}` }).should('be.visible')

        cy.findByTestId('full-error-view').should('not.exist')
      })
    })
  })
})
