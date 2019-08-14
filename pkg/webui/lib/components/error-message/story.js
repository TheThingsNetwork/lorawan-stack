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

import React from 'react'
import { storiesOf } from '@storybook/react'
import { withInfo } from '@storybook/addon-info'

import ErrorMessage from '.'

const exampleError = {
  code: 2,
  message:
    'error:pkg/assets:http (HTTP error: `` is not a valid ID. Must be at least 2 and at most 36 characters long and may consist of only letters, numbers and dashes. It may not start or end with a dash)',
  details: [
    {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
      namespace: 'pkg/assets',
      name: 'http',
      message_format: 'HTTP error: {message}',
      attributes: {
        message:
          '`` is not a valid ID. Must be at least 2 and at most 36 characters long and may consist of only letters, numbers and dashes. It may not start or end with a dash',
      },
    },
  ],
}

storiesOf('Utility Components/ErrorMessage', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      propTables: [ErrorMessage],
    })(story)(context),
  )
  .add('Default', () => <ErrorMessage content={exampleError} />)
