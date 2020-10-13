// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

describe('Application Pub/Sub create', () => {
  const userId = 'create-app-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'create-app-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const appId = 'pub-sub-test-application'
  const application = {
    ids: {
      application_id: appId,
    },
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.loginConsole({ user_id: userId, password: user.password })
    cy.createApplication(application, userId)
    cy.clearLocalStorage()
    cy.clearCookies()
  })

  describe('MQTT', () => {
    beforeEach(() => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
      )
      cy.findByLabelText('MQTT').check()
    })

    it('succeeds adding pub-sub without credentials', () => {
      const pubSub = {
        id: 'no-creds-pub-sub',
        serverUrl: 'mqtts://example.com',
        clientId: 'no-creds-client-id',
        subscribe_qos: 'AT_MOST_ONCE',
        publish_qos: 'AT_MOST_ONCE',
        format: 'json',
        uplinkSubTopic: 'uplink-test-sub-topic',
      }
      cy.findByLabelText('Pub/Sub ID').type(pubSub.id)
      cy.findByLabelText('Server URL').type(pubSub.serverUrl)
      cy.findByLabelText('Client ID').type(pubSub.clientId)
      cy.findByLabelText('Use credentials').uncheck({ force: true })
      cy.findByLabelText('Subscribe QoS').selectOption(pubSub.subscribe_qos)
      cy.findByLabelText('Publish QoS').selectOption(pubSub.publish_qos)
      cy.findByLabelText('Pub/Sub format').selectOption(pubSub.format)
      cy.get('#uplink_message_checkbox').check()
      cy.findByLabelText('Uplink message').type(pubSub.uplinkSubTopic)
      cy.findByRole('button', { name: 'Add Pub/Sub' }).click()

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
      )
    })

    it('succeeds adding pub-sub with credentials', () => {
      const pubSub = {
        id: 'with-creds-pub-sub',
        serverUrl: 'mqtts://example.com',
        clientId: 'with-creds-client-id',
        username: 'test-username',
        subscribe_qos: 'AT_MOST_ONCE',
        publish_qos: 'AT_MOST_ONCE',
        format: 'json',
        uplinkSubTopic: 'uplink-test-sub-topic',
      }
      cy.findByLabelText('Pub/Sub ID').type(pubSub.id)
      cy.findByLabelText('Server URL').type(pubSub.serverUrl)
      cy.findByLabelText('Client ID').type(pubSub.clientId)
      cy.findByLabelText('Use credentials').check({ force: true })
      cy.findByLabelText('Username').type(pubSub.username)
      cy.findByLabelText('Subscribe QoS').selectOption(pubSub.subscribe_qos)
      cy.findByLabelText('Publish QoS').selectOption(pubSub.publish_qos)
      cy.findByLabelText('Pub/Sub format').selectOption(pubSub.format)
      cy.get('#uplink_message_checkbox').check()
      cy.findByLabelText('Uplink message').type(pubSub.uplinkSubTopic)
      cy.findByRole('button', { name: 'Add Pub/Sub' }).click()

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
      )
    })
  })

  describe('Nats', () => {
    beforeEach(() => {
      cy.login({ user_id: userId, password: user.password })
      cy.visit(`/applications/${userId}/integrations/pubsubs/add`)
      cy.findByLabelText('Nats').check()
    })
  })
})
