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

// TODO: Update tests
describe('Managed Gateway WiFi profiles', () => {
  const generateUUID = () =>
    'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
      const r = (Math.random() * 16) | 0,
        v = c === 'x' ? r : (r & 0x3) | 0x8
      return v.toString(16)
    })

  const userId = generateUUID()
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'managed-gateway-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const organizationId = generateUUID()
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

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createGateway(gateway, userId)
    cy.createOrganization(organization, userId)
    cy.createCollaborator('gateways', gatewayId, collaborator)
  })

  beforeEach(() => {
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
    // Cy.findByTestId('error-notification').should('not.exist')
  })

  it('succeeds to display UI elements in place', () => {
    cy.findByText('WiFi profiles', { selector: 'h1' }).should('be.visible')
    cy.contains('button', 'Add WiFi profile').should('be.visible')
    // Cy.findByText('No items found').should('be.visible')
  })

  // eslint-disable-next-line jest/no-commented-out-tests
  /* Describe('when creating a WiFi profile', () => {
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
      cy.findByLabelText('Profile name').type('Open WiFi profile')
      cy.findByText('AccessPoint1').click()
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile created')
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
      )
      cy.findByLabelText('Show profiles of').should('be.visible').selectOption(userId)
      cy.findByRole('rowgroup').within(() => {
        cy.findAllByRole('row').should('have.length', 1)
      })
      cy.findByRole('cell', { name: 'Open WiFi profile' }).should('be.visible')
    })

    it('succeeds to create WiFi profile with secured access point', () => {
      cy.findByLabelText('Profile name').type('Secured WiFi profile')
      cy.findByText('AccessPoint2').click()
      cy.findByLabelText('WiFi password').type('ABCDefg123!')
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile created')
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
      )
      cy.findByLabelText('Show profiles of').should('be.visible').selectOption(userId)
      cy.findByRole('rowgroup').within(() => {
        cy.findAllByRole('row').should('have.length', 2)
      })
      cy.findByRole('cell', { name: 'Secured WiFi profile' }).should('be.visible')
    })

    it('succeeds to create WiFi profile with other access point', () => {
      cy.findByLabelText('Profile name').type('Other WiFi profile')
      cy.findByText('Other...').click()
      cy.findByLabelText('SSID').type('AccessPoint3')
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile created')
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
      )
      cy.findByLabelText('Show profiles of').should('be.visible').selectOption(userId)
      cy.findByRole('rowgroup').within(() => {
        cy.findAllByRole('row').should('have.length', 3)
      })
      cy.findByRole('cell', { name: 'Other WiFi profile' }).should('be.visible')
    })

    it('succeeds to create WiFi profile with custom network settings', () => {
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
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile created')
      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
      )
      cy.findByLabelText('Show profiles of').should('be.visible').selectOption(userId)
      cy.findByRole('rowgroup').within(() => {
        cy.findAllByRole('row').should('have.length', 4)
      })
      cy.findByRole('cell', { name: 'Custom WiFi profile' }).should('be.visible')
    })
  })

  describe('when updating a WiFi profile', () => {
    it('succeeds to update WiFi profile', () => {
      cy.findByRole('row', { name: /Open WiFi profile/ })
        .should('be.visible')
        .within(() => {
          cy.get('button').first().click()
        })
      cy.findByLabelText('Profile name').clear()
      cy.findByLabelText('Profile name').type('Updated WiFi profile')
      cy.findByText('AccessPoint2').click()
      cy.findByLabelText('WiFi password').type('ABCDefg123!')
      cy.findByRole('button', { name: 'Save changes' }).click()
      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile updated')

      cy.visit(
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/managed-gateway/wifi-profiles`,
      )
      cy.findByLabelText('Show profiles of').should('be.visible').selectOption(userId)
      cy.findByRole('cell', { name: 'Updated WiFi profile' }).should('be.visible')
      cy.findByRole('cell', { name: 'Open WiFi profile' }).should('not.exist')
    })
  })

  describe('when deleting a WiFi profile', () => {
    it('succeeds to delete WiFi profile', () => {
      cy.findByRole('row', { name: /Updated WiFi profile/ })
        .should('be.visible')
        .within(() => {
          cy.get('button').eq(1).click()
        })
      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Confirm deletion', { selector: 'h1' }).should('be.visible')
          cy.findByRole('button', { name: /Delete profile/ }).click()
        })

      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .and('contain', 'WiFi profile deleted')
      cy.findByRole('rowgroup').within(() => {
        cy.findAllByRole('row').should('have.length', 3)
      })
      cy.findByRole('cell', { name: 'Updated WiFi profile' }).should('not.exist')
    })
  })

  describe('when listing WiFi profiles for organization', () => {
    it('succeeds to show correct WiFi profiles', () => {
      cy.findByText('No items found').should('not.exist')
      cy.findByRole('rowgroup').within(() => {
        cy.findAllByRole('row').should('have.length', 3)
      })
      cy.findByLabelText('Show profiles of').selectOption(organizationId)
      cy.findByText('No items found').should('be.visible')
      cy.findByRole('rowgroup').should('not.exist')
    })
  })*/
})
