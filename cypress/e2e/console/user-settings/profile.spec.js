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

const user = {
  ids: { user_id: 'profile-settings-test-user' },
  name: 'Test User',
  primary_email_address: 'test-user@example.com',
  password: 'ABCDefg123!',
  password_confirm: 'ABCDefg123!',
}

describe('User settings / profile', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  it('displays UI elements in place', () => {
    cy.createUser(user)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/profile`)

    cy.findByText('Profile picture').should('be.visible')
    cy.findByLabelText('Use Gravatar').should('exist').and('be.checked')
    cy.findByLabelText('Upload an image').should('exist').and('not.be.checked')
    cy.findByLabelText('User ID')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', user.ids.user_id)
    cy.findByLabelText('Name').should('be.visible')
    cy.findByLabelText('Email address')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', user.primary_email_address)

    cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
    cy.findByRole('button', { name: /Delete account/ }).should('be.visible')
  })

  it('succeeds changing profile information', () => {
    const userUpdate = {
      name: 'Jane Doe',
      primary_email_address: 'jane.doe@example.com',
    }
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/profile`)

    cy.findByLabelText('Use Gravatar').check()
    cy.findByLabelText('Name').type(userUpdate.name)
    cy.findByLabelText('Email address').clear()
    cy.findByLabelText('Email address').type(userUpdate.primary_email_address)

    // Check if the profile picture (preview) was updated properly.
    cy.get('form').within(() => {
      cy.findByAltText('Profile picture')
        .should('be.visible')
        .and('have.attr', 'src')
        // `jane.doe@example.com` has no gravatar image associated, so the `src`
        // is expected to be the src of the missing profile picture placeholder.
        .and('match', /missing-profile-picture/)
    })

    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .findByText('Profile updated')
      .should('be.visible')
  })

  it('succeeds using an uploaded profile picture', () => {
    const imageFile = 'test-image.png'

    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/profile`)

    cy.findByLabelText('Upload an image').check()

    // Upload the test image as profile picture.
    cy.findByLabelText('Image upload').attachFile(imageFile)
    cy.findByAltText('Current image')
      .should('be.visible')
      .and('have.attr', 'src')
      .then(src => {
        cy.fixture(imageFile).then(img => {
          expect(src).to.equal(`data:image/png;base64,${img}`)
        })
      })

    cy.findByRole('button', { name: 'Save changes' }).click()
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .findByText('Profile updated')
      .should('be.visible')

    // Check if the profile picture (preview) was updated properly.
    cy.findByAltText('Current image')
      .should('be.visible')
      .and('have.attr', 'src')
      .and('match', /^\/assets\/blob\/profile_pictures\/profile-settings-test-user.*\.png/)
  })

  it('succeeds deleting the account', () => {
    cy.intercept('DELETE', `/api/v3/users/${user.ids.user_id}`).as('delete-user')
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/profile`)

    cy.findByRole('button', { name: /Delete account/ }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Confirm deletion', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Delete account/ }).should('have.attr', 'disabled')
        cy.findByTestId('confirm_deletion').type(user.ids.user_id)
        cy.findByRole('button', { name: /Delete account/ }).click()
      })

    cy.wait('@delete-user')

    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should('eq', `${Cypress.config('accountAppRootPath')}/login`)
    cy.findByTestId('notification')
      .should('be.visible')
      .findByText(`Account deleted`)
      .should('be.visible')
  })
})

describe('User settings / profile with disabled upload', () => {
  before(() => {
    cy.dropAndSeedDatabase()
    cy.augmentIsConfig({ profile_picture: { disable_upload: true } })
  })
  it('displays UI elements in place', () => {
    cy.createUser(user)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/profile`)

    cy.findByText('Gravatar image').should('be.visible')
    cy.findByText('Upload image', { exact: false }).should('not.exist')
    cy.findByTestId('notification')
      .should('be.visible')
      .findByText(/follow the instructions on the Gravatar website to change your profile picture/)
      .should('be.visible')
    cy.findByLabelText('User ID')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', user.ids.user_id)
    cy.findByLabelText('Name').should('be.visible')
    cy.findByLabelText('Email address')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', user.primary_email_address)

    cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
    cy.findByRole('button', { name: /Delete account/ }).should('be.visible')
  })
})

describe('User settings / profile without gravatar', () => {
  before(() => {
    cy.dropAndSeedDatabase()
    cy.augmentIsConfig({ profile_picture: { use_gravatar: false } })
  })
  it('displays UI elements in place', () => {
    cy.createUser(user)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/profile`)

    cy.findByText('Gravatar', { exact: false }).should('not.exist')
    cy.findByLabelText('Image upload').should('be.visible')
    cy.findByTestId('notification').should('not.exist')
    cy.findByLabelText('User ID')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', user.ids.user_id)
    cy.findByLabelText('Name').should('be.visible')
    cy.findByLabelText('Email address')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', user.primary_email_address)

    cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
    cy.findByRole('button', { name: /Delete account/ }).should('be.visible')
  })
})

describe('User settings / profile without gravatar and upload', () => {
  before(() => {
    cy.dropAndSeedDatabase()
    cy.augmentIsConfig({ profile_picture: { use_gravatar: false, disable_upload: true } })
  })

  it('displays UI elements in place', () => {
    cy.createUser(user)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/profile`)

    cy.findByText('Gravatar', { exact: false }).should('not.exist')
    cy.findByText('Upload image', { exact: false }).should('not.exist')
    cy.findByTestId('notification')
      .should('be.visible')
      .findByText(/profile picture.*disabled/)
      .should('be.visible')
    cy.findByLabelText('User ID')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', user.ids.user_id)
    cy.findByLabelText('Name').should('be.visible')
    cy.findByLabelText('Email address')
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', user.primary_email_address)

    cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
    cy.findByRole('button', { name: /Delete account/ }).should('be.visible')
  })
})
