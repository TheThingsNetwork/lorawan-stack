// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

describe('Gateway API key modal in overview', () => {
  const userId = 'managed-gateway-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'gateway-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const gatewayId = 'test-gateway'
  const gateway = { ids: { gateway_id: gatewayId } }

  beforeEach(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createGateway(gateway, userId)

    cy.intercept('POST', `/api/v3/gateways/test-gateway/api-keys`, {
      statusCode: 200,
      body: {
        id: 'api-key-id',
        key: 'GTWAPIKEY',
      },
    }).as('create-api-key')

    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}`)
    cy.findByRole('heading', { name: gatewayId })
  })

  it('shows API key creation modal when clicking the API key button', () => {
    cy.findByRole('button', { name: /API key/i }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText(`Create a new API key for ${gatewayId}`).should('be.visible')
        cy.findByLabelText('Name').should('be.visible')
        cy.findByLabelText('Expiry date').should('be.visible')
        cy.findByText('Rights').should('be.visible')
        cy.findByText('Gateway connection (also LNS Key)').should('be.visible')
        cy.findByText('CUPS Key').should('be.visible')
        cy.findByText('Get gateway statistics').scrollIntoView()
        cy.findByText('Get gateway statistics').should('be.visible')

        cy.findByRole('button', { name: /Create API key/ }).should('be.disabled')
      })
  })

  it('enables API key creation button when valid data is entered', () => {
    cy.findByRole('button', { name: /API key/i }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByLabelText('Name').type('Test Gateway API')
        cy.findByLabelText('Expiry date').type('2030-01-01')

        cy.findByText('Gateway connection (also LNS Key)').should('be.visible').click()

        cy.findByRole('button', { name: /Create API key/ }).should('not.be.disabled')
      })
  })

  it('shows error if error happens while creating API key', () => {
    cy.intercept('POST', `/api/v3/gateways/test-gateway/api-keys`, {
      statusCode: 500,
    }).as('create-api-key-error')
    cy.findByRole('button', { name: /API key/i }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Gateway connection (also LNS Key)').should('be.visible').click()

        cy.findByRole('button', { name: /Create API key/ }).click()
        cy.wait('@create-api-key-error')
      })
    cy.findByTestId('toast-notification-error')
      .should('be.visible')
      .findByText('Failed to create API key')
      .should('be.visible')
  })

  it('creates API key and shows confirmation modal', () => {
    cy.window().then(win => {
      cy.stub(win.navigator.clipboard, 'writeText').as('clipboardWrite')
    })

    cy.findByRole('button', { name: /API key/i }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByLabelText('Name').type('Copy Test API Key')
        cy.findByLabelText('Expiry date').type('2030-12-31')
        cy.findByText('Gateway connection (also LNS Key)').should('be.visible').click()

        cy.findByRole('button', { name: /Create API key/ }).click()
        cy.wait('@create-api-key')

        cy.findByText('Your API key has been created successfully').should('be.visible')
        cy.findByText(/After closing this window, the value of the key secret/i).should(
          'be.visible',
        )
        cy.findByText('API key')
          .parent()
          .within(() => {
            cy.get('div').eq(1).should('contain.text', '••••')

            cy.get('button[title="Toggle visibility"]').click()

            cy.get('div')
              .eq(1)
              .invoke('text')
              .should('not.contain', '•')
              .and('contain', 'GTWAPIKEY')
          })

        cy.findByRole('button', { name: /Copy/ }).click()
        cy.get('@clipboardWrite').should('have.been.calledOnce')
        cy.findByRole('button', { name: /Copied to clipboard/ }).should('be.visible')

        cy.findByRole('button', { name: /Download/ }).should('be.visible')
        cy.findByRole('button', { name: /I have copied the key/ }).click()
      })

    cy.findByTestId('modal-window').should('not.exist')
  })

  it('creates API key and downloads it to a file', () => {
    cy.window().then(win => {
      cy.spy(win.document.body, 'appendChild').as('appendChildSpy')
    })

    cy.findByRole('button', { name: /API key/i }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByLabelText('Name').type('Download Test API Key')
        cy.findByLabelText('Expiry date').type('2030-12-31')
        cy.findByText('Gateway connection (also LNS Key)').should('be.visible').click()

        cy.findByRole('button', { name: /Create API key/ }).click()
        cy.wait('@create-api-key')

        cy.findByText('Your API key has been created successfully').should('be.visible')
        cy.findByText(/After closing this window, the value of the key secret/i).should(
          'be.visible',
        )

        cy.findByRole('button', { name: /Download/ }).click()
        cy.get('@appendChildSpy').should('have.been.called')
        cy.findByRole('button', { name: /I have copied the key/ }).click()
      })

    cy.findByTestId('modal-window').should('not.exist')
  })
})
