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
import classnames from 'classnames'

import { warn } from '../../log'
import PropTypes from '../../prop-types'

import style from './message.styl'

const warned = {}
const warning = function(message) {
  if (!warned[message]) {
    warned[message] = true
    warn(`Message is not translated: "${message}"`)
  }
}

const Message = function({
  content,
  values = {},
  component: Component,
  lowercase,
  uppercase,
  firstToUpper,
  firstToLower,
  capitalize,
  className,
  ...rest
}) {
  if (React.isValidElement(content)) {
    return content
  }

  if (typeof content === 'string' || typeof content === 'number') {
    warning(content)
    return (
      <Component className={className} {...rest}>
        {content}
      </Component>
    )
  }

  let vals = values
  if (content.values && Object.keys(values).length === 0) {
    vals = content.values
  }

  const cls = classnames(className, {
    [style.lowercase]: lowercase,
    [style.uppercase]: uppercase,
    [style.firstToUpper]: firstToUpper,
    [style.firstToLower]: firstToLower,
    [style.capitalize]: capitalize,
  })

  if (content.id) {
    return (
      <FormattedMessage {...content} values={vals}>
        {(...children) => (
          <Component className={cls} {...rest}>
            {children}
          </Component>
        )}
      </FormattedMessage>
    )
  }

  return null
}

Message.propTypes = {
  /** Flag specifying whether the message should be capitalized */
  capitalize: PropTypes.bool,
  /** The className to be attached to the container */
  className: PropTypes.string,
  /**
   * The wrapping element component can also be set explicitly
   * (defaults to span). This can be useful to avoid unnecessary wrapping.
   */
  component: PropTypes.node,
  /**
   * The translatable message, should be an object (with `id` or
   * `defaultMessage` key). Additionally the content can contain a `values` key,
   * containing values for the message's placeholders. A string will also work,
   * but output a warning. If a a valid dom/react element is passed it will be
   * passed through without any modifications.
   */
  content: PropTypes.message.isRequired,
  /** Flag specifying whether the first letter of the message should be
   * transformed to lowercase
   */
  firstToLower: PropTypes.bool,
  /**
   * Flag specifying whether the first letter of the message should be
   * transformed to uppercase
   */
  firstToUpper: PropTypes.bool,
  /**
   * Flag specifying whether the the message should be transformed to
   * lowercase
   */
  lowercase: PropTypes.bool,
  /**
   * Flag specifying whether the the message should be transformed to
   * uppercase
   */
  uppercase: PropTypes.bool,
  /** Values can also be given as a separate property (will have precedence) */
  values: PropTypes.shape({}),
}

Message.defaultProps = {
  capitalize: false,
  className: undefined,
  component: 'span',
  firstToLower: false,
  firstToUpper: false,
  lowercase: false,
  uppercase: false,
  values: undefined,
}

export default Message
