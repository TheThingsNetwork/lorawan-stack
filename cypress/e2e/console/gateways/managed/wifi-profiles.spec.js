// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

describe('Managed Gateway WiFi profiles', () => {
  const generateUUID = () =>
    'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
      const r = (Math.random() * 16) | 0,
        v = c === 'x' ? r : (r & 0x3) | 0x8
      return v.toString(16)
    })

  const userId = 'managed-gateway-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'managed-gateway-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const organizationId = 'test-organization'
  const organization = { ids: { organization_id: organizationId }, name: 'Test organization' }

  const gatewayId = 'test-managed-gateway'
  const gateway = { ids: { gateway_id: gatewayId } }

  const gatewayVersionIds = {
    hardware_version: 'v1.1',
    firmware_version: 'v1.1',
    model_id: 'Managed gateway',
  }

  const collaborator = {
    collaborator: {
      ids: {
        organization_ids: {
          organization_id: organizationId,
        },
      },
      rights: ['RIGHT_GATEWAY_ALL'],
    },
  }

  const profiles = [
    {
      profile_id: generateUUID(),
      profile_name: 'Test profile1',
      shared: true,
      ssid: 'AccessPoint1',
    },
    {
      profile_id: generateUUID(),
      profile_name: 'Test profile2',
      shared: true,
      ssid: 'AccessPoint2',
      password: 'ABCDefg123!',
    },
    {
      profile_name: 'Test profile3',
      profile_id: generateUUID(),
      shared: true,
      ssid: 'AccessPoint1',
      network_interface_addresses: {
        ip_addresses: ['198.168.100.5'],
        subnet_mask: '255.255.255.0',
        gateway: '198.168.255.10',
        dns_servers: ['198.168.100.5'],
      },
    },
  ]

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createGateway(gateway, userId)
    cy.createOrganization(organization, userId)
    cy.createCollaborator('gateways', gatewayId, collaborator)
  })

  beforeEach(() => {
    cy.intercept('GET', `/api/v3/gcs/gateways/profiles/wifi/users/${userId}`, {
      statusCode: 200,
      body: {
        profiles,
      },
    })

    cy.intercept('POST', `/api/v3/gcs/gateways/profiles/wifi/users/${userId}`, {
      statusCode: 200,
    }).as('create-profile')

    cy.intercept('GET', `/api/v3/gcs/gateways/profiles/wifi/organizations/${organizationId}`, {
      statusCode: 200,
      body: {
        profiles,
      },
    })

    cy.intercept('POST', `/api/v3/gcs/gateways/managed/${gatewayId}/wifi/scan`, {
      statusCode: 200,
      body: {
        access_points: [
          {
            ssid: 'AccessPoint1',
            bssid: 'EC656E000100',
            channel: 0,
            authentication_mode: 'open',
            rssi: -50,
          },
          {
            ssid: 'AccessPoint2',
            bssid: 'EC656E000101',
            channel: 0,
            authentication_mode: 'secured',
            rssi: -70,
          },
        ],
      },
    }).as('scan-access-points')
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}`)
    cy.intercept('GET', `/api/v3/gcs/gateways/managed/${gatewayId}*`, {
      statusCode: 200,
      body: {
        ids: {
          gateway_id: `eui-${gateway.eui}`,
          eui: gateway.eui,
        },
        version_ids: gatewayVersionIds,
      },
    }).as('get-is-gtw-managed')
    cy.wait('@get-is-gtw-managed')
    cy.findByRole('heading', { name: 'test-managed-gateway' })
    cy.get('button').contains('Managed gateway').click()
    cy.get('a').contains('WiFi profiles').click()
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
    )
    cy.findByLabelText('Show profiles of').should('be.visible').selectOption(userId)
    cy.findByTestId('error-notification').should('not.exist')
  })

  it('succeeds to display UI elements in place', () => {
    cy.findByText('WiFi profiles', { selector: 'h1' }).should('be.visible')
    cy.contains('button', 'Add WiFi profile').should('be.visible')
    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 3)
    })
  })

  describe('when creating a WiFi profile', () => {
    beforeEach(() => {
      cy.contains('button', 'Add WiFi profile').click()
      cy.findByText('Add WiFi profile', { selector: 'h1' }).should('be.visible')
    })

    it('succeeds to validate WiFi profile fields', () => {
      cy.findByLabelText(/Use default network interface settings/).uncheck()

      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.get('#profile_name-field-error').should('be.visible')
      cy.get('#_access_point-field-error').should('be.visible')
      cy.get('#network_interface_addresses\\.ip_addresses-field-error').should('be.visible')
      cy.get('#network_interface_addresses\\.subnet_mask-field-error').should('be.visible')
      cy.get('#network_interface_addresses\\.gateway-field-error').should('be.visible')
      cy.findByText('AccessPoint2').click()
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.get('#password-field-error').should('be.visible')
      cy.findByText('Other...').click()
      cy.get('#ssid-field-error').should('be.visible')
    })

    it('succeeds to create WiFi profile with open access point and default network settings', () => {
      const expectedRequest = {
        profile: {
          profile_name: 'Open WiFi profile',
          profile_id: '',
          shared: true,
          ssid: 'AccessPoint1',
        },
      }
      cy.findByLabelText('Profile name').type('Open WiFi profile')
      cy.findByText('AccessPoint1').click()
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.wait('@create-profile').its('request.body').should('deep.equal', expectedRequest)
      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile created')
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
      )
    })

    it('succeeds to create WiFi profile with secured access point', () => {
      const expectedRequest = {
        profile: {
          profile_name: 'Secured WiFi profile',
          profile_id: '',
          shared: true,
          ssid: 'AccessPoint2',
          password: 'ABCDefg123!',
        },
      }
      cy.findByLabelText('Profile name').type('Secured WiFi profile')
      cy.findByText('AccessPoint2').click()
      cy.findByLabelText('WiFi password').type('ABCDefg123!')
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.wait('@create-profile').its('request.body').should('deep.equal', expectedRequest)
      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile created')
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
      )
    })

    it('succeeds to create WiFi profile with other access point', () => {
      const expectedRequest = {
        profile: {
          profile_name: 'Other WiFi profile',
          profile_id: '',
          shared: true,
          ssid: 'AccessPoint3',
        },
      }
      cy.findByLabelText('Profile name').type('Other WiFi profile')
      cy.findByText('Other...').click()
      cy.findByLabelText('SSID').type('AccessPoint3')
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.wait('@create-profile').its('request.body').should('deep.equal', expectedRequest)
      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile created')
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
      )
    })

    it('succeeds to create WiFi profile with custom network settings', () => {
      const expectedRequest = {
        profile: {
          profile_name: 'Custom WiFi profile',
          profile_id: '',
          shared: true,
          ssid: 'AccessPoint1',
          network_interface_addresses: {
            ip_addresses: ['198.168.100.5'],
            subnet_mask: '255.255.255.0',
            gateway: '198.168.255.10',
            dns_servers: ['198.168.100.5'],
          },
        },
      }
      cy.findByLabelText('Profile name').type('Custom WiFi profile')
      cy.findByText('AccessPoint1').click()
      cy.findByLabelText(/Use default network interface settings/).uncheck()
      cy.findByText('IP addresses')
        .parents('div[data-test-id="form-field"]')
        .find('input')
        .first()
        .type('198.168.100.5')
      cy.findByLabelText('Subnet mask').type('255.255.255.0')
      cy.findByLabelText('Gateway').type('198.168.255.10')
      cy.get('button[name="network_interface_addresses.dns_servers.push"]').click()
      cy.findByText('DNS servers')
        .parents('div[data-test-id="form-field"]')
        .find('input')
        .first()
        .type('198.168.100.5')
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.wait('@create-profile').its('request.body').should('deep.equal', expectedRequest)
      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile created')
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
      )
    })
  })
  describe('when updating a WiFi profile', () => {
    it('succeeds to return to list view if the profile id is not UUID', () => {
      cy.visit(
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles/edit/test-id`,
      )
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
      )
    })

    it('succeeds to update WiFi profile', () => {
      cy.intercept(
        'GET',
        `/api/v3/gcs/gateways/profiles/wifi/users/${userId}/${profiles[0].profile_id}`,
        {
          statusCode: 200,
          body: {
            profile: profiles[0],
          },
        },
      )

      cy.intercept(
        'PUT',
        `/api/v3/gcs/gateways/profiles/wifi/users/${userId}/${profiles[0].profile_id}`,
        {
          statusCode: 200,
        },
      ).as('update-profile')

      const expectedRequest = {
        profile_name: 'Updated WiFi profile',
        profile_id: '',
        shared: true,
        ssid: 'AccessPoint2',
        password: 'ABCDefg123!',
      }

      cy.findByRole('row', { name: /Test profile1/ })
        .should('be.visible')
        .within(() => {
          cy.get('button').first().click()
        })
      cy.findByLabelText('Profile name').clear()
      cy.findByLabelText('Profile name').type('Updated WiFi profile')
      cy.findByText('AccessPoint2').click()
      cy.findByLabelText('WiFi password').type('ABCDefg123!')
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.wait('@update-profile')
        .its('request.body')
        .should(body => {
          expect(body.profile).to.deep.equal(expectedRequest)
        })
      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile updated')
    })
  })

  describe('when deleting a WiFi profile', () => {
    it('succeeds to delete WiFi profile', () => {
      cy.intercept(
        'DELETE',
        `/api/v3/gcs/gateways/profiles/wifi/users/${userId}/${profiles[0].profile_id}`,
        {
          statusCode: 200,
        },
      ).as('update-profile')
      cy.findByRole('row', { name: /Test profile1/ })
        .should('be.visible')
        .within(() => {
          cy.get('button').eq(1).click()
        })
      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Confirm deletion', { selector: 'h1' }).should('be.visible')
          cy.findByRole('button', { name: /Delete/ }).click()
        })

      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile deleted')
      cy.findByRole('rowgroup').within(() => {
        cy.findAllByRole('row').should('have.length', 2)
      })
      cy.findByRole('cell', { name: /Test profile1/ }).should('not.exist')
    })
  })
})
