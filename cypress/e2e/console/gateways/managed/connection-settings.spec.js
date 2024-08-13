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

describe('Managed Gateway connection settings', () => {
  const userId = 'managed-gateway-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'managed-gateway-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const gatewayId = 'test-managed-gateway'
  const gateway = { ids: { gateway_id: gatewayId } }

  const gatewayVersionIds = {
    hardware_version: 'v1.1',
    firmware_version: 'v1.1',
    model_id: 'Managed gateway',
  }

  const wifiProfileId = 'test-profile1'
  const ethernetProfileId = 'ethernet-profile'

  const organizationId = 'test-organization'

  beforeEach(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createGateway(gateway, userId)

    // Interceptors
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

    cy.intercept('GET', `/api/v3/gcs/gateways/profiles/wifi/users/${userId}`, {
      statusCode: 200,
      body: {
        profiles: [
          {
            profile_id: 'test-profile1',
            profile_name: 'Test profile1',
            shared: true,
            ssid: 'profile1',
          },
          {
            profile_id: 'test-profile2',
            profile_name: 'Test profile2',
            shared: true,
            ssid: 'profile2',
          },
        ],
      },
    })

    cy.intercept('PUT', `/api/v3/gcs/gateways/managed/${gatewayId}*`, {
      statusCode: 200,
      body: {
        ids: {
          gateway_id: gatewayId,
        },
      },
    }).as('update-connection-settings')

    cy.intercept('POST', `/api/v3/gcs/gateways/managed/${gatewayId}/wifi/scan`, {
      statusCode: 200,
      body: {
        access_points: [
          {
            ssid: 'AccessPoint1',
            bssid: 'EC656E000100',
            channel: 0,
            authentication_mode: 'open',
            rssi: -70,
          },
        ],
      },
    }).as('scan-access-points')

    cy.intercept('POST', `/api/v3/gcs/gateways/profiles/wifi/users/${userId}`, {
      statusCode: 200,
      body: {
        profile_id: wifiProfileId,
      },
    })

    cy.intercept('POST', `/api/v3/gcs/gateways/profiles/ethernet/users/${userId}`, {
      statusCode: 200,
      body: {
        profile_id: ethernetProfileId,
      },
    })

    cy.intercept('GET', `/api/v3/gcs/gateways/profiles/wifi/organizations/${organizationId}`, {
      statusCode: 200,
      body: {
        profiles: [
          {
            profile_id: 'test-profile1',
            profile_name: 'Test profile1',
            shared: true,
            ssid: 'profile1',
          },
          {
            profile_id: 'test-profile2',
            profile_name: 'Test profile2',
            shared: true,
            ssid: 'profile2',
          },
        ],
      },
    })
    // End interceptors.

    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}`)
    cy.wait('@get-is-gtw-managed')
    cy.findByRole('heading', { name: 'test-managed-gateway' })
    cy.get('button').contains('Managed gateway').click()
    cy.get('a').contains('Connection settings').click()
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/connection-settings`,
    )
    cy.findByTestId('error-notification').should('not.exist')
  })

  it('succeeds to display UI elements in place', () => {
    cy.findByText('WiFi connection', { selector: 'h3' }).should('be.visible')
    cy.findByText('Ethernet connection', { selector: 'h3' }).should('be.visible')
    cy.findByText('Connection settings profiles can be shared within the same organization').should(
      'be.visible',
    )
    cy.findByLabelText('Show profiles of').should('have.attr', 'disabled')
    cy.findByRole('button', { name: 'Save changes' }).should('be.visible')

    cy.findByText(gatewayVersionIds.model_id, { selector: 'h3' }).should('be.visible')
  })

  it('succeeds to set WiFi connection with already created profile', () => {
    cy.findByLabelText('Settings profile').selectOption(wifiProfileId)
    cy.findByText(
      'Please click "Save changes" to start using this WiFi profile for the gateway',
    ).should('be.visible')
    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.wait('@update-connection-settings')
      .its('request.body')
      .should(body => {
        expect(body).to.have.nested.property('gateway.wifi_profile_id', wifiProfileId)
      })
    cy.findByText('The gateway WiFi is currently attempting to connect using this profile').should(
      'be.visible',
    )
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .and('contain', 'Connection settings updated')
  })

  it('succeeds to validate new WiFi profile fields', () => {
    cy.findByLabelText('Settings profile').selectOption('shared')
    cy.wait('@scan-access-points')
    cy.findByLabelText(/Use default network interface settings/).uncheck()

    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.get('#wifi_profile\\.profile_name-field-error').should('be.visible')
    cy.get('#wifi_profile\\._access_point-field-error').should('be.visible')
    cy.get('#wifi_profile\\.network_interface_addresses\\.ip_addresses-field-error').should(
      'be.visible',
    )
    cy.get('#wifi_profile\\.network_interface_addresses\\.subnet_mask-field-error').should(
      'be.visible',
    )
    cy.get('#wifi_profile\\.network_interface_addresses\\.gateway-field-error').should('be.visible')
  })

  it('succeeds to set WiFi connection with new shared profile', () => {
    cy.findByLabelText('Settings profile').selectOption('shared')
    cy.wait('@scan-access-points')
    cy.findByLabelText('Profile name').type('New WiFi profile')
    cy.findByText('AccessPoint1').click()
    cy.findByText(
      'Please click "Save changes" to start using this WiFi profile for the gateway',
    ).should('be.visible')
    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.wait('@update-connection-settings')
      .its('request.body')
      .should(body => {
        expect(body).to.have.nested.property('gateway.wifi_profile_id', wifiProfileId)
      })
    cy.findByText('The gateway WiFi is currently attempting to connect using this profile').should(
      'be.visible',
    )
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .and('contain', 'Connection settings updated')
  })

  it('succeeds to set WiFi connection with new non-shared profile', () => {
    cy.findByLabelText('Settings profile').selectOption('non-shared')
    cy.wait('@scan-access-points')
    cy.findByText('AccessPoint1').click()
    cy.findByText(
      'Please click "Save changes" to start using this WiFi profile for the gateway',
    ).should('be.visible')
    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.wait('@update-connection-settings')
      .its('request.body')
      .should(body => {
        expect(body).to.have.nested.property('gateway.wifi_profile_id', wifiProfileId)
      })
    cy.findByText('The gateway WiFi is currently attempting to connect using this profile').should(
      'be.visible',
    )
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .and('contain', 'Connection settings updated')
  })

  it('succeeds to set Ethernet connection with default network settings', () => {
    cy.findByLabelText(/Enable ethernet connection/).check()
    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.wait('@update-connection-settings')
      .its('request.body')
      .should(body => {
        expect(body).to.have.nested.property('gateway.ethernet_profile_id', ethernetProfileId)
      })
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .and('contain', 'Connection settings updated')
  })

  it('succeeds to validate custom ethernet network settings fields', () => {
    cy.findByLabelText(/Enable ethernet connection/).check()
    cy.findByLabelText(/Use a static IP address/).check()
    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.get('#ethernet_profile\\.network_interface_addresses\\.ip_addresses-field-error').should(
      'be.visible',
    )
    cy.get('#ethernet_profile\\.network_interface_addresses\\.subnet_mask-field-error').should(
      'be.visible',
    )
    cy.get('#ethernet_profile\\.network_interface_addresses\\.gateway-field-error').should(
      'be.visible',
    )
  })

  it('succeeds to set Ethernet connection with custom network settings', () => {
    cy.findByLabelText(/Enable ethernet connection/).check()
    cy.findByLabelText(/Use a static IP address/).check()
    cy.findByText('IP addresses')
      .parents('div[data-test-id="form-field"]')
      .find('input')
      .first()
      .type('198.168.100.5')
    cy.findByLabelText('Subnet mask').type('255.255.255.0')
    cy.findByLabelText('Gateway').type('198.168.255.10')
    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.wait('@update-connection-settings')
      .its('request.body')
      .should(body => {
        expect(body).to.have.nested.property('gateway.ethernet_profile_id', ethernetProfileId)
      })
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .and('contain', 'Connection settings updated')
  })

  it("succeeds to set organization's WiFi profile", () => {
    const organization = { ids: { organization_id: organizationId }, name: 'Test organization' }
    cy.createOrganization(organization, userId)
    cy.createCollaborator('gateways', gatewayId, {
      collaborator: {
        ids: {
          organization_ids: {
            organization_id: organizationId,
          },
        },
        rights: ['RIGHT_GATEWAY_ALL'],
      },
    })
    cy.reload()
    cy.findByLabelText('Show profiles of').should('not.have.attr', 'disabled')
    cy.findByLabelText('Show profiles of').selectOption(organizationId)
    cy.findByLabelText('Settings profile').selectOption(wifiProfileId)
    cy.findByText(
      'Please click "Save changes" to start using this WiFi profile for the gateway',
    ).should('be.visible')
    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.wait('@update-connection-settings')
      .its('request.body')
      .should(body => {
        expect(body).to.have.nested.property('gateway.wifi_profile_id', wifiProfileId)
      })
    cy.findByText('The gateway WiFi is currently attempting to connect using this profile').should(
      'be.visible',
    )
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .and('contain', 'Connection settings updated')
  })
})
