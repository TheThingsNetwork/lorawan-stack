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

class ConfigurationStep {
  checkOTAA() {
    cy.findByLabelText('Over the air activation (OTAA)')
      .should('exist')
      .check()
  }

  checkABP() {
    cy.findByLabelText('Activation by personalization (ABP)')
      .should('exist')
      .check()
  }

  checkMulticast() {
    cy.findByLabelText('Multicast')
      .should('exist')
      .check()
  }

  checkNone() {
    cy.findByLabelText('Do not configure activation')
      .should('exist')
      .check()
  }

  selectLorawanVersion(version) {
    cy.findByLabelText('LoRaWAN version').selectOption(version)
  }

  checkExternalJS() {
    cy.findByLabelText('Join Server address').then($input => {
      cy.wrap($input).should('not.be.disabled')
      cy.findByLabelText('External Join Server')
        .should('exist')
        .check()
      cy.wrap($input)
        .should('be.disabled')
        .and('have.attr', 'placeholder', 'External')
        .and('be.visible')
    })
  }

  submit() {
    cy.findByRole('button', { name: 'Start' })
      .should('be.visible')
      .click()
  }
}

class BasicSettingsStep {
  fillId(id) {
    cy.findByLabelText('End device ID').then($input => {
      cy.wrap($input)
        .should('be.visible')
        .and('have.attr', 'placeholder', 'my-new-device')
      cy.wrap($input).type(id)
    })
  }

  fillJoinEUI(eui) {
    cy.findDescriptionByLabelText('JoinEUI')
      .should(
        'contain',
        'The JoinEUI identifies the Join Server. If no JoinEUI is provided by the device manufacturer (usually for development), it can be filled with zeros.',
      )
      .should('be.visible')
    cy.findByLabelText('JoinEUI')
      .should('be.visible')
      .type(eui)
  }

  fillAppEUI(eui) {
    cy.findDescriptionByLabelText('AppEUI')
      .should(
        'contain',
        'The AppEUI uniquely identifies the owner of the end device. If no AppEUI is provided by the device manufacturer (usually for development), it can be filled with zeros.',
      )
      .should('be.visible')
    cy.findByLabelText('AppEUI')
      .should('be.visible')
      .type(eui)
  }

  fillDevEUI(eui) {
    cy.findDescriptionByLabelText('DevEUI')
      .should('contain', 'The DevEUI is the unique identifier for this end device')
      .should('be.visible')
    cy.findByLabelText('DevEUI')
      .should('be.visible')
      .type(eui)
  }

  fillName(name) {
    cy.findByLabelText('End device name').then($input => {
      cy.wrap($input)
        .should('be.visible')
        .and('have.attr', 'placeholder', 'My new end device')
      cy.wrap($input).type(name)
    })
  }

  fillDescription(description) {
    cy.findDescriptionByLabelText('End device description')
      .should(
        'contain',
        'Optional end device description; can also be used to save notes about the end device',
      )
      .should('be.visible')
    cy.findByLabelText('End device description')
      .should('be.visible')
      .type(description)
  }

  goToNetworkLayerStep() {
    cy.findByRole('button', { name: /Network layer settings/ })
      .should('be.visible')
      .click()
  }
}

class NetworkLayerStep {
  selectFrequencyPlan(plan) {
    cy.findByLabelText('Frequency plan').selectOption(plan)
  }

  selectPhyVersion(version) {
    cy.findByLabelText('Regional Parameters version').selectOption(version)
  }

  checkClassB() {
    cy.findByLabelText('Supports class B')
      .should('exist')
      .check()
  }

  checkClassC() {
    cy.findByLabelText('Supports class C')
      .should('exist')
      .check()
  }

  fillDevAddress(address) {
    cy.findDescriptionByLabelText('Device address')
      .should(
        'contain',
        'Device address, issued by the Network Server or chosen by device manufacturer in case of testing range',
      )
      .and('be.visible')
    cy.findByLabelText('Device address')
      .should('be.visible')
      .type(address)
  }

  fillNwkSKey(key) {
    cy.findDescriptionByLabelText('NwkSKey')
      .should('contain', 'Network session key')
      .and('be.visible')
    cy.findByLabelText('NwkSKey')
      .should('be.visible')
      .type(key)
  }

  openAdvancedSettings() {
    cy.get('[id="mac-settings"]').should('not.be.visible')
    cy.findByRole('heading', { name: /Advanced settings/ })
      .should('be.visible')
      .click()
  }

  closeAdvancedSettings() {
    cy.get('[id="mac-settings"]').should('be.visible')
    cy.findByRole('heading', { name: /Advanced settings/ })
      .should('be.visible')
      .click()
  }

  check16BitFCnt() {
    cy.findByLabelText('16 bit')
      .should('exist')
      .check()
  }

  check32BitFCnt() {
    cy.findByLabelText('32 bit')
      .should('exist')
      .check()
  }

