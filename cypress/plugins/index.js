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

const fs = require('fs')

const tasks = require('./tasks')

const failedSpecsFilename = `./.cache/.failed-specs-${
  process.env.CYPRESS_MACHINE_NUMBER || '0'
}.txt`

module.exports = (on, config) => {
  tasks.stackConfigTask(on, config)
  tasks.sqlTask(on, config)
  tasks.stackLogTask(on, config)
  tasks.fileExistsTask(on, config)

  if (process.env.NODE_ENV === 'development') {
    tasks.codeCoverageTask(on, config)
  }

  on('before:browser:launch', (browser = {}, launchOptions) => {
    if (browser.name === 'chrome' && browser.isHeadless) {
      launchOptions.args.push('--disable-gpu')
    }

    return launchOptions
  })

  return config
}
