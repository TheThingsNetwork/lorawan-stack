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

import path from 'path'

import fs from 'fs-extra'
import inquirer from 'inquirer'

const header = `// Copyright © ${new Date().getFullYear()} The Things Network Foundation, The Things Industries B.V.
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

`

const createHeaderFile = async (content, filePath) => {
  const dir = path.dirname(filePath)
  await fs.ensureDir(dir)
  const fileContent = `${header}${content}`
  const fileExists = await fs.pathExists(filePath)
  if (fileExists) {
    const { overwrite } = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'overwrite',
        message: `File ${filePath} already exists. Overwrite?`,
        default: false,
      },
    ])
    if (!overwrite) {
      console.log('File not created.')
      return
    }
  }

  await fs.writeFile(filePath, fileContent)
  console.log(`File created: ${filePath}`)
}

export default createHeaderFile