  fillRx2DataRateIndex(index) {
    cy.findDescriptionByLabelText('RX2 Data Rate Index')
      .should('contain', 'The default RX2 data rate index value the device uses after a reset')
      .and('be.visible')
    cy.findByLabelText('RX2 Data Rate Index')
      .should('be.visible')
      .type(index)
  }

  fillRx2Frequency(frequency) {
    cy.findDescriptionByLabelText('RX2 Frequency')
      .should('contain', 'Frequency for RX2 (Hz)')
      .and('be.visible')
    cy.findByLabelText('RX2 Frequency').then($input => {
      cy.wrap($input)
        .should('have.attr', 'placeholder', 'e.g. 869525000 for 869,525 MHz')
        .and('be.visible')
      cy.wrap($input).type(frequency)
    })
  }

  fillFactoryPresetFrequencies(frequencies) {
    cy.findByRole('button', { name: /Add Frequency/ }).as('addFactoryPresetFreqBtn')
    for (const idx in frequencies) {
      cy.get('@addFactoryPresetFreqBtn').click()
      cy.get(`[name="mac_settings.factory_preset_frequencies[${idx}].value"]`).type(
        frequencies[idx],
      )
    }
  }

  selectPingSlotPeriodicity(periodicity) {
    cy.findByLabelText('Ping Slot Periodicity').selectOption(periodicity)
  }

  fillPingSlotFrequency(frequency) {
    cy.findDescriptionByLabelText('Ping Slot Frequency')
      .should('contain', 'Frequency of the class B ping slot (Hz)')
      .and('be.visible')
    cy.findByLabelText('Ping Slot Frequency').then($input => {
      cy.wrap($input).should('have.attr', 'placeholder', 'e.g. 869525000 for 869,525 MHz')
      cy.wrap($input).type(frequency)
    })
  }

  goToBasicSettingsStep() {
    cy.findByRole('button', { name: /Basic settings/ })
      .should('be.visible')
      .click()
  }

  goToJoinSettingsStep() {
    cy.findByRole('button', { name: /Join settings/ })
      .should('be.visible')
      .click()
  }

  goToApplicationLayerStep() {
    cy.findByRole('button', { name: /Application layer settings/ })
      .should('be.visible')
      .click()
  }

  submit() {
    cy.findByRole('button', { name: 'Add end device' })
      .should('be.visible')
      .click()
  }
}

class JoinSettingsStep {
  constructor({ lorawanVersion = '1.0.0' }) {
    this._lorawanVersion = parseInt(lorawanVersion.replace(/\D/g, '').padEnd(3, 0))
  }

  fillAppKey(key) {
    if (this._lorawanVersion >= 110) {
      cy.findDescriptionByLabelText('AppKey')
        .should(
          'contain',
          'The root key to derive the application session key to secure communication between the end device and the application',
        )
        .and('be.visible')
    } else {
      cy.findDescriptionByLabelText('AppKey')
        .should(
          'contain',
          'The root key to derive session keys to secure communication between the end device and the application',
        )
        .and('be.visible')
    }

    cy.findByLabelText('AppKey')
      .should('be.visible')
      .type(key)
  }

  fillNwkKey(key) {
    cy.findDescriptionByLabelText('NwkKey')
      .should(
        'contain',
        'The root key to derive network session keys to secure communication between the end device and the network',
      )
      .should('be.visible')
    cy.findByLabelText('NwkKey')
      .should('be.visible')
      .type(key)
  }

  openAdvancedSettings() {
    cy.get('[id="advanced-settings"]').should('not.be.visible')
    cy.findByRole('heading', { name: /Advanced settings/ })
      .should('be.visible')
      .click()
    cy.findByRole('heading', { name: 'MAC settings' }).should('be.visible')
  }

  closeAdvancedSettings() {
    cy.get('[id="advanced-settings"]').should('be.visible')
    cy.findByRole('heading', { name: /Advanced settings/ })
      .should('be.visible')
      .click()
  }

  goToNetworkLayerSettingsStep() {
    cy.findByRole('button', { name: /Network layer settings/ })
      .should('be.visible')
      .click()
  }

  submit() {
    cy.findByRole('button', { name: 'Add end device' })
      .should('be.visible')
      .click()
  }
}

class ApplicationLayerStep {
  checkSkipCrypto() {
    cy.findDescriptionByLabelText('Skip payload encryption and decryption')
      .should('contain', 'Skip decryption of uplink payloads and encryption of downlink payloads')
      .should('be.visible')
    cy.findByLabelText('Skip payload encryption and decryption')
      .should('exist')
      .check()
  }

  fillAppSKey(key) {
    cy.findDescriptionByLabelText('AppSKey')
      .should('contain', 'Application session key')
      .should('be.visible')
    cy.findByLabelText('AppSKey')
      .should('be.visible')
      .type(key)
  }

  submit() {
    cy.findByRole('button', { name: 'Add end device' })
      .should('be.visible')
      .click()
  }
}

export {
  ConfigurationStep,
  BasicSettingsStep,
  NetworkLayerStep,
  JoinSettingsStep,
  ApplicationLayerStep,
}
