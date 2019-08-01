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
import classnames from 'classnames'

import Message from '../message'

import PropTypes from '../../prop-types'
import errorMessages from '../../errors/error-messages'

import {
  isBackend,
  isTranslated,
  getBackendErrorId,
  getBackendErrorDefaultMessage,
  getBackendErrorMessageAttributes,
} from '../../errors/utils'

import style from './error-message.styl'

const ErrorMessage = function ({ content, ...rest }) {

  const props = {
    content: {},
    ...rest,
  }

  // Check if it is a error message and transform it to a intl message
  if (isBackend(content)) {
    props.content.id = getBackendErrorId(content)
    props.content.defaultMessage = getBackendErrorDefaultMessage(content)
    props.values = getBackendErrorMessageAttributes(content)
    props.className = classnames(rest.className, style.message)
  } else if (isTranslated(content)) {
    // Fall back to normal message
    props.content = content
  } else {
    // Fall back to generic error message
    props.content = errorMessages.genericError
  }

  return <Message {...props} />
}

ErrorMessage.propTypes = {
  /**
   * Content contains the error data. It will be marshalled into a `react-intl`
   * message in case of backend errors and then output as such. Can also
   * be a usual message type, in case of frontend-defined errors.
   */
  content: PropTypes.error,
}

export default ErrorMessage
