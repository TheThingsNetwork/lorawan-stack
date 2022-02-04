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

import React, { useContext } from 'react'
import { FormattedMessage, IntlContext } from 'react-intl'
import classnames from 'classnames'
import reactStringReplace from 'react-string-replace'
import { isPlainObject } from 'lodash'

import PropTypes from '@ttn-lw/lib/prop-types'
import { warn } from '@ttn-lw/lib/log'
import interpolate from '@ttn-lw/lib/interpolate'

import style from './message.styl'

const renderContent = (content, component, props) => {
  const Component = component

  if (!Boolean(component)) {
    if (Boolean(props.className)) {
      return <span {...props}>{content}</span>
    }

    return <React.Fragment key={props.key}>{content}</React.Fragment>
  }

  if (Boolean(Component) || Boolean(props.className)) {
    return <Component {...props}>{content}</Component>
  }

  return content
}

const Message = ({
  content,
  values = {},
  component,
  lowercase,
  uppercase,
  firstToUpper,
  firstToLower,
  capitalize,
  className,
  convertBackticks,
  ...rest
}) => {
  const cls = classnames(className, {
    [style.lowercase]: lowercase,
    [style.uppercase]: uppercase,
    [style.firstToUpper]: firstToUpper,
    [style.firstToLower]: firstToLower,
    [style.capitalize]: capitalize,
  })

  if (cls) {
    rest.className = cls
  }

  const intlContext = useContext(IntlContext)

  if (typeof content === 'string' || typeof content === 'number') {
    return renderContent(content, component, rest)
  }

  if (!isPlainObject(content)) {
    // Better to render nothing rather than throwing an error
    // that potentially causes the whole view to crash.
    return renderContent(null, component, rest)
  }

  let vals = values
  if (content.values && Object.keys(values).length === 0) {
    vals = content.values
  }

  if (!Boolean(intlContext)) {
    // Displaying the default message is a last resort that should only
    // be considered when all other options failed. Note also that this
    // will only do basic value interpolation! It's still better than not
    // showing anything at all or crashing altogether.
    warn(
      'Attempting to render a <Message /> without Intl context. Falling back to default message!',
      content,
    )
    return renderContent(interpolate(content.defaultMessage, vals), component, rest)
  }

  const { formatMessage } = intlContext

  if (convertBackticks) {
    const contentWithMarkdown = formatMessage(content, vals)
    return renderContent(
      reactStringReplace(contentWithMarkdown, /`([^`[]+)`/g, (match, i) => (
        <code key={i}>{match}</code>
      )),
      component,
      rest,
    )
  }

  return (
    <FormattedMessage {...content} values={vals}>
      {(...children) => renderContent(children, component, rest)}
    </FormattedMessage>
  )
}

Message.propTypes = {
  /** Flag specifying whether the message should be capitalized. */
  capitalize: PropTypes.bool,
  /** The className to be attached to the container. */
  className: PropTypes.string,
  /**
   * The wrapping element component can also be set explicitly (defaults to
   * span). This can be useful to avoid unnecessary wrapping.
   */
  component: PropTypes.oneOfType([PropTypes.node, PropTypes.elementType]),
  /**
   * The translatable message, should be an object (with `id` or
   * `defaultMessage` key). Additionally the content can contain a `values` key,
   * containing values for the message's placeholders. A string will also work,
   * but output a warning. If a a valid dom/react element is passed it will be
   * passed through without any modifications.
   */
  content: PropTypes.message.isRequired,
  /**
   * Flag specifying whether the the backticks should be converted to `<code/>` tag.
   */
  convertBackticks: PropTypes.bool,
  /** Flag specifying whether the first letter of the message should be
   * transformed to lowercase.
   */
  firstToLower: PropTypes.bool,
  /**
   * Flag specifying whether the first letter of the message should be
   * transformed to uppercase.
   */
  firstToUpper: PropTypes.bool,
  /**
   * Flag specifying whether the the message should be transformed to
   * lowercase.
   */
  lowercase: PropTypes.bool,
  /**
   * Flag specifying whether the the message should be transformed to
   * uppercase.
   */
  uppercase: PropTypes.bool,
  /** Values can also be given as a separate property (will have precedence). */
  values: PropTypes.shape({}),
}

Message.defaultProps = {
  capitalize: false,
  className: undefined,
  component: undefined,
  convertBackticks: true,
  firstToLower: false,
  firstToUpper: false,
  lowercase: false,
  uppercase: false,
  values: undefined,
}
export default Message
