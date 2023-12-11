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

import { generateCollaborator } from '../../../support/utils'

describe('Gateway general settings', () => {
  let user
  let gateway
  let gateway2
  const collabUserId = 'test-collab-user'
  const collabUser = {
    ids: { user_id: collabUserId },
    primary_email_address: 'test-collab-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    user = {
      ids: { user_id: 'gtw-settings-test-user' },
      primary_email_address: 'gtw-settings-test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
    cy.createUser(user)
    cy.createUser(collabUser)
    gateway = {
      ids: { gateway_id: 'test-gateway', eui: '0000000000000000' },
      name: 'Test Gateway',
      description: 'Gateway for testing gateway general settings',
      schedule_anytime_delay: '523ms',
      enforce_duty_cycle: true,
      gateway_server_address: 'localhost',
      attributes: {
        key: 'value',
      },
    }
    gateway2 = {
      ids: { gateway_id: 'test-gateway-frequency-plans', eui: '0000000000000001' },
      name: 'Test Gateway Frequency Plans',
      description: 'Gateway for testing multiple frequency plans',
      schedule_anytime_delay: '523ms',
      enforce_duty_cycle: true,
      gateway_server_address: 'localhost',
      attributes: {
        key: 'value',
      },
      frequency_plan_ids: ['EU_863_870', 'US_902_928_FSB_1'],
    }
    cy.createGateway(gateway, user.ids.user_id)
    cy.createGateway(gateway2, user.ids.user_id)
  })

  it('displays newly created gateway values', () => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(
      `${Cypress.config('consoleRootPath')}/gateways/${gateway.ids.gateway_id}/general-settings`,
    )
    cy.findByRole('heading', { name: 'Basic settings' }).should('be.visible')
    cy.findByLabelText('Gateway ID')
      .should('be.disabled')
      .and('have.attr', 'value')
      .and('eq', gateway.ids.gateway_id)
    cy.findByLabelText('Gateway EUI')
      .should('not.be.disabled')
      .and('have.attr', 'value')
      .and('eq', gateway.ids.eui)
    cy.findByLabelText('Gateway name')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', gateway.name)
    cy.findByLabelText('Gateway description')
      .should('be.visible')
      .and('have.text', gateway.description)
    cy.findDescriptionByLabelText('Gateway description')
      .should(
        'contain',
        'Optional gateway description; can also be used to save notes about the gateway',
      )
      .and('be.visible')
    cy.findByLabelText('Gateway Server address')
      .should('be.visible')
      .and('have.attr', 'value', gateway.gateway_server_address)
    cy.findDescriptionByLabelText('Gateway Server address')
      .should('contain', 'The address of the Gateway Server to connect to')
      .and('be.visible')
    cy.findByLabelText('Require authenticated connection')
      .should('exist')
      .and('have.attr', 'value', 'false')
    cy.findDescriptionByLabelText('Require authenticated connection')
      .should(
        'contain',
        'Controls whether this gateway may only connect if it uses an authenticated Basic Station or MQTT connection',
      )
      .and('be.visible')
    cy.findByLabelText('Gateway status').should('exist').and('have.attr', 'value', 'false')
    cy.findDescriptionByLabelText('Gateway status')
      .should('contain', 'The status of this gateway may be visible to other users')
      .and('be.visible')
    cy.findDescriptionByLabelText('Gateway location').should('contain', 'public').and('be.visible')
    cy.findByTestId('key-value-map').within(() => {
      cy.findByTestId('attributes[0].key').should('be.visible').and('have.attr', 'value', 'key')
      cy.findByTestId('attributes[0].value')
        .should('be.visible')
        .and('have.attr', 'value', gateway.attributes.key)
    })
    cy.findByLabelText('Automatic updates').should('exist').and('have.attr', 'value', 'false')
    cy.findDescriptionByLabelText('LoRa Basics Station LNS Authentication Key')
      .should(
        'contain',
        'The Authentication Key for Lora Basics Station LNS connections. This field is ignored for other gateways.',
      )
      .and('be.visible')
    cy.findDescriptionByLabelText('Automatic updates')
      .should('contain', 'Gateway can be updated automatically')
      .and('be.visible')
    cy.findByLabelText('Channel')
      .should('be.visible')
      .and('have.attr', 'placeholder')
      .and('eq', 'Stable')
    cy.findDescriptionByLabelText('Channel')
      .should('contain', 'Channel for gateway automatic updates')
      .and('be.visible')
    cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
    cy.findByRole('button', { name: /Delete gateway/ }).should('be.visible')
    cy.findByRole('heading', { name: 'LoRaWAN options' }).should('be.visible')
    cy.findByText('Frequency plan').should('not.exist')
    cy.findByRole('button', { name: 'Expand' }).click()
    cy.findByText('Frequency plan').should('be.visible')
    cy.findByLabelText(/Enforce duty cycle/)
      .should('exist')
      .and('have.attr', 'value', 'true')
    cy.findDescriptionByLabelText(/Enforce duty cycle/).should('be.visible')
    cy.findByTestId('schedule_anytime_delay')
      .should('be.visible')
      .and('have.attr', 'value', '0.523')
  })

  it('succeeds changing gateway information', () => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(
      `${Cypress.config('consoleRootPath')}/gateways/${gateway.ids.gateway_id}/general-settings`,
    )

    const newGatewayName = 'New Gateway Name'
    const newGatewayDesc = 'New Gateway Desc'
    const newFrequencyPlan = 'Europe 863-870 MHz (SF12 for RX2)'
    const address = 'otherhost'
    const lnsKey = '1234'

    cy.findByLabelText('Gateway name').clear()
    cy.findByLabelText('Gateway name').type(newGatewayName)
    cy.findByLabelText('Gateway description').clear()
    cy.findByLabelText('Gateway description').type(newGatewayDesc)
    cy.findByLabelText('Gateway Server address').clear()
    cy.findByLabelText('Gateway Server address').type(address)
    cy.findByLabelText('Require authenticated connection').check()
    cy.findByLabelText('LoRa Basics Station LNS Authentication Key').type(lnsKey)
    cy.findByLabelText('Gateway status').check()
    cy.findByLabelText('Gateway location').check()
    cy.findByPlaceholderText('key').type('-changed')
    cy.findByPlaceholderText('value').type('-changed')
    cy.findByLabelText('Automatic updates').check()
    cy.findByLabelText('Channel').type('test')
    cy.findByLabelText('Packet Broker').check()

    cy.findByLabelText('Require authenticated connection').check()
    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText('Gateway updated').should('be.visible')
    cy.reload()

    cy.findByLabelText('Gateway name').should('have.value', newGatewayName)
    cy.findByLabelText('Gateway description').should('have.value', newGatewayDesc)
    cy.findByLabelText('Gateway Server address').should('have.value', address)
    cy.findByLabelText('Require authenticated connection').should('have.attr', 'checked')
    cy.findByLabelText('LoRa Basics Station LNS Authentication Key')
      .should('have.attr', 'value')
      .and('eq', lnsKey)
    cy.findByLabelText('Gateway status').should('have.attr', 'checked')
    cy.findByLabelText('Gateway location').should('have.attr', 'checked')
    cy.findByPlaceholderText('key').should('have.value', 'key-changed')
    cy.findByPlaceholderText('value').should('have.value', 'value-changed')
    cy.findByLabelText('Automatic updates').should('have.attr', 'checked')
    cy.findByLabelText('Channel').should('have.value', 'test')
    cy.findByLabelText('Packet Broker').should('have.attr', 'checked')

    cy.findByText('LoRaWAN options', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByLabelText('Schedule downlink late').check()
        cy.findByLabelText(/Enforce duty cycle/).uncheck()
        cy.findByLabelText('Schedule any time delay').clear()
        cy.findByLabelText('Schedule any time delay').type('1')
        cy.findByText('Frequency plan')
          .parents('div[data-test-id="form-field"]')
          .find('input')
          .first()
          .selectOption(newFrequencyPlan)
        cy.findByRole('button', { name: 'Save changes' }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText('Gateway updated').should('be.visible')
    cy.reload()

    cy.findByText('LoRaWAN options', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByText(newFrequencyPlan)
        cy.findByLabelText('Schedule downlink late').should('have.attr', 'checked')
        cy.findByLabelText(/Enforce duty cycle/).should('not.have.attr', 'checked')
        cy.findByLabelText('Schedule any time delay').should('have.value', '1')
      })
  })

  it('fails adding non-collaborator contact information', () => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    const entity = 'gateways'
    const userCollaborator = generateCollaborator(entity, 'user')
    cy.createCollaborator(entity, gateway.ids.gateway_id, userCollaborator)

    cy.visit(
      `${Cypress.config('consoleRootPath')}/gateways/${gateway.ids.gateway_id}/general-settings`,
    )

    cy.findByText('Contact information').should('be.visible')
    cy.findByLabelText('Administrative contact').clear()
    cy.findByLabelText('Administrative contact').type('test-non-collab-user')
    cy.findByText('No matching user or organization was found')
  })

  it('succeeds adding contact information', () => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    const entity = 'gateways'
    const userCollaborator = generateCollaborator(entity, 'user')
    cy.createCollaborator(entity, gateway.ids.gateway_id, userCollaborator)

    cy.visit(
      `${Cypress.config('consoleRootPath')}/gateways/${gateway.ids.gateway_id}/general-settings`,
    )

    cy.findByText('Contact information').should('be.visible')
    cy.findByLabelText('Administrative contact').clear()
    cy.findByLabelText('Administrative contact').selectOption(collabUserId)

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText(`Gateway updated`).should('be.visible')
  })

  it('succeeds setting current user as contact', () => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    const entity = 'gateways'
    const userCollaborator = generateCollaborator(entity, 'user')
    cy.createCollaborator(entity, gateway.ids.gateway_id, userCollaborator)

    cy.intercept('GET', `/api/v3/is/configuration`, { fixture: 'restricted-user-config.json' })
    cy.visit(
      `${Cypress.config('consoleRootPath')}/gateways/${gateway.ids.gateway_id}/general-settings`,
    )

    cy.findByText('Contact information').should('be.visible')
    cy.findByLabelText('Administrative contact').should('have.attr', 'disabled')
    cy.findByLabelText('Administrative contact')
      .parent()
      .parent()
      .within(() => {
        cy.findByText(collabUserId).should('be.visible')
      })
    cy.findByRole('button', { name: /Set yourself as administrative contact/ }).click()
    cy.findByLabelText('Administrative contact')
      .parent()
      .parent()
      .within(() => {
        cy.findByText(user.ids.user_id).should('be.visible')
      })
  })

  it('succeeds editing multiple frequency plans', () => {
    const newFrequencyPlan = 'Europe 863-870 MHz, 6 channels for roaming (Draft)'
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(
      `${Cypress.config('consoleRootPath')}/gateways/${gateway2.ids.gateway_id}/general-settings`,
    )

    cy.findByText('LoRaWAN options', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByText('Frequency plan')
          .parents('div[data-test-id="form-field"]')
          .find('input')
          .first()
          .selectOption(newFrequencyPlan)
        cy.findByRole('button', { name: 'Save changes' }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText('Gateway updated').should('be.visible')
    cy.reload()

    cy.findByText('LoRaWAN options', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByText('Frequency plan')
        cy.findByText(newFrequencyPlan)
      })
  })

  it('succeeds deleting the gateway', () => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(
      `${Cypress.config('consoleRootPath')}/gateways/${gateway.ids.gateway_id}/general-settings`,
    )
    cy.findByRole('button', { name: /Delete gateway/ }).click()
    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Confirm deletion', { selector: 'h1' }).should('be.visible')

        cy.findByRole('button', { name: /Cancel/ }).should('be.visible')
        cy.get('input').type(gateway.ids.gateway_id)
        cy.findByRole('button', { name: /Delete gateway/ })
          .should('be.visible')
          .click()
      })

    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/gateways`)

    cy.findByRole('cell', { name: gateway.ids.gateway_id }).should('not.exist')
  })
})
