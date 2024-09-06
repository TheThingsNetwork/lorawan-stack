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

/* eslint-disable no-invalid-this */

import '@testing-library/cypress/add-commands'
import { configure } from '@testing-library/cypress'
import './commands'

const failedSpecsFilename = `./.cache/.failed-specs-${Cypress.env('MACHINE_NUMBER') || '0'}.txt`

configure({ testIdAttribute: 'data-test-id' })
Cypress.SelectorPlayground.defaults({
  selectorPriority: ['data-test-id', 'id', 'class', 'tag', 'attributes', 'nth-child'],
})

afterEach(function () {
  // Enable fail-early, if set.:
  if (this.currentTest.state === 'failed' && Cypress.env('FAIL_FAST')) {
    cy.log('Skipping rest of run due to test failure (fail fast)')
    const file = this.currentTest?.invocationDetails?.relativeFile
    // Sometimes `invocationDetails` is not set, see:
    // https://github.com/cypress-io/cypress/issues/3090#issuecomment-1068059581
    // In that case, we should just skip the fail fast logic and run the
    // remaining tests as usual.
    if (!file) {
      cy.log('Skipping fail fast logic due to missing `invocationDetails`')
      return
    }
    // The file will be relative to the `./config` directory, so we need to
    // remove the `../` prefix.
    const relativeFile = file.replace(/^\.\.\//, '')
    cy.writeFile(failedSpecsFilename, relativeFile)
  } else {
    // Apply a workaround for requests spilling over to the subsequent test.
    // See also https://github.com/cypress-io/cypress/issues/686.
    cy.window().then(win => {
      win.location.href = 'about:blank'
    })
  }
})

// Skip remaining runs if fail early is set.
const skipIfNecessary = function () {
  cy.task('fileExists', failedSpecsFilename).then(content => {
    if (content !== '' && content !== false && Cypress.env('FAIL_FAST')) {
      this.currentTest.pending = true
      Cypress.runner.stop()
    }
  })
}

before(skipIfNecessary)

beforeEach(() => {
  skipIfNecessary()
})
