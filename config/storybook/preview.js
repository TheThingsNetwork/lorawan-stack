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

/* eslint-disable import/prefer-default-export */

import React from 'react'
import { IntlProvider } from 'react-intl'
import { MemoryRouter } from 'react-router-dom'

import messages from '@ttn-lw/locales/en.json'
import backendMessages from '@ttn-lw/locales/.backend/en.json'

import '../../pkg/webui/styles/main.styl'
import '../../pkg/webui/styles/utilities/general.styl'
import '../../pkg/webui/styles/utilities/spacing.styl'
import '../../pkg/webui/styles/utilities/tokens.styl'
import 'focus-visible/dist/focus-visible'
import Center from './center'

export const decorators = [
  Story => (
    <IntlProvider key="key" messages={{ ...messages, ...backendMessages }} locale="en-US">
      <MemoryRouter>
        <Center>{Story()}</Center>
      </MemoryRouter>
    </IntlProvider>
  ),
]
