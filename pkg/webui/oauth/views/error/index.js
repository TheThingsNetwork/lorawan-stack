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
import { Container, Row, Col } from 'react-grid-system'

import Button from '../../../components/button'
import Message from '../../../lib/components/message'
import ErrorMessage from '../../../lib/components/error-message'
import { withEnv } from '../../../lib/components/env'
import IntlHelmet from '../../../lib/components/intl-helmet'
import sharedMessages from '../../../lib/shared-messages'
import errorMessages from '../../../lib/errors/error-messages'

import {
  httpStatusCode,
  isUnknown as isUnknownError,
  isNotFoundError,
} from '../../../lib/errors/utils'

import statusCodeMessages from '../../../lib/errors/status-code-messages'

import style from './full-view.styl'

const reload = () => location.reload()

const FullViewError = function({ error, env }) {
  const isUnknown = isUnknownError(error)
  const statusCode = httpStatusCode(error)
  const isNotFound = isNotFoundError(error)

  let errorTitleMessage = errorMessages.unknownErrorTitle
  let errorMessageMessage = errorMessages.contactAdministrator
  if (!isUnknown) {
    errorMessageMessage = error
  } else if (isNotFound) {
    errorMessageMessage = errorMessages.genericNotFound
  }
  if (statusCode) {
    errorTitleMessage = statusCodeMessages[statusCode]
  }

  return (
    <div className={style.fullViewError}>
      <Container>
        <Row>
          <Col>
            <IntlHelmet title={errorMessages.error} />
            <Message
              className={style.fullViewErrorHeader}
              component="h2"
              content={errorTitleMessage}
            />
            <ErrorMessage className={style.fullViewErrorSub} content={errorMessageMessage} />
            {isNotFoundError(error) ? (
              <Button.AnchorLink
                icon="keyboard_arrow_left"
                message={sharedMessages.takeMeBack}
                href={env.appRoot}
              />
            ) : (
              <Button icon="refresh" message={sharedMessages.refreshPage} onClick={reload} />
            )}
          </Col>
        </Row>
      </Container>
    </div>
  )
}

export default withEnv(FullViewError)
