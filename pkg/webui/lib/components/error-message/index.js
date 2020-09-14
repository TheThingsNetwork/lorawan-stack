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
import { hasCauses, getBackendErrorRootCause, toMessageProps } from '@ttn-lw/lib/errors/utils'

const ErrorMessage = function({ content, withRootCause, className, ...rest }) {
  const baseProps = {
    className,
    firstToUpper: true,
    ...rest,
  }

  if (withRootCause && hasCauses(content)) {
    const rootProps = { ...baseProps, ...toMessageProps(content) }
    const causeProps = { ...baseProps, ...toMessageProps(getBackendErrorRootCause(content)) }

    return (
      <span className={baseProps.className}>
        <Message {...rootProps} />: <Message {...causeProps} />
      </span>
    )
  }

  const props = {
    ...toMessageProps(content),
    ...baseProps,
  }

  return <Message {...props} />
}

ErrorMessage.propTypes = {
  className: PropTypes.string,
  /**
   * Content contains the error data. It will be marshalled into a `react-intl`
   * message in case of backend errors and then output as such. Can also
   * be a usual message type, in case of frontend-defined errors.
   */
  content: PropTypes.error.isRequired,
  withRootCause: PropTypes.bool,
}

ErrorMessage.defaultProps = {
  className: undefined,
  withRootCause: false,
}

export default ErrorMessage
