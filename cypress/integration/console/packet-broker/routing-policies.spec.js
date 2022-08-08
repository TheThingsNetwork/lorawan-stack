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

describe('Packet Broker routing policies', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  beforeEach(() => {
    cy.fixture('console/packet-broker/policies-home-network.json').as('homeNetworkPolicies')

    cy.intercept('/api/v3/pba/info', { fixture: 'console/packet-broker/info-registered.json' })
    cy.intercept('/api/v3/pba/networks*', { fixture: 'console/packet-broker/networks.json' })
    cy.intercept('/api/v3/pba/forwarders/policies*', {
      fixture: 'console/packet-broker/policies-forwarder.json',
    })
    cy.intercept('GET', '/api/v3/pba/home-networks/gateway-visibilities/default', {
      statusCode: 404,
    })

    cy.loginConsole({ user_id: 'admin', password: 'admin' })
  })

  it('succeeds setting a default routing policy', () => {
    cy.intercept('GET', '/api/v3/pba/home-networks/policies/default', { statusCode: 404 })
    cy.intercept('PUT', '/api/v3/pba/home-networks/policies/default', {})
    cy.visit(`${Cypress.config('consoleRootPath')}/admin/packet-broker`)

    cy.findByLabelText('Use default routing policy for this network').check()

    // Check routing policy form checkboxes.
    cy.findByText('Uplink')
      .parent()
      .within(() => {
        cy.findByLabelText('Join request').check()
        cy.findByLabelText('MAC data').check()
        cy.findByLabelText('Application data').check()
        cy.findByLabelText('Signal quality information').check()
        cy.findByLabelText('Localization information').check()
      })
    cy.findByText('Downlink')
      .parent()
      .within(() => {
        cy.findByLabelText('Join accept').check()
        cy.findByLabelText('MAC data').check()
        cy.findByLabelText('Application data').check()
      })
    cy.findByRole('button', { name: 'Save default policy' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText('Default routing policy set')
      .should('be.visible')
  })

  it('succeeds unsetting a default routing policy', () => {
    cy.intercept('GET', '/api/v3/pba/home-networks/policies/default', {
      fixture: 'console/packet-broker/default-policy.json',
    })
    cy.intercept('DELETE', '/api/v3/pba/home-networks/policies/default', {})

    cy.visit(`${Cypress.config('consoleRootPath')}/admin/packet-broker`)

    cy.findByLabelText('Do not use a default routing policy for this network').check()
    cy.findByRole('button', { name: 'Save default policy' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText('Default routing policy set')
      .should('be.visible')
  })

  it('succeeds setting individual per-network routing policy', () => {
    cy.intercept('GET', '/api/v3/pba/home-networks/policies/default', {
      fixture: 'console/packet-broker/default-policy.json',
    })
    cy.intercept('GET', '/api/v3/pba/home-networks/policies/19', {
      statusCode: 404,
      fixture: '404-body.json',
    })
    cy.intercept('PUT', '/api/v3/pba/home-networks/policies/19', {})

    cy.visit(`${Cypress.config('consoleRootPath')}/admin/packet-broker/networks/19`)

    // Check routing policy form checkboxes.
    cy.findByLabelText('Use network specific routing policy').check()
    cy.findByText('Set routing policy towards this network')
      .parent()
      .within(() => {
        cy.findByText('Uplink')
          .parent()
          .within(() => {
            cy.findByLabelText('Join request').check()
            cy.findByLabelText('MAC data').check()
            cy.findByLabelText('Application data').check()
            cy.findByLabelText('Signal quality information').check()
            cy.findByLabelText('Localization information').check()
          })
        cy.findByText('Downlink')
          .parent()
          .within(() => {
            cy.findByLabelText('Join accept').check()
            cy.findByLabelText('MAC data').check()
            cy.findByLabelText('Application data').check()
          })
        cy.findByRole('button', { name: 'Save routing policy' }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText('Routing policy set')
      .should('be.visible')
  })

  it('succeeds unsetting individual per-network routing policy (without default policy)', function () {
    cy.intercept(
      'GET',
      '/api/v3/pba/home-networks/policies/19',
      this.homeNetworkPolicies.policies[1],
    )
    cy.intercept('GET', '/api/v3/pba/home-networks/policies/default', {
      statusCode: 404,
      fixture: '404-body.json',
    })
    cy.intercept('DELETE', '/api/v3/pba/home-networks/policies/19', {})

    cy.visit(`${Cypress.config('consoleRootPath')}/admin/packet-broker/networks/19`)

    cy.findByLabelText('Do not use a routing policy for this network').check()
    cy.findByRole('button', { name: 'Save routing policy' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText('Routing policy set')
      .should('be.visible')
  })

  it('succeeds unsetting individual per-network routing policy (with default policy)', function () {
    cy.intercept(
      'GET',
      '/api/v3/pba/home-networks/policies/19',
      this.homeNetworkPolicies.policies[1],
    )
    cy.intercept('GET', '/api/v3/pba/home-networks/policies/default', {
      fixture: 'console/packet-broker/default-policy.json',
    })
    cy.intercept('DELETE', '/api/v3/pba/home-networks/policies/19', {})

    cy.visit(`${Cypress.config('consoleRootPath')}/admin/packet-broker/networks/19`)

    cy.findByLabelText('Use default routing policy for this network').check()
    cy.findByRole('button', { name: 'Save routing policy' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText('Routing policy set')
      .should('be.visible')
  })
})
