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

import inquirer from 'inquirer'
import fs from 'fs-extra'

import createHeaderFile from './utils/create-header-file.js'
import { insertAtLineAtIndex } from './utils/insert-at-line.js'

const camelCaseToUpperSnakeCase = str => str.replace(/([A-Z])/g, '_$1').toUpperCase()
const snakeCaseToCamelCase = str =>
  str.replace(/([-_][a-z])/g, group => group.toUpperCase().replace('-', '').replace('_', ''))
const snakeCaseToPascalCase = str =>
  snakeCaseToCamelCase(str).replace(/(^[a-z])/g, group => group.toUpperCase())

let originalReducerIndexContent
let originalMiddlewareIndexContent

const generateActionFile = actionName => {
  const actionNameUpperSnakeCase = camelCaseToUpperSnakeCase(actionName)
  return `import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'

export const ${actionNameUpperSnakeCase}_BASE = '${actionNameUpperSnakeCase}'
export const [
  {
    request: ${actionNameUpperSnakeCase},
    success: ${actionNameUpperSnakeCase}_SUCCESS,
    failure: ${actionNameUpperSnakeCase}_FAILURE,
  },
  {
    request: ${actionName},
    success: ${actionName}Success,
    failure: ${actionName}Failure,
  },
] = createRequestActions(${actionNameUpperSnakeCase}_BASE, (params) => ({ ...params }))
`
}

const generateReducerFile = (
  storeFileName,
  actionName,
) => `import { handleActions } from 'redux-actions'

import { ${camelCaseToUpperSnakeCase(actionName)}_SUCCESS } from '@console/store/actions/${storeFileName}'

const defaultState = {
  data: [],
}

export default handleActions(
  {
    [${camelCaseToUpperSnakeCase(actionName)}_SUCCESS]: (state, { payload }) => ({
      ...state,
      data: payload.data,
    }),
  },
  defaultState,
)
`

const generateSelectorFile =
  actionName => `const select${snakeCaseToPascalCase(actionName)}Store = state => state.${actionName}

export const select${snakeCaseToPascalCase(actionName)}Data = state => select${snakeCaseToPascalCase(actionName)}Store(state).data
`

const generateMiddlewareFile = (
  storeFileName,
  actionName,
) => `import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import * as actions from '@console/store/actions/${storeFileName}'

const ${actionName}Logic = createRequestLogic({
  type: actions.${camelCaseToUpperSnakeCase(actionName)},
  process: async ({ action }, dispatch, done) => {
    // Process logic here
    done()
  }
})

export default [${actionName}Logic]
`

const main = async () => {
  const answers = await inquirer.prompt([
    {
      type: 'input',
      name: 'storeFileName',
      message: 'Enter the name of the store file: (eg. device)',
    },
    {
      type: 'input',
      name: 'actionName',
      message: 'Enter the base name for the action: (eg. getDevice)',
    },
    {
      type: 'checkbox',
      name: 'filesToGenerate',
      message: 'Which files do you want to generate?',
      choices: ['actions', 'reducers', 'selectors', 'middleware'],
      default: ['actions', 'reducers', 'selectors', 'middleware'],
    },
  ])

  const { storeFileName, actionName, filesToGenerate } = answers

  if (filesToGenerate.includes('actions')) {
    await createHeaderFile(
      generateActionFile(actionName),
      `pkg/webui/console/store/actions/${storeFileName}.js`,
    )
  }
  if (filesToGenerate.includes('reducers')) {
    await createHeaderFile(
      generateReducerFile(storeFileName, actionName),
      `pkg/webui/console/store/reducers/${storeFileName}.js`,
    )
    // Add the reducer to the index file
    const reducerIndexFile = `pkg/webui/console/store/reducers/index.js`
    originalReducerIndexContent = await fs.readFile(reducerIndexFile, 'utf8')
    let reducerIndexContent = originalReducerIndexContent
    // Find the last import statement and add the new import statement after it
    const lastImportIndex = reducerIndexContent.lastIndexOf('import ')
    const importStatement = `import ${snakeCaseToCamelCase(storeFileName)} from './${storeFileName}'\n`
    reducerIndexContent = insertAtLineAtIndex(reducerIndexContent, lastImportIndex, importStatement)
    // Find the last item of the exported array and add the new reducer after it
    const lastExportIndex = reducerIndexContent.lastIndexOf('})')
    const reducerExport = `  ${snakeCaseToCamelCase(storeFileName)},\n`
    reducerIndexContent = insertAtLineAtIndex(reducerIndexContent, lastExportIndex, reducerExport)

    await fs.writeFile(reducerIndexFile, reducerIndexContent)
  }
  if (filesToGenerate.includes('selectors')) {
    await createHeaderFile(
      generateSelectorFile(actionName),
      `pkg/webui/console/store/selectors/${storeFileName}.js`,
    )
  }
  if (filesToGenerate.includes('middleware')) {
    await createHeaderFile(
      generateMiddlewareFile(storeFileName, actionName),
      `pkg/webui/console/store/middleware/logics/${storeFileName}.js`,
    )
    // Add the middleware to the index file
    const middlewareIndexFile = `pkg/webui/console/store/middleware/logics/index.js`
    originalMiddlewareIndexContent = await fs.readFile(middlewareIndexFile, 'utf8')
    let middlewareIndexContent = originalMiddlewareIndexContent
    // Find the last import statement and add the new import statement after it
    const lastImportIndex = middlewareIndexContent.lastIndexOf('import ')
    const importStatement = `import ${snakeCaseToCamelCase(storeFileName)} from './${storeFileName}'\n`
    middlewareIndexContent = insertAtLineAtIndex(
      middlewareIndexContent,
      lastImportIndex,
      importStatement,
    )
    // Find the last item of the exported array and add the new middleware after it
    const lastExportIndex = middlewareIndexContent.lastIndexOf(']')
    const middlewareExport = `  ...${snakeCaseToCamelCase(storeFileName)},\n`
    middlewareIndexContent = insertAtLineAtIndex(
      middlewareIndexContent,
      lastExportIndex,
      middlewareExport,
    )

    await fs.writeFile(middlewareIndexFile, middlewareIndexContent)
  }

  // Allow to delete the generated files
  const { deleteFiles } = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'deleteFiles',
      message:
        'Please check the files. If there is an issue, you can choose to undo the changes. Do you want to undo?',
      default: false,
    },
  ])

  if (deleteFiles) {
    if (filesToGenerate.includes('actions')) {
      await fs.remove(`pkg/webui/console/store/actions/${storeFileName}.js`)
    }
    if (filesToGenerate.includes('reducers')) {
      await fs.remove(`pkg/webui/console/store/reducers/${storeFileName}.js`)
      await fs.writeFile(`pkg/webui/console/store/reducers/index.js`, originalReducerIndexContent)
    }
    if (filesToGenerate.includes('selectors')) {
      await fs.remove(`pkg/webui/console/store/selectors/${storeFileName}.js`)
    }
    if (filesToGenerate.includes('middleware')) {
      await fs.remove(`pkg/webui/console/store/middleware/logics/${storeFileName}.js`)
      await fs.writeFile(
        `pkg/webui/console/store/middleware/index.js`,
        originalMiddlewareIndexContent,
      )
    }
  }
}

export default main
