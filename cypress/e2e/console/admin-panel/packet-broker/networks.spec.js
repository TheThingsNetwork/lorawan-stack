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

describe('Packet Broker networks', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  beforeEach(() => {
    cy.fixture('console/packet-broker/policies-forwarder.json').as('forwarderPolicies')
    cy.fixture('console/packet-broker/networks.json').as('networks')
    cy.fixture('console/packet-broker/default-policy.json').as('defaultPolicy')

    cy.intercept('/api/v3/pba/info', { fixture: 'console/packet-broker/info-registered.json' })
    cy.intercept('/api/v3/pba/networks*', { fixture: 'console/packet-broker/networks.json' })
    cy.intercept('/api/v3/pba/forwarders/policies*', {
      fixture: 'console/packet-broker/policies-forwarder.json',
    })
    cy.intercept('GET', '/api/v3/pba/home-networks/policies/default', {
      fixture: 'console/packet-broker/default-policy.json',
    })
    cy.intercept('GET', '/api/v3/pba/home-networks/gateway-visibilities/default', {
      statusCode: 404,
    })
    cy.intercept('GET', '/api/v3/pba/home-networks/policies/19/johan', {
      statusCode: 404,
      fixture: '404-body.json',
    })

    cy.loginConsole({ user_id: 'admin', password: 'admin' })
  })

  it('displays the network table correctly', function () {
    cy.intercept('/api/v3/pba/home-networks/policies*', {
      fixture: 'console/packet-broker/policies-home-network.json',
    })
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/admin-panel/packet-broker/routing-configuration/networks`,
    )
    cy.findByLabelText('Use custom routing policies').check()

    const { networks } = this.networks
    const networksFiltered = networks.filter(
      n => !(n.id.net_id === 19 && n.id.tenant_id === 'packet-broker-test'),
    )
    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', networksFiltered.length)
      networksFiltered.forEach(n => {
        cy.findAllByText(n.id.net_id.toString(16).padStart(6, '0'))
        if (n.id.tenant_id) {
          cy.findByText(n.id.tenant_id)
        }
        if (n.name) {
          cy.findByText(n.name)
        }
      })
    })
  })

  it('displays single networks correctly', function () {
    const { networks } = this.networks
    const network = networks.find(n => n.id.net_id === 19 && n.id.tenant_id === 'johan')
    const { policies } = this.forwarderPolicies
    const { uplink, downlink } = policies.find(
      n => n.forwarder_id.net_id === 19 && n.forwarder_id.tenant_id === 'johan',
    )

    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/admin-panel/packet-broker/routing-configuration/networks/19/johan`,
    )

    cy.findAllByText(`${network.id.net_id.toString(16).padStart(6, '0')}/${network.id.tenant_id}`)
    cy.findByText(
      `${network.dev_addr_blocks[0].dev_addr_prefix.dev_addr}/${network.dev_addr_blocks[0].dev_addr_prefix.length}`,
    )

    cy.findByText("This network's routing policy towards us")
      .siblings('[data-test-id="routing-policy-sheet"]')
      .within(() => {
        cy.findByText('Uplink')
          .parent()
          .within(() => {
            cy.findByText('Join request').should(
              'have.attr',
              'data-enabled',
              (uplink.join_request || false).toString(),
            )
            cy.findByText('MAC data').should(
              'have.attr',
              'data-enabled',
              (uplink.mac_data || false).toString(),
            )
            cy.findByText('Application data').should(
              'have.attr',
              'data-enabled',
              (uplink.application_data || false).toString(),
            )
            cy.findByText('Signal quality information').should(
              'have.attr',
              'data-enabled',
              (uplink.signal_quality || false).toString(),
            )
            cy.findByText('Localization information').should(
              'have.attr',
              'data-enabled',
              (uplink.localization || false).toString(),
            )
          })
        cy.findByText('Downlink')
          .parent()
          .within(() => {
            cy.findByText('Join accept').should(
              'have.attr',
              'data-enabled',
              (downlink.join_accept || false).toString(),
            )
            cy.findByText('MAC data').should(
              'have.attr',
              'data-enabled',
              (downlink.mac_data || false).toString(),
            )
            cy.findByText('Application data').should(
              'have.attr',
              'data-enabled',
              (downlink.application_data || false).toString(),
            )
          })
      })

    const { uplink: uplink2, downlink: downlink2 } = this.defaultPolicy
    cy.findByText('Set routing policy towards this network')
      .parent()
      .find('[data-test-id="routing-policy-sheet"]')
      .within(() => {
        cy.findByText('Uplink')
          .parent()
          .within(() => {
            cy.findByText('Join request').should(
              'have.attr',
              'data-enabled',
              (uplink2.join_request || false).toString(),
            )
            cy.findByText('MAC data').should(
              'have.attr',
              'data-enabled',
              (uplink2.mac_data || false).toString(),
            )
            cy.findByText('Application data').should(
              'have.attr',
              'data-enabled',
              (uplink2.application_data || false).toString(),
            )
            cy.findByText('Signal quality information').should(
              'have.attr',
              'data-enabled',
              (uplink2.signal_quality || false).toString(),
            )
            cy.findByText('Localization information').should(
              'have.attr',
              'data-enabled',
              (uplink2.localization || false).toString(),
            )
          })
        cy.findByText('Downlink')
          .parent()
          .within(() => {
            cy.findByText('Join accept').should(
              'have.attr',
              'data-enabled',
              (downlink2.join_accept || false).toString(),
            )
            cy.findByText('MAC data').should(
              'have.attr',
              'data-enabled',
              (downlink2.mac_data || false).toString(),
            )
            cy.findByText('Application data').should(
              'have.attr',
              'data-enabled',
              (downlink2.application_data || false).toString(),
            )
          })
      })
  })
})
