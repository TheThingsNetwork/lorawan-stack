// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import Icon, { IconExclamationCircle } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'
import ErrorMessage from '@ttn-lw/lib/components/error-message'

import PropTypes from '@ttn-lw/lib/prop-types'
import { isBackend, isNotFoundError, httpStatusCode } from '@ttn-lw/lib/errors/utils'
import errorMessages from '@ttn-lw/lib/errors/error-messages'
import statusCodeMessages from '@ttn-lw/lib/errors/status-code-messages'

import style from './sub-view.styl'

const SubViewErrorComponent = ({ error }) => {
  const isNotFound = isNotFoundError(error)
  const statusCode = httpStatusCode(error)
  let errorExplanation = errorMessages.subviewErrorExplanation
  let errorTitleMessage = errorMessages.subviewErrorTitle
  if (isNotFound) {
    errorExplanation = errorMessages.genericNotFound
  }
  if (statusCode) {
    errorTitleMessage = statusCodeMessages[statusCode]
  }

  return (
    <div className="container container--lg">
      <div className={style.title}>
        <Icon icon={IconExclamationCircle} large />
        <Message component="h2" content={errorTitleMessage} firstToUpper />
      </div>
      <p>
        <Message component="span" content={errorExplanation} />
        <br />
        <Message component="span" content={errorMessages.contactAdministrator} />
      </p>
      {isBackend(error) && (
        <React.Fragment>
          <hr />
          <ErrorMessage content={error} />
        </React.Fragment>
      )}
    </div>
  )
}

SubViewErrorComponent.propTypes = {
  error: PropTypes.error.isRequired,
}

const subViewErrorRender = error => <SubViewErrorComponent error={error} />

export default subViewErrorRender
