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

describe('Application Webhook create', () => {
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
    cy.intercept(
      'GET',
      `/api/v3/as/webhook-templates?field_mask=base_url,create_downlink_api_key,description,documentation_url,downlink_ack,downlink_failed,downlink_nack,downlink_queue_invalidated,downlink_queued,downlink_sent,fields,format,headers,ids,info_url,join_accept,location_solved,logo_url,name,service_data,uplink_message,uplink_normalized`,
      { fixture: 'console/application/integrations/webhook/template.json' },
    )
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${appId}/integrations/webhooks/add/template`,
    )
    cy.findByText('Integrate with Test Platform').click()
  })

  it('displays UI elements in place', () => {
    cy.findByRole('heading', { name: 'Setup webhook for Test Platform' }).should('be.visible')
    cy.findByLabelText('Webhook ID')
      .should('have.attr', 'placeholder')
      .and('eq', 'my-new-test-template-webhook')
    cy.findByLabelText('Domain Secret').should('be.visible')
    cy.findByText('Akenza Core domain secret').should('be.visible')
    cy.findByLabelText('Device ID').should('be.visible')
    cy.findByText('Akenza Core device ID').should('be.visible')
    cy.findByRole('button', { name: 'Create Test Platform webhook' }).should('be.visible')
  })

  it('validates before submitting an empty form', () => {
    cy.findByRole('button', { name: 'Create Test Platform webhook' }).click()

    cy.findErrorByLabelText('Webhook ID')
      .should('contain.text', 'Webhook ID is required')
      .and('be.visible')
    cy.findErrorByLabelText('Domain Secret')
      .should('contain.text', 'Domain Secret is required')
      .and('be.visible')
    cy.findErrorByLabelText('Device ID')
      .should('contain.text', 'Device ID is required')
      .and('be.visible')
    cy.location('pathname').should(
      'eq',
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${appId}/integrations/webhooks/add/template/test-template`,
    )
  })

  it('succeeds adding webhook', () => {
    const webhook = {
      id: 'my-new-test-webhook',
      domainSecret: 'secret',
      deviceId: 'end-device-id',
    }
    cy.findByLabelText('Webhook ID').type(webhook.id)
    cy.findByLabelText('Domain Secret').type(webhook.domainSecret)
    cy.findByLabelText('Device ID').type(webhook.deviceId)

    cy.findByRole('button', { name: 'Create Test Platform webhook' }).click()

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/webhooks`,
    )

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('full-error-view').should('not.exist')
    cy.findByText('my-new-test-webhook').should('be.visible').click()

    cy.findByLabelText('Webhook ID').should('have.attr', 'value', webhook.id)
  })
})
