// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import { FormattedMessage } from 'react-intl'

import { warn } from '../../lib/log'

const warned = {}
const warning = function (message) {
  if (!warned[message]) {
    warned[message] = true
    warn(`Message is not translated: "${message}"`)
  }
}

const Message = function ({
  content,
  values = {},
  component: Component = 'span',
  ...rest
}) {

  if (content === null) {
    return content
  }

  if (React.isValidElement(content)) {
    return content
  }

  if (typeof content === 'string') {
    warning(content)
    return <Component {...rest}>{content}</Component>
  }

  if (content.id) {
    return (
      <FormattedMessage {...content} values={values}>
        {(...children) => <Component {...rest}>{children}</Component>}
      </FormattedMessage>
    )
  }
}

export default Message
