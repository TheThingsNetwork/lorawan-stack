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
import { FormattedMessage } from 'react-intl'

import { warn } from '../../log'
import PropTypes from '../../prop-types'

const warned = {}
const warning = function(message) {
  if (!warned[message]) {
    warned[message] = true
    warn(`Message is not translated: "${message}"`)
  }
}

const Message = function({ content, values = {}, component: Component = 'span', ...rest }) {
  if (!content && content !== 0) {
    return null
  }

  if (React.isValidElement(content)) {
    return content
  }

  if (typeof content === 'string' || typeof content === 'number') {
    warning(content)
    return <Component {...rest}>{content}</Component>
  }

  let vals = values
  if (content.values && Object.keys(values).length === 0) {
    vals = content.values
  }

  if (content.id) {
    return (
      <FormattedMessage {...content} values={vals}>
        {(...children) => <Component {...rest}>{children}</Component>}
      </FormattedMessage>
    )
  }

  return null
}

Message.propTypes = {
  /**
   * The translatable message, should be an object (with `id` or
   * `defaultMessage` key). Additionally the content can contain a `values` key,
   * containing values for the message's placeholders. A string will also work,
   * but output a warning. If a a valid dom/react element is passed it will be
   * passed through without any modifications.
   */
  content: PropTypes.message,
  /** Values can also be given as a separate property (will have precedence) */
  values: PropTypes.object,
  /**
   * The wrapping element component can also be set explicitly
   * (defaults to span). This can be useful to avoid unnecessary wrapping.
   */
  component: PropTypes.node,
}

export default Message
