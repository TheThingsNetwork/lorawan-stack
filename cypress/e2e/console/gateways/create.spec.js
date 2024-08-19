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

import { disableGatewayServer, generateHexValue } from '../../../support/utils'

describe('Gateway create', () => {
  const userId = 'create-gateway-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'edit-gateway-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/add`)
  })

  it('displays UI elements in place', () => {
    cy.findByText('Register gateway', { selector: 'h1' }).should('be.visible')
    cy.findByLabelText('Gateway EUI').should('be.visible')

    cy.findByText(
      'To continue, please confirm the Gateway EUI so we can determine onboarding options',
    ).should('be.visible')
    cy.findByRole('button', { name: 'Continue without EUI' }).should('be.visible')
  })

  it('succeeds to show manual registration form', () => {
    cy.findByLabelText('Gateway EUI').type(generateHexValue(16))
    cy.findByRole('button', { name: 'Confirm' }).click()

    cy.findByRole('button', { name: 'Reset' }).should('be.visible')
    cy.findByLabelText('Gateway EUI').should('be.visible').should('be.disabled')
    cy.findByLabelText('Gateway name').should('be.visible')
    cy.findByTestId('key-value-map').should('be.visible')

    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/gateways/add`)
    cy.findByTestId('error-notification').should('not.exist')
  })

  it('succeeds to reset manual registration form', () => {
    cy.findByLabelText('Gateway EUI').type(generateHexValue(16))
    cy.findByRole('button', { name: 'Confirm' }).click()
    cy.findByRole('button', { name: 'Reset' }).should('be.visible').click()
    cy.findByLabelText('Gateway EUI')
      .should('be.visible')
      .should('not.be.disabled')
      .should('have.value', '')
    cy.findByRole('button', { name: 'Continue without EUI' }).should('be.visible')

    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/gateways/add`)
  })

  it('succeeds adding gateway manually', () => {
    const gateway = {
      frequency_plan: 'EU_863_870',
      eui: generateHexValue(16),
    }
    cy.intercept('GET', '/api/v3/configuration/frequency-plans').as('get-frequency-plans')
    cy.findByLabelText('Gateway EUI').type(gateway.eui)
    cy.findByLabelText('Gateway EUI').blur()
    cy.findByRole('button', { name: 'Confirm' }).click()

    cy.wait('@get-frequency-plans')
    cy.findByLabelText('Gateway ID').type(`eui-${gateway.eui}`)
    cy.findByLabelText('Gateway name').type('Test Gateway')
    cy.findByText('Frequency plan')
      .parents('div[data-test-id="form-field"]')
      .find('input')
      .first()
      .selectOption(gateway.frequency_plan)
    cy.findByRole('button', { name: 'Register gateway' }).click()

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/gateways/eui-${gateway.eui}`,
    )
    cy.findByRole('heading', { name: 'Test Gateway' })
    cy.findByText('Europe 863-870 MHz (SF12 for RX2)').should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')
  })

  it('displays claiming gateway form', () => {
    const gateway = {
      frequency_plan: 'EU_863_870',
      eui: '58A0CBFFFE000001',
    }

    cy.intercept('POST', '/api/v3/gcls/claim/info', {
      body: {
        supports_claiming: true,
      },
    })

    cy.findByLabelText('Gateway EUI').type(gateway.eui)
    cy.findByRole('button', { name: 'Confirm' }).click()
    cy.findByTestId('notification').should('exist')
    cy.findByLabelText('Claim authentication code').type('12345')
    cy.findByLabelText('Gateway ID').type(`eui-${gateway.eui}`)
    cy.findByText('Frequency plan')
      .parents('div[data-test-id="form-field"]')
      .find('input')
      .first()
      .selectOption(gateway.frequency_plan)
  })

  it('succeeds claiming gateway', () => {
    const gateway = {
      frequency_plan: 'EU_863_870',
      eui: generateHexValue(16),
    }
    const expectedRequest = {
      collaborator: {
        user_ids: {
          user_id: user.ids.user_id,
        },
      },
      authenticated_identifiers: {
        gateway_eui: gateway.eui.toUpperCase(),
        authentication_code: 'MTIzNDU=',
      },
      target_gateway_id: `eui-${gateway.eui}`,
      target_frequency_plan_id: gateway.frequency_plan,
      target_gateway_server_address: window.location.hostname,
    }
    cy.intercept('GET', '/api/v3/gateways/*', { statusCode: 200 })
    cy.intercept('POST', '/api/v3/gcls/claim/info', {
      body: {
        supports_claiming: true,
      },
    })

    cy.intercept('POST', '/api/v3/gcls/claim', {
      statusCode: 201,
      body: {
        gateway_id: `eui-${gateway.eui}`,
        eui: gateway.eui,
      },
    }).as('claim-request')

    cy.findByLabelText('Gateway EUI').type(gateway.eui)
    cy.findByRole('button', { name: 'Confirm' }).click()
    cy.findByLabelText('Frequency plan').selectOption(gateway.frequency_plan)
    cy.findByLabelText('Gateway ID').type(`eui-${gateway.eui}`)
    cy.findByLabelText('Claim authentication code').type('12345')
    cy.findByRole('button', { name: 'Claim gateway' }).click()
    cy.wait('@claim-request').its('request.body').should('deep.equal', expectedRequest)
    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/gateways/eui-${gateway.eui}`,
    )
  })

  it('succeeds adding gateway without frequency plan and without EUI', () => {
    const gateway = {
      gateway_id: 'gateway-without-fp-eui',
      frequency_plan: 'EU_863_870',
    }
    cy.intercept('GET', '/api/v3/configuration/frequency-plans').as('get-frequency-plans')
    cy.findByRole('button', { name: 'Continue without EUI' }).click()
    cy.findByLabelText('Gateway ID').type(gateway.gateway_id)
    cy.wait('@get-frequency-plans')
    cy.findByText('Frequency plan')
      .parents('div[data-test-id="form-field"]')
      .find('input')
      .first()
      .selectOption('no-frequency-plan')
    cy.findByText(/Without choosing a frequency plan/)
    cy.findByRole('button', { name: 'Register gateway' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/gateways/${gateway.gateway_id}`,
    )
    cy.findByRole('heading', { name: gateway.gateway_id })
    cy.findByTestId('error-notification').should('not.exist')
  })

  it('succeeds converting MAC to EUI', () => {
    const gatewayMac = generateHexValue(12)

    cy.findByLabelText('Gateway EUI').type(gatewayMac)
    cy.contains('Convert MAC to EUI', { timeout: 3500 }).should('be.visible').click()
    cy.contains('Convert MAC to EUI').should('not.exist')
    cy.findByRole('button', { name: 'Confirm' }).click()

    const gatewayEui = `${gatewayMac.substring(0, 6)}fffe${gatewayMac.substring(6)}`
    cy.findByLabelText('Gateway ID').type(`eui-${gatewayEui}`)
  })

  it('succeeds showing modal when generating API keys for CUPS and LNS', () => {
    const gateway = {
      frequency_plan: 'EU_863_870',
      eui: generateHexValue(16),
    }
    cy.intercept('GET', '/api/v3/configuration/frequency-plans').as('get-frequency-plans')
    cy.findByLabelText('Gateway EUI').type(gateway.eui)
    cy.findByLabelText('Gateway EUI').blur()
    cy.findByRole('button', { name: 'Confirm' }).click()

    cy.wait('@get-frequency-plans')
    cy.findByLabelText('Gateway ID').type(`eui-${gateway.eui}`)
    cy.findByLabelText('Gateway name').type('Test Gateway')
    cy.findByText('Frequency plan')
      .parents('div[data-test-id="form-field"]')
      .find('input')
      .first()
      .selectOption(gateway.frequency_plan)
    cy.findByLabelText(/Require authenticated connection/).check()
    cy.findByLabelText(/Generate API key for CUPS/).check()
    cy.findByLabelText(/Generate API key for LNS/).check()
    cy.findByRole('button', { name: 'Register gateway' }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Download gateway API keys', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Download LNS key/ }).click()
        cy.findByRole('button', { name: /Download CUPS key/ }).click()
        cy.findByText(
          'Note: After closing this window, these API keys will not be accessible for download anymore. Please make sure to download and store them now.',
        ).should('be.visible')
        cy.findByRole('button', { name: /I have downloaded the keys/ }).click()
      })

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/gateways/eui-${gateway.eui}`,
    )

    cy.findByRole('heading', { name: 'Test Gateway' })
    cy.findByText('Europe 863-870 MHz (SF12 for RX2)').should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')
  })

  it('succeeds adding gateway with multiple frequency plans', () => {
    const gateway = {
      frequency_plan: 'EU_863_870',
      eui: generateHexValue(16),
    }
    cy.intercept('GET', '/api/v3/configuration/frequency-plans').as('get-frequency-plans')
    cy.findByLabelText('Gateway EUI').type(gateway.eui)
    cy.findByLabelText('Gateway EUI').blur()
    cy.findByRole('button', { name: 'Confirm' }).click()

    cy.findByLabelText('Gateway ID').type(`eui-${gateway.eui}`)
    cy.wait('@get-frequency-plans')

    cy.findByText('Frequency plan')
      .parents('div[data-test-id="form-field"]')
      .find('input')
      .first()
      .selectOption(gateway.frequency_plan)
    cy.findByRole('button', { name: /Add frequency plan/ }).click()
    cy.findByText('Frequency plan')
      .parent()
      .parent()
      .find('input')
      .eq(2)
      .selectOption('EU_863_870_TTN')
    cy.findByRole('button', { name: 'Register gateway' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/gateways/eui-${gateway.eui}`,
    )
    cy.findByRole('heading', { name: `eui-${gateway.eui}` })
    cy.findByText('Frequency plan')
    cy.findByText(
      'Europe 863-870 MHz (SF12 for RX2) , Europe 863-870 MHz (SF9 for RX2 - recommended)',
    ).should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')
  })

  describe('Gateway Server disabled', () => {
    beforeEach(() => {
      cy.augmentStackConfig(disableGatewayServer)
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/gateways/add`)
    })

    it('succeeds adding gateway without frequency plan', () => {
      const gateway = {
        frequency_plan: 'EU_863_870',
        eui: generateHexValue(16),
      }

      cy.findByLabelText('Gateway EUI').type(gateway.eui)
      cy.findByRole('button', { name: 'Confirm' }).click()
      cy.findByLabelText('Gateway ID').type(`eui-${gateway.eui}`)
      cy.findByTestId('key-value-map').should('not.exist')
      cy.findByRole('button', { name: 'Register gateway' }).click()

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/eui-${gateway.eui}`,
      )
      cy.findByRole('heading', { name: `eui-${gateway.eui}` })
      cy.findByTestId('error-notification').should('not.exist')
    })
  })
})
