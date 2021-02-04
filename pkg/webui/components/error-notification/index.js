// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import React, { useEffect } from 'react'

import Notification from '@ttn-lw/components/notification'

import { isBackend, toMessageProps, ingestError } from '@ttn-lw/lib/errors/utils'
import PropTypes from '@ttn-lw/lib/prop-types'

const ErrorNotification = ({ content, title, ...rest }) => {
  const message = toMessageProps(content)
  let details = undefined

  useEffect(() => {
    ingestError(content, { ingestedBy: 'ErrorNotification' })
  }, [content])

  if (isBackend(content)) {
    details = content
  }
  return (
    <Notification
      error
      content={message.content}
      title={title || message.title}
      messageValues={message.values}
      details={details}
      data-test-id="error-notification"
      {...rest}
    />
  )
}

ErrorNotification.propTypes = {
  content: PropTypes.oneOfType([PropTypes.message, PropTypes.error, PropTypes.string]).isRequired,
  title: PropTypes.message,
}

ErrorNotification.defaultProps = {
  title: undefined,
}

export default ErrorNotification
