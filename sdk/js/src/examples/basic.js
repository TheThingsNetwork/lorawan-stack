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

import TTS from '..'

const token = 'access-token-or-api-key'
const tts = new TTS({
  authorization: {
    mode: 'key',
    key: token,
  },
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

const createApplication = async () => {
  const firstApp = await tts.Applications.create('testuser', appData)
  console.log(firstApp)
}

const getApplication = async () => {
  const firstApp = await tts.Applications.getById(appName)
  console.log(firstApp)
}

const listApplications = async () => {
  const apps = await tts.Applications.getAll()
  console.log(apps)
}

const updateApplication = async () => {
  const patch = { description: 'New description' }
  const res = await tts.Applications.updateById(appName, patch)
  console.log(res)
}

const deleteApplication = async () => {
  await tts.Applications.deleteById(appName)
  await tts.Applications.deleteById(`second-app-${hash}`)
}

const main = async () => {
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
