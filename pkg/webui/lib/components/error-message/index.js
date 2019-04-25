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

import Message from '../message'

import PropTypes from '../../prop-types'
import sharedMessages from '../../shared-messages'

const ErrorMessage = function ({ content, ...rest }) {

  const props = {
    content: {},
    ...rest,
  }

  // Check if it is a error message and transform it to a intl message
  if (typeof content === 'object' && !('id' in content) && content.message && content.details) {
    props.content.id = content.message.split(' ')[0]
    props.content.defaultMessage = content.details[0].message_format || content.message.replace(/^.*\s/, '')
    props.values = content.details[0].attributes
  } else if (typeof content === 'object' && content.id && content.defaultMessage) {
    // Fall back to normal message
    props.content = content
  } else {
    // Fall back to generic error message
    props.content = sharedMessages.genericError
  }

  return <Message {...props} />
}

ErrorMessage.propTypes = {
  /**
   * Content contains the error message data, returned from the backend. It will
   * be marshalled into a `react-intl` message and then output as such. Can also
   * be a usual message type, in case of frontend-defined errors.
   */
  content: PropTypes.error,
}

export default ErrorMessage
