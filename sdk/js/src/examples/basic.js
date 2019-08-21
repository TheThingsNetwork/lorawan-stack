// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

/* eslint-disable no-console */

import TTN from '../.'

const token = 'access-token-or-api-key'
const ttn = new TTN(token, {
  connectionType: 'http',
  baseURL: 'http://localhost:1885/api/v3',
  defaultUserId: 'testuser',
})

const hash = new Date().valueOf()
const appName = `first-app-${hash}`
const appData = {
  ids: {
    application_id: appName,
  },
  name: 'Test App',
  description: `Some description ${hash}`,
}

async function createApplication() {
  // Via Applications object
  const firstApp = await ttn.Applications.create('testuser', appData)
  console.log(firstApp)

  // Via Application class
  appData.ids.application_id = `second-app-${hash}`
  const secondApp = new ttn.Application(appData)
  await secondApp.save()
  console.log(secondApp)
}

async function getApplication() {
  const firstApp = await ttn.Applications.getById(appName)
  console.log(firstApp)
}

async function listApplications() {
  const apps = await ttn.Applications.getAll()
  console.log(apps)
}

async function updateApplication() {
  // Via Applications object
  const patch = { description: 'New description' }
  const res = await ttn.Applications.updateById(appName, patch)
  console.log(res)

  // Via Application instance
  const app = await ttn.Applications.getById(appName)
  app.description = 'Another description'
  await app.save()
  console.log(app)
}

async function deleteApplication() {
  await ttn.Applications.deleteById(appName)
  await ttn.Applications.deleteById(`second-app-${hash}`)
}

async function main() {
  try {
    await createApplication()
    await getApplication()
    await listApplications()
    await updateApplication()
    await deleteApplication()
  } catch (err) {
    console.log(err)
  }
}

main()
