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

const { cypressBrowserPermissionsPlugin } = require('cypress-browser-permissions')

const tasks = require('./tasks')

module.exports = (on, config) => {
  const configWithPermissions = cypressBrowserPermissionsPlugin(on, config)

  tasks.stackConfigTask(on, configWithPermissions)
  tasks.sqlTask(on, configWithPermissions)
  tasks.fileExistsTask(on, configWithPermissions)
  tasks.emailTask(on, configWithPermissions)

  if (process.env.NODE_ENV === 'development') {
    tasks.codeCoverageTask(on, configWithPermissions)
  }

  on('before:browser:launch', (browser = {}, launchOptions) => {
    if (browser.family === 'chromium' && browser.name !== 'electron') {
      launchOptions.args.push(
        '--use-file-for-fake-video-capture=cypress/fixtures/qr-code-mock-feed.y4m',
      )
    }

    if (browser.name === 'chrome' && browser.isHeadless) {
      launchOptions.args.push('--disable-gpu')
    }

    return launchOptions
  })

  return configWithPermissions
}
