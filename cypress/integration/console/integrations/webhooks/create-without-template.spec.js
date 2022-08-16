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

describe('Application Webhook create without template', () => {
  const userId = 'create-app-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'create-app-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const appId = 'webhook-test-application'
  const application = {
    ids: {
      application_id: appId,
    },
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: userId, password: user.password })
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${appId}/integrations/webhooks/add/template`,
    )
    cy.findByText('Custom webhook').click()
  })

  it('displays UI elements in place', () => {
    cy.findByRole('heading', { name: 'Add webhook' }).should('be.visible')
    cy.findByRole('heading', { name: 'General settings' }).should('be.visible')
    cy.findByText(/The Webhooks feature allows/).should('be.visible')
    cy.findByLabelText('Webhook ID').should('have.attr', 'placeholder').and('eq', 'my-new-webhook')
    cy.findByLabelText('Base URL')
      .should('have.attr', 'placeholder')
      .and('eq', 'https://example.com/webhooks')
    cy.findDescriptionByLabelText('Downlink API key').should(
      'contain',
      'The API key will be provided to the endpoint using the "X-Downlink-Apikey" header',
    )
    cy.findByText('Add header entry').should('be.visible')
    cy.findByText(/Filter event data/).should('be.visible')
    cy.findByText('Add filter path').should('be.visible')
    cy.findByRole('heading', { name: 'Enabled event types' }).should('be.visible')
    cy.findByText(
      'For each enabled event type an optional path can be defined which will be appended to the base URL',
    ).should('be.visible')

    cy.findByText('Uplink message').should('be.visible').and('not.be.checked')
    cy.findByText('Join accept').should('be.visible').and('not.be.checked')
    cy.findByText('Downlink ack').should('be.visible').and('not.be.checked')
    cy.findByText('Downlink nack').should('be.visible').and('not.be.checked')
    cy.findByText('Downlink sent').should('be.visible').and('not.be.checked')
    cy.findByText('Downlink failed').should('be.visible').and('not.be.checked')
    cy.findByText('Downlink queued').should('be.visible').and('not.be.checked')
    cy.findByText('Downlink queue invalidated').should('be.visible').and('not.be.checked')
    cy.findByText('Location solved').should('be.visible').and('not.be.checked')
    cy.findByText('Service data').should('be.visible').and('not.be.checked')

    cy.findByRole('button', { name: 'Add webhook' }).should('be.visible')
  })

  it('validates before submitting an empty form', () => {
    cy.findByRole('button', { name: 'Add webhook' }).click()

    cy.findErrorByLabelText('Webhook ID')
      .should('contain.text', 'Webhook ID is required')
      .and('be.visible')
    cy.findErrorByLabelText('Base URL')
      .should('contain.text', 'Base URL is required')
      .and('be.visible')
    cy.location('pathname').should(
      'eq',
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${appId}/integrations/webhooks/add/template/custom`,
    )
  })

  it('succeeds adding webhook', () => {
    const webhook = {
      id: 'my-new-webhook',
      format: 'JSON',
      baseUrl: 'https://example.com/webhooks',
      path: '/path/to/webhook',
    }
    cy.findByLabelText('Webhook ID').type(webhook.id)
    cy.findByLabelText('Webhook format').selectOption(webhook.format)
    cy.findByLabelText('Base URL').type(webhook.baseUrl)
    cy.findByLabelText('Uplink message')
      .check()
      .parents('[data-test-id="form-field"]')
      .within(() => {
        cy.findByPlaceholderText('/path/to/webhook').type(webhook.path)
      })

    cy.findByRole('button', { name: /Add header entry/ }).click()
    cy.findByTestId('_headers[0].key').type('webhook-test-key')
    cy.findByTestId('_headers[0].value').type('webhook-test-value')

    cy.findByRole('button', { name: /Add filter path/ }).click()
    cy.findByText('Filter event data')
      .parents('div[data-test-id="form-field"]')
      .find('input')
      .first()
      .selectOption('received_at')

    cy.findByRole('button', { name: 'Add webhook' }).click()

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/webhooks`,
    )

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('full-error-view').should('not.exist')
    cy.findByText('my-new-webhook').should('be.visible')

    // Displays saved created webhook settings
    cy.visit(
      `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/webhooks/${
        webhook.id
      }`,
    )

    cy.findByRole('heading', { name: 'Edit webhook' }).should('be.visible')
    cy.findByLabelText('Webhook ID')
      .should('be.disabled')
      .and('have.attr', 'value')
      .and('eq', webhook.id)
    cy.findByLabelText('Base URL').and('have.attr', 'value', webhook.baseUrl)
    cy.findByLabelText('Uplink message')
      .check()
      .parents('[data-test-id="form-field"]')
      .within(() => {
        cy.findByPlaceholderText('/path/to/webhook').should('have.attr', 'value', webhook.path)
      })
  })
})
