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

describe('Application Pub/Sub', () => {
  const userId = 'create-app-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'create-app-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const applicationId = 'pub-sub-edit-test-application'
  const application = {
    ids: {
      application_id: applicationId,
    },
  }

  const pubSubId = 'edit-pub-sub'

  const fieldMaskPaths = [
    'base_topic',
    'downlink_ack',
    'downlink_failed',
    'downlink_nack',
    'downlink_push',
    'downlink_queue_invalidated',
    'downlink_queued',
    'downlink_replace',
    'downlink_sent',
    'format',
    'ids',
    'ids.application_ids',
    'ids.application_ids.application_id',
    'ids.pub_sub_id',
    'join_accept',
    'location_solved',
    'provider',
    'provider.nats',
    'provider.nats.server_url',
    'service_data',
    'uplink_message',
  ]

  const pubsub = {
    ids: {
      application_ids: {
        application_id: applicationId,
      },
      pub_sub_id: pubSubId,
    },
    base_topic: '',
    format: 'json',
    downlink_ack: null,
    downlink_failed: null,
    downlink_nack: null,
    downlink_push: null,
    downlink_queued: null,
    downlink_queue_invalidated: null,
    downlink_replace: null,
    downlink_sent: null,
    join_accept: null,
    location_solved: null,
    service_data: null,
    uplink_message: null,
    nats: {
      server_url: 'nats://test-username:test-password@test.address.com:4222',
    },
  }

  const pubSub = {
    field_mask: {
      paths: fieldMaskPaths,
    },
    pubsub,
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
    cy.createPubSub(applicationId, pubSub)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  })

  it('succeeds editing pub sub', () => {
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${applicationId}/integrations/pubsubs/${pubSubId}`,
    )

    cy.findByLabelText('Use secure connection').check()
    cy.findByLabelText('Username').type('-updated')
    cy.findByLabelText('Password').type('-updated')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`Pub/Sub updated`)
      .should('be.visible')
  })

  it('succeeds deleting pub sub', () => {
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${applicationId}/integrations/pubsubs/${pubSubId}`,
    )
    cy.findByRole('button', { name: new RegExp('Delete Pub/Sub') }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Delete Pub/Sub', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: new RegExp('Delete Pub/Sub') }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${applicationId}/integrations/pubsubs`,
    )

    cy.findByRole('cell', { name: pubSubId }).should('not.exist')
  })
})
