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

import { generateHexValue, disableJoinServer, disableNetworkServer } from '../../../support/utils'

describe('End device manual create 2', () => {
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
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
      cy.findByLabelText('Over the air activation (OTAA)').check()

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
          id: 'otaa-test-1-0-class-a',
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppEUI').type(device.app_eui)
        cy.findByLabelText('AppKey').type(device.app_key)
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
          id: 'otaa-test-1-0-class-b',
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
          class_b_timeout: 10,
        }

        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByRole('heading', {
          name: 'Show advanced activation, LoRaWAN class and cluster settings',
        }).click()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-b')
        cy.findByLabelText('Class B timeout').type(device.class_b_timeout)
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppEUI').type(device.app_eui)
        cy.findByLabelText('AppKey').type(device.app_key)
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
          id: 'otaa-test-1-0-class-c',
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByRole('heading', {
          name: 'Show advanced activation, LoRaWAN class and cluster settings',
        }).click()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-c')
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppEUI').type(device.app_eui)
        cy.findByLabelText('AppKey').type(device.app_key)
        cy.findByLabelText('End device ID').type(device.id)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.0.2', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          id: 'otaa-test-1-0-2-class-a',
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_0_2',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
        }

        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('AppEUI').type(device.app_eui)
        cy.findByLabelText('AppKey').type(device.app_key)
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
          id: 'otaa-test-1-1-class-a',
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          lorawan_version: 'MAC_V1_1',
          frequency_plan_id: '863-870 MHz',
          app_key: generateHexValue(32),
          nwk_key: generateHexValue(32),
        }

        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByLabelText('DevEUI').type(device.dev_eui)
        cy.findByLabelText('JoinEUI').type(device.app_eui)
        cy.findByLabelText('AppKey').type(device.app_key)
        cy.findByLabelText('NwkKey').type(device.nwk_key)
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
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
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
        }

        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByRole('heading', {
          name: 'Show advanced activation, LoRaWAN class and cluster settings',
        }).click()
        cy.findByLabelText('Activation by personalization (ABP)').check()
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
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
          class_b_timeout: 10,
          ping_slot_periodicity: 'EVERY_2S',
        }

        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByRole('heading', {
          name: 'Show advanced activation, LoRaWAN class and cluster settings',
        }).click()
        cy.findByLabelText('Activation by personalization (ABP)').check()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-b')
        cy.findByLabelText('Class B timeout').type(device.class_b_timeout)
        cy.findByLabelText('Ping Slot Periodicity').selectOption(device.ping_slot_periodicity)
        cy.findByLabelText('Device address').type(device.dev_addr)
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
          id: 'abp-test-1-0-class-c',
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_0',
          frequency_plan_id: '863-870 MHz',
          nwk_s_key: generateHexValue(32),
        }

        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByRole('heading', {
          name: 'Show advanced activation, LoRaWAN class and cluster settings',
        }).click()
        cy.findByLabelText('Activation by personalization (ABP)').check()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-c')
        cy.findByLabelText('Device address').type(device.dev_addr)
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

    describe('LoRaWAN V1.1', () => {
      it('succeeds registering a new class A end device', () => {
        const device = {
          id: 'abp-test-1-1-class-a',
          dev_addr: generateHexValue(8),
          lorawan_version: 'MAC_V1_1',
          frequency_plan_id: '863-870 MHz',
          f_nwk_s_int_key: generateHexValue(32),
          s_nwk_s_int_key: generateHexValue(32),
          nwk_s_enc_key: generateHexValue(32),
        }

        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByRole('heading', {
          name: 'Show advanced activation, LoRaWAN class and cluster settings',
        }).click()
        cy.findByLabelText('Activation by personalization (ABP)').check()
        cy.findByLabelText('Device address').type(device.dev_addr)
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
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
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
          nwk_s_key: generateHexValue(32),
          ping_slot_periodicity: 'EVERY_4S',
        }

        cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
        cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
        cy.findByRole('heading', {
          name: 'Show advanced activation, LoRaWAN class and cluster settings',
        }).click()
        cy.findByLabelText('Define multicast group (ABP & Multicast)').check()
        cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-b')
        cy.findByLabelText('Ping Slot Periodicity').selectOption(device.ping_slot_periodicity)
        cy.findByLabelText('Device address').type(device.dev_addr)
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

  describe('Without activation', () => {
    const appId = 'no-activation-test-app'
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
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
      cy.findByLabelText('Do not configure activation').check()

      cy.findByRole('button', { name: 'Register end device' }).click()
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/manual`,
      )
      cy.findByTestId('full-error-view').should('not.exist')
      cy.findByTestId('error-notification').should('not.exist')

      cy.findErrorByLabelText('End device ID')
        .should('contain.text', 'End device ID is required')
        .and('be.visible')
    })

    it('succeeds registering a new end device', () => {
      const devId = 'test-device-wo-activation'
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
      cy.findByLabelText('Do not configure activation').check()

      cy.findByLabelText('End device ID').type(devId)

      cy.findByRole('button', { name: 'Register end device' }).click()

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${devId}`,
      )
      cy.findByTestId('full-error-view').should('not.exist')
    })
  })

  describe('Join server disabled', () => {
    const appId = 'no-js-test-app'
    const application = {
      ids: { application_id: appId },
    }

    before(() => {
      cy.createApplication(application, user.ids.user_id)
    })

    beforeEach(() => {
      cy.augmentStackConfig(disableJoinServer)
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/manual`)
    })

    it('fails to select OTAA', () => {
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
      cy.findByLabelText('Over the air activation (OTAA)').should('have.attr', 'disabled')
      cy.findByLabelText('Activation by personalization (ABP)').should('have.attr', 'checked')
      cy.findByLabelText('AppEUI').should('not.exist')
      cy.findByLabelText('AppKey').should('not.exist')
      cy.findByLabelText('NwkKey').should('not.exist')
    })

    it('succeeds registering a new ABP end device', () => {
      const device = {
        id: 'abp-test-no-js',
        dev_addr: generateHexValue(8),
        lorawan_version: 'MAC_V1_0',
        frequency_plan_id: '863-870 MHz',
        nwk_s_key: generateHexValue(32),
      }

      cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
      cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
      cy.findByLabelText('Activation by personalization (ABP)').check()
      cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-c')
      cy.findByLabelText('Device address').type(device.dev_addr)
      cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
      cy.findByLabelText('End device ID').type(device.id)

      cy.findByRole('button', { name: 'Register end device' }).click()

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
      )
      cy.findByTestId('full-error-view').should('not.exist')
    })

    it('succeeds registering a new Multicast end device', () => {
      const device = {
        id: 'multicast-test-no-js',
        dev_addr: generateHexValue(8),
        lorawan_version: 'MAC_V1_0',
        frequency_plan_id: '863-870 MHz',
        nwk_s_key: generateHexValue(32),
        ping_slot_periodicity: 'EVERY_4S',
      }

      cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
      cy.findByLabelText('Frequency plan').selectOption(device.frequency_plan_id)
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
      cy.findByLabelText('Define multicast group (ABP & Multicast)').check()
      cy.findByLabelText('Additional LoRaWAN class capabilities').selectOption('class-b')
      cy.findByLabelText('Ping Slot Periodicity').selectOption(device.ping_slot_periodicity)
      cy.findByLabelText('Device address').type(device.dev_addr)
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

  describe('Network server disabled', () => {
    const appId = 'no-ns-test-app'
    const application = {
      ids: { application_id: appId },
    }

    before(() => {
      cy.createApplication(application, user.ids.user_id)
    })

    beforeEach(() => {
      cy.augmentStackConfig(disableNetworkServer)
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/manual`)
    })

    it('succeeds registering a new OTAA end device', () => {
      const device = {
        id: 'otaa-no-ns-1-0-class-a',
        app_eui: generateHexValue(16),
        dev_eui: generateHexValue(16),
        lorawan_version: 'MAC_V1_0',
        frequency_plan_id: '863-870 MHz',
        app_key: generateHexValue(32),
      }

      cy.findByLabelText('LoRaWAN version').selectOption(device.lorawan_version)
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
      cy.findByLabelText('Over the air activation (OTAA)').should('have.attr', 'checked')
      cy.findByLabelText('Over the air activation (OTAA)').should('not.have.attr', 'disabled')
      cy.findByLabelText('DevEUI').type(device.dev_eui)
      cy.findByLabelText('AppEUI').type(device.app_eui)
      cy.findByLabelText('AppKey').type(device.app_key)
      cy.findByLabelText('End device ID').type(device.id)

      cy.findByRole('button', { name: 'Register end device' }).click()

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
      )
      cy.findByTestId('full-error-view').should('not.exist')
    })

    it('fails to select ABP', () => {
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
      cy.findByLabelText('Over the air activation (OTAA)').should('have.attr', 'checked')
      cy.findByLabelText('Activation by personalization (ABP)').should('have.attr', 'disabled')
      cy.findByLabelText('Device address').should('not.exist')
      cy.findByLabelText('NwkSKey').should('not.exist')
      cy.findByLabelText('NwkKey').should('not.exist')
    })

    it('fails to select Multicast', () => {
      cy.findByRole('heading', {
        name: 'Show advanced activation, LoRaWAN class and cluster settings',
      }).click()
      cy.findByLabelText('Over the air activation (OTAA)').should('have.attr', 'checked')
      cy.findByLabelText('Define multicast group (ABP & Multicast)').should('have.attr', 'disabled')
      cy.findByLabelText('Device address').should('not.exist')
      cy.findByLabelText('NwkSKey').should('not.exist')
      cy.findByLabelText('NwkKey').should('not.exist')
    })
  })
})
