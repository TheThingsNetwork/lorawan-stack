// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

describe('Packet Broker routing policies', () => {
  before(() => {
    cy.dropAndSeedDatabase()
    cy.loginConsole({ user_id: 'admin', password: 'admin' })
  })

  it('succeeds showing UI elements in place', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/admin/packet-broker/default-gateway-visibility`)

    cy.findByLabelText('Location').should('be.visible')
    cy.findByLabelText('Antenna placement').should('be.visible')
    cy.findByLabelText('Antenna count').should('be.visible')
    cy.findByLabelText('Fine timestamps').should('be.visible')
    cy.findByLabelText('Contact information').should('be.visible')
    cy.findByLabelText('Status').should('be.visible')
    cy.findByLabelText('Frequecy plan').should('be.visible')
    cy.findByLabelText('Packet rates').should('be.visible')

    cy.findByRole('button', { name: 'Save default gateway visibility' }).should('be.visible')
  })

  it('succeeds setting default gateway visibility configuration', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/admin/packet-broker/default-gateway-visibility`)

    cy.findByLabelText('Location').check()
    cy.findByLabelText('Antenna placement').check()
    cy.findByLabelText('Antenna count').check()
    cy.findByLabelText('Fine timestamps').check()

    cy.findByRole('button', { name: 'Save default gateway visibility' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .findByText('Default gateway visibility set')
      .should('be.visible')

    cy.findByLabelText('Location').should('have.attr', 'checked')
    cy.findByLabelText('Antenna placement').should('have.attr', 'checked')
    cy.findByLabelText('Antenna count').should('have.attr', 'checked')
    cy.findByLabelText('Fine timestamps').should('have.attr', 'checked')
  })

  it('succeeds unsetting default gateway visibility configuration', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/admin/packet-broker/default-gateway-visibility`)

    cy.findByLabelText('Location').uncheck()
    cy.findByLabelText('Antenna placement').uncheck()
    cy.findByLabelText('Antenna count').uncheck()
    cy.findByLabelText('Fine timestamps').uncheck()

    cy.findByRole('button', { name: 'Save default gateway visibility' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .findByText('Default gateway visibility set')
      .should('be.visible')

    cy.findByLabelText('Location').should('not.have.attr', 'checked')
    cy.findByLabelText('Antenna placement').should('not.have.attr', 'checked')
    cy.findByLabelText('Antenna count').should('not.have.attr', 'checked')
    cy.findByLabelText('Fine timestamps').should('not.have.attr', 'checked')
  })
})
