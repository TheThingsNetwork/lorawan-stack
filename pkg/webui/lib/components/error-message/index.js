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

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import { toMessageProps, isBackend } from '@ttn-lw/lib/errors/utils'

const ErrorMessage = ({ content, withRootCause, useTopmost, className, ...rest }) => {
  const baseProps = {
    className,
    firstToUpper: true,
    convertBackticks: Boolean(isBackend(content)),
    ...rest,
  }

  const messageProps = toMessageProps(content, true)

  if (withRootCause && messageProps.length > 1) {
    return (
      <span className={baseProps.className}>
        <Message {...baseProps} {...messageProps[1]} />:{' '}
        <Message {...baseProps} {...messageProps[0]} />
      </span>
    )
  }

  const index = useTopmost ? messageProps.length - 1 : 0

  return <Message {...baseProps} {...messageProps[index]} />
}

ErrorMessage.propTypes = {
  className: PropTypes.string,
  /**
   * Content contains the error data. It will be marshalled into a `react-intl`
   * message in case of backend errors and then output as such. Can also
   * be a usual message type, in case of frontend-defined errors.
   */
  content: PropTypes.error.isRequired,
  useTopmost: PropTypes.bool,
  withRootCause: PropTypes.bool,
}

ErrorMessage.defaultProps = {
  className: undefined,
  useTopmost: false,
  withRootCause: false,
}

export default ErrorMessage
