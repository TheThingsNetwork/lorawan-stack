// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import {
  ConfigurationStep,
  BasicSettingsStep,
  NetworkLayerStep,
  JoinSettingsStep,
  ApplicationLayerStep,
} from './utils'

describe('End device create', () => {
  const user = {
    ids: { user_id: 'create-device-test-user' },
    primary_email_address: 'create-device-test-user@example.com',
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
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.createApplication(application, user.ids.user_id)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`)
    })

    describe('LoRaWAN V1.0', () => {
      it('succeeds setting MAC settings for class A', () => {
        const device = {
          id: 'otaa-test-mac-settings-class-a',
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          name: 'Test OTAA MAC settings',
          description: 'Test MAC settings device',
          lorawan_version: 'MAC V1.0',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.0',
          app_key: generateHexValue(32),
          rx2_data_rate_index: 1,
          rx2_frequency: '869525000',
          factory_frequencies: ['868100000', '868300000'],
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkOTAA()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillAppEUI(device.app_eui)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.openAdvancedSettings()
        networkLayerStep.check32BitFCnt()
        networkLayerStep.fillRx2DataRateIndex(device.rx2_data_rate_index)
        networkLayerStep.fillRx2Frequency(device.rx2_frequency)
        networkLayerStep.fillFactoryPresetFrequencies(device.factory_frequencies)
        networkLayerStep.goToJoinSettingsStep()

        const joinSettingsStep = new JoinSettingsStep({ lorawanVersion: device.lorawan_version })
        joinSettingsStep.fillAppKey(device.app_key)
        joinSettingsStep.submit()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds setting MAC settings for class B', () => {
        const device = {
          id: 'otaa-test-mac-settings-class-b',
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          name: 'Test OTAA MAC settings',
          description: 'Test MAC settings device',
          lorawan_version: 'MAC V1.0',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.0',
          app_key: generateHexValue(32),
          rx2_data_rate_index: 1,
          rx2_frequency: '869525000',
          factory_frequencies: ['868100000', '868300000'],
          ping_slot_periodicity: 'PING_EVERY_1S',
          ping_slot_frequency: '869525000',
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkOTAA()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillAppEUI(device.app_eui)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.checkClassB()
        networkLayerStep.openAdvancedSettings()
        networkLayerStep.check32BitFCnt()
        networkLayerStep.fillRx2DataRateIndex(device.rx2_data_rate_index)
        networkLayerStep.fillRx2Frequency(device.rx2_frequency)
        networkLayerStep.fillFactoryPresetFrequencies(device.factory_frequencies)
        networkLayerStep.selectPingSlotPeriodicity(device.ping_slot_periodicity)
        networkLayerStep.fillPingSlotFrequency(device.ping_slot_frequency)
        networkLayerStep.goToJoinSettingsStep()

        const joinSettingsStep = new JoinSettingsStep({ lorawanVersion: device.lorawan_version })
        joinSettingsStep.fillAppKey(device.app_key)
        joinSettingsStep.submit()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds using external Join Server', () => {
        const device = {
          id: 'otaa-test-mac-settings-ext-js',
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          name: 'Test OTAA MAC settings',
          description: 'Test MAC settings device',
          lorawan_version: 'MAC V1.0',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.0',
          app_key: generateHexValue(32),
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkOTAA()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.checkExternalJS()
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillAppEUI(device.app_eui)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.submit()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.0.2', () => {
      it('succeeds creating class A device', () => {
        const device = {
          id: 'otaa-test-uno',
          app_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          name: 'Test OTAA The Things Uno',
          description: 'The Things Uno test device',
          lorawan_version: 'MAC V1.0.2',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.0.2 REV B',
          app_key: generateHexValue(32),
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkOTAA()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillAppEUI(device.app_eui)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.goToJoinSettingsStep()

        const joinSettingsStep = new JoinSettingsStep({ lorawanVersion: device.lorawan_version })
        joinSettingsStep.fillAppKey(device.app_key)
        joinSettingsStep.submit()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.1', () => {
      it('succeeds creating class A device', () => {
        const device = {
          id: 'otaa-test-v1-1-class-a',
          join_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          name: 'Test LoRaWAN V1.1 OTAA device',
          description: 'Test LoRaWAN V1.1 OTAA device',
          lorawan_version: 'MAC V1.1',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.1 REV A',
          app_key: generateHexValue(32),
          nwk_key: generateHexValue(32),
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkOTAA()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillJoinEUI(device.join_eui)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.goToJoinSettingsStep()

        const joinSettingsStep = new JoinSettingsStep({ lorawanVersion: device.lorawan_version })
        joinSettingsStep.fillAppKey(device.app_key)
        joinSettingsStep.fillNwkKey(device.nwk_key)
        joinSettingsStep.submit()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds creating class B device', () => {
        const device = {
          id: 'otaa-test-v1-1-class-b',
          join_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          name: 'Test LoRaWAN V1.1 OTAA device',
          description: 'Test LoRaWAN V1.1 OTAA device',
          lorawan_version: 'MAC V1.1',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.1 REV A',
          app_key: generateHexValue(32),
          nwk_key: generateHexValue(32),
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkOTAA()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillJoinEUI(device.join_eui)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.checkClassB()
        networkLayerStep.goToJoinSettingsStep()

        const joinSettingsStep = new JoinSettingsStep({ lorawanVersion: device.lorawan_version })
        joinSettingsStep.fillAppKey(device.app_key)
        joinSettingsStep.fillNwkKey(device.nwk_key)
        joinSettingsStep.submit()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds creating class C device', () => {
        const device = {
          id: 'otaa-test-v1-1-class-c',
          join_eui: generateHexValue(16),
          dev_eui: generateHexValue(16),
          name: 'Test LoRaWAN V1.1 OTAA device',
          description: 'Test LoRaWAN V1.1 OTAA device',
          lorawan_version: 'MAC V1.1',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.1 REV A',
          app_key: generateHexValue(32),
          nwk_key: generateHexValue(32),
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkOTAA()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillJoinEUI(device.join_eui)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.checkClassC()
        networkLayerStep.goToJoinSettingsStep()

        const joinSettingsStep = new JoinSettingsStep({ lorawanVersion: device.lorawan_version })
        joinSettingsStep.fillAppKey(device.app_key)
        joinSettingsStep.fillNwkKey(device.nwk_key)
        joinSettingsStep.submit()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })
  })

  describe('ABP', () => {
    const application = {
      ids: { application_id: 'abp-test-application' },
    }
    const appId = application.ids.application_id

    before(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.createApplication(application, user.ids.user_id)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`)
    })

    describe('LoRaWAN V1.0.2', () => {
      it('succeeds creating class A device', () => {
        const device = {
          id: 'abp-test-uno',
          dev_eui: generateHexValue(16),
          dev_addr: generateHexValue(8),
          name: 'Test ABP The Things Uno',
          description: 'The Things Uno test device',
          lorawan_version: 'MAC V1.0.2',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.0.2 REV B',
          nwk_s_key: generateHexValue(32),
          rx2_data_rate_index: 3,
          factory_frequencies: [
            '868100000',
            '868300000',
            '868500000',
            '867100000',
            '867300000',
            '867500000',
            '867700000',
            '867900000',
          ],
          app_s_key: generateHexValue(32),
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkABP()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.fillDevAddress(device.dev_addr)
        networkLayerStep.fillNwkSKey(device.nwk_s_key)
        networkLayerStep.openAdvancedSettings()
        networkLayerStep.check32BitFCnt()
        networkLayerStep.fillRx2DataRateIndex(device.rx2_data_rate_index)
        networkLayerStep.fillFactoryPresetFrequencies(device.factory_frequencies)
        networkLayerStep.goToApplicationLayerStep()

        const applicationLayerStep = new ApplicationLayerStep()
        applicationLayerStep.fillAppSKey(device.app_s_key)
        applicationLayerStep.submit()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })

    describe('LoRaWAN V1.0.4', () => {
      it('succeeds creating class A device', () => {
        const device = {
          id: 'abp-test-v1-0-4',
          dev_eui: generateHexValue(16),
          dev_addr: generateHexValue(8),
          name: 'Test LoRaWAN V1.0.4 ABP device',
          description: 'Test LoRaWAN V1.0.4 ABP device',
          lorawan_version: 'MAC V1.0.4',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.0',
          nwk_s_key: generateHexValue(32),
          app_s_key: generateHexValue(32),
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkABP()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.fillDevAddress(device.dev_addr)
        networkLayerStep.fillNwkSKey(device.nwk_s_key)
        networkLayerStep.goToApplicationLayerStep()

        const applicationLayerStep = new ApplicationLayerStep()
        applicationLayerStep.fillAppSKey(device.app_s_key)
        applicationLayerStep.submit()

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
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.createApplication(application, user.ids.user_id)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`)
    })

    describe('LoRaWAN V1.0.4', () => {
      it('succeeds creating class B device', () => {
        const device = {
          id: 'multicast-test-class-b',
          dev_eui: generateHexValue(16),
          dev_addr: generateHexValue(8),
          name: 'Test v1.0.4 multicast class B device',
          description: 'Test v1.0.4 multicast class B device',
          lorawan_version: 'MAC V1.0.4',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.0',
          nwk_s_key: generateHexValue(32),
          app_s_key: generateHexValue(32),
          ping_slot_periodicity: 'PING_EVERY_2S',
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkMulticast()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.checkClassB()
        networkLayerStep.fillDevAddress(device.dev_addr)
        networkLayerStep.fillNwkSKey(device.nwk_s_key)
        networkLayerStep.selectPingSlotPeriodicity(device.ping_slot_periodicity)
        networkLayerStep.goToApplicationLayerStep()

        const applicationLayerStep = new ApplicationLayerStep()
        applicationLayerStep.fillAppSKey(device.app_s_key)
        applicationLayerStep.submit()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds creating class C device', () => {
        const device = {
          id: 'multicast-test-class-c',
          dev_eui: generateHexValue(16),
          dev_addr: generateHexValue(8),
          name: 'Test v1.0.4 multicast class C device',
          description: 'Test v1.0.4 multicast class C device',
          lorawan_version: 'MAC V1.0.4',
          frequency_plan_id: '863-870 MHz',
          lorawan_phy_version: 'PHY V1.0',
          nwk_s_key: generateHexValue(32),
          app_s_key: generateHexValue(32),
          ping_slot_periodicity: 'PING_EVERY_4S',
        }

        const configurationStep = new ConfigurationStep()
        configurationStep.checkMulticast()
        configurationStep.selectLorawanVersion(device.lorawan_version)
        configurationStep.submit()

        const basicSettingsStep = new BasicSettingsStep()
        basicSettingsStep.fillId(device.id)
        basicSettingsStep.fillDevEUI(device.dev_eui)
        basicSettingsStep.fillName(device.name)
        basicSettingsStep.fillDescription(device.description)
        basicSettingsStep.goToNetworkLayerStep()

        const networkLayerStep = new NetworkLayerStep()
        networkLayerStep.selectFrequencyPlan(device.frequency_plan_id)
        networkLayerStep.selectPhyVersion(device.lorawan_phy_version)
        networkLayerStep.checkClassC()
        networkLayerStep.fillDevAddress(device.dev_addr)
        networkLayerStep.fillNwkSKey(device.nwk_s_key)
        networkLayerStep.selectPingSlotPeriodicity(device.ping_slot_periodicity)
        networkLayerStep.goToApplicationLayerStep()

        const applicationLayerStep = new ApplicationLayerStep()
        applicationLayerStep.fillAppSKey(device.app_s_key)
        applicationLayerStep.submit()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device.id}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })
  })
})
