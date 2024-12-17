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

describe('Application Webhook', () => {
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
  const webhookId = 'my-edit-test-webhook'
  const webhookBody = {
    webhook: {
      base_url: 'https://example.com/edit-webhooks-test',
      format: 'json',
      ids: {
        application_ids: {},
        webhook_id: webhookId,
      },
    },
    field_mask: {
      paths: ['base_url', 'format', 'ids', 'ids.application_ids', 'ids.webhook_id'],
    },
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
    cy.createWebhook(appId, webhookBody)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: userId, password: user.password })
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${appId}/integrations/webhooks/${webhookId}`,
    )
  })

  it('succeeds editing webhook', () => {
    const webhook = {
      format: 'Protocol Buffers',
      url: 'https://example.com/webhooks-updated',
      path: 'path/to/webhook',
    }

    cy.findByLabelText('Webhook format').selectOption(webhook.format)
    cy.findByLabelText('Base URL').clear()
    cy.findByLabelText('Base URL').type(webhook.url)
    cy.findByLabelText('Uplink message').check()
    cy.findByLabelText('Uplink message')
      .parents('[data-test-id="form-field"]')
      .within(() => {
        cy.findByPlaceholderText('/path/to/webhook').type(webhook.path)
      })
    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification-success').findByText('Webhook updated').should('be.visible')

    cy.reload()
    cy.findByLabelText('Base URL').should('have.attr', 'value', webhook.url)
    cy.get('input[name="uplink_message.enable"]').scrollIntoView()
    cy.findByLabelText('Uplink message')
      .should('be.checked')
      .parents('[data-test-id="form-field"]')
      .within(() => {
        cy.findByDisplayValue(webhook.path).should('be.visible').clear()
      })
    cy.findByLabelText('Join accept').should('not.be.checked')
    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification-success').findByText(`Webhook updated`).should('be.visible')

    cy.reload()
    cy.findByLabelText('Base URL').should('have.attr', 'value', webhook.url)
    cy.findByLabelText('Uplink message')
      .should('be.checked')
      .parents('[data-test-id="form-field"]')
      .within(() => {
        cy.findByPlaceholderText('/path/to/webhook').should('have.attr', 'value', '')
      })
    cy.findByLabelText('Uplink message').uncheck()
    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification-success').findByText(`Webhook updated`).should('be.visible')

    cy.reload()
    cy.findByLabelText('Uplink message').should('not.be.checked')
  })

  it('succeeds adding headers and filters', () => {
    cy.findByRole('button', { name: /Add header entry/ }).click()

    cy.findByTestId('_headers[0].key').type('webhook-test-key')
    cy.findByTestId('_headers[0].value').type('webhook-test-value')

    cy.findByRole('button', { name: /Add filter path/ }).click()
    cy.findByText('Filter event data')
      .parents('div[data-test-id="form-field"]')
      .find('input')
      .first()
      .selectOption('received_at')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .findByText(`Webhook updated`)
      .should('be.visible')

    cy.reload()

    cy.findByTestId('_headers[0].key')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', 'webhook-test-key')
    cy.findByTestId('_headers[0].value')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', 'webhook-test-value')
    cy.get('input[name="field_mask.paths[0].value"]')
      .should('exist')
      .and('have.attr', 'value')
      .and('eq', 'received_at')
  })

  it('succeeds adding basic authorization header', () => {
    cy.findByLabelText('Request authentication').check()

    cy.findByTestId('_headers[0].key')
      .should('have.attr', 'value', 'Authorization')
      .and('have.attr', 'readonly')
    cy.findByTestId('_headers[0].value')
      .should('have.attr', 'value', 'Basic ...')
      .and('have.attr', 'readonly')

    cy.findByTestId('basic-auth-username').should('be.visible').type('test-user')
    cy.findByTestId('basic-auth-password').should('be.visible').type('1234QUERTY!')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .findByText(`Webhook updated`)
      .should('be.visible')

    cy.reload()

    cy.findByLabelText('Request authentication').should('have.attr', 'value', 'true')
    cy.findByTestId('basic-auth-username')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', 'test-user')
    cy.findByTestId('basic-auth-password')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', '1234QUERTY!')
  })

  it('succeeds pausing and activating webhook', () => {
    cy.findByRole('button', { name: /Pause/ }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Pause webhook?', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Pause webhook/ }).click()
      })
    cy.findByTestId('toast-notification-success').findByText('Webhook paused').should('be.visible')

    cy.findAllByRole('button', { name: /Activate/ }).should('have.length', 2)
    cy.findByTestId('notification')
      .should('exist')
      .findByRole('button', { name: /Activate/ })
      .click()

    cy.findByTestId('toast-notification-success').findByText('Webhook active').should('be.visible')
  })

  it('succeeds deleting webhook', () => {
    cy.findByRole('button', { name: /Delete Webhook/ }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Delete Webhook', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Delete Webhook/ }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/webhooks`,
    )

    cy.findByRole('cell', { name: webhookId }).should('not.exist')
  })
})
