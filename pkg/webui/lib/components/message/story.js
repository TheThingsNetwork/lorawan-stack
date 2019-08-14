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
import { IntlProvider } from 'react-intl'

import doc from './message.md'
import Message from '.'

const exampleMessage = {
  id: '$.some.id',
  defaultMessage: 'This is the default message',
}

const placeholderMessage = {
  id: '$.placeholder',
  defaultMessage: 'There are {number} gateways.',
}

const messages = {
  '$.some.id': 'This is the translated message',
  '$.placeholder': 'There are {number} gateways.',
}

const IntlDecorator = storyFn => (
  <IntlProvider key="key" messages={messages} locale="en-US">
    {storyFn()}
  </IntlProvider>
)

storiesOf('Utility Components/Message', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      text: doc,
      propTables: [Message],
    })(story)(context),
  )
  .addDecorator(IntlDecorator)
  .add('Default', () => <Message content={exampleMessage} />)
  .add('Placeholder', () => <Message content={placeholderMessage} values={{ number: 5 }} />)
  .add('String', () => <Message content="I can also be just a string, but will issue a warning" />)
