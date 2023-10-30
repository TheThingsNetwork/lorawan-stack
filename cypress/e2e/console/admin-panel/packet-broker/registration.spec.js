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

describe('Packet Broker registration', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  it('redirects back if user has no admin rights', () => {
    const user = {
      ids: { user_id: 'packet-broker-test-user' },
      primary_email_address: 'packet-broker-test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
    cy.createUser(user)

    cy.intercept('/api/v3/pba/networks*', { fixture: 'console/packet-broker/networks.json' })
    cy.intercept('/api/v3/pba/home-networks/policies*', {
      fixture: 'console/packet-broker/policies-home-network.json',
    })
    cy.intercept('/api/v3/pba/forwarders/policies*', {
      fixture: 'console/packet-broker/policies-forwarder.json',
    })

    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })

    cy.visit(`${Cypress.config('consoleRootPath')}/admin-panel/packet-broker`)
    cy.location('pathname').should('eq', Cypress.config('consoleRootPath'))

    cy.visit(`${Cypress.config('consoleRootPath')}/admin-panel/packet-broker/networks/19`)
    cy.location('pathname').should('eq', Cypress.config('consoleRootPath'))
  })

  it('displays a notification when Packet Broker is not set up', () => {
    cy.intercept('/api/v3/pba/info', { statusCode: 403 })

    cy.loginConsole({ user_id: 'admin', password: 'admin' })
    cy.visit(`${Cypress.config('consoleRootPath')}/admin-panel/packet-broker`)

    cy.findByTestId('notification')
      .findByText(/not set up/)
      .should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')
  })

  it('displays UI elements in place', () => {
    cy.intercept('/api/v3/pba/info', { fixture: 'console/packet-broker/info.json' })
    cy.loginConsole({ user_id: 'admin', password: 'admin' })
    cy.visit(`${Cypress.config('consoleRootPath')}/admin-panel/packet-broker`)

    cy.findByText('Packet Broker', { selector: 'h1' }).should('be.visible')
    cy.findByText(/Packet Broker is a service by The Things Industries/).should('be.visible')
    cy.findByText('Packet Broker', { selector: 'a' }).should('be.visible')
    cy.findByText('Packet Broker website', { selector: 'a' }).should('be.visible')
    cy.findByText('Enable Packet Broker', { selector: 'span' }).should('be.visible')
    cy.findByTestId('switch')
      .should('be.visible')
      .and('not.be.checked')
      .and('not.have.attr', 'disabled')
    cy.findByText(/Enabling will allow/).should('be.visible')

    cy.findByText('Default routing policy').should('not.exist')
    cy.findByText('Networks').should('not.exist')

    cy.findByTestId('notification').should('not.exist')
    cy.findByTestId('error-notification').should('not.exist')
  })

  it('succeeds registering with Packet Broker', () => {
    cy.intercept('GET', '/api/v3/pba/home-networks/gateway-visibilities/default', {
      statusCode: 404,
    })
    cy.intercept('GET', '/api/v3/pba/home-networks/policies/default', { statusCode: 404 })
    cy.intercept('/api/v3/pba/registration', { fixture: 'console/packet-broker/registration.json' })
    cy.intercept('/api/v3/pba/info', { fixture: 'console/packet-broker/info.json' })

    cy.loginConsole({ user_id: 'admin', password: 'admin' })
    cy.visit(`${Cypress.config('consoleRootPath')}/admin-panel/packet-broker`)

    cy.findByText('Enable Packet Broker').click()
    cy.findByText('Enable Packet Broker').next().findByTestId('switch').should('be.checked')
    cy.findByText('List my network in Packet Broker publicly')
      .should('be.visible')
      .next()
      .findByTestId('switch')
      .should('be.checked')
    cy.findByLabelText('Forward traffic to all networks registered in Packet Broker').should(
      'exist',
    )
    cy.findByLabelText(
      'Forward traffic to The Things Stack Sandbox (community network) only',
    ).should('exist')
    cy.findByLabelText('Use custom routing policies').should('exist')

    cy.findByTestId('error-notification').should('not.exist')
  })
})
