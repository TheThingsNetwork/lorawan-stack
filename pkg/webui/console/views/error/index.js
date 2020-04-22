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

import Button from '@ttn-lw/components/button'
import Footer from '@ttn-lw/components/footer'

import Message from '@ttn-lw/lib/components/message'
import ErrorMessage from '@ttn-lw/lib/components/error-message'
import { withEnv } from '@ttn-lw/lib/components/env'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import Header from '@console/containers/header'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import errorMessages from '@ttn-lw/lib/errors/error-messages'
import {
  httpStatusCode,
  isUnknown as isUnknownError,
  isNotFoundError,
} from '@ttn-lw/lib/errors/utils'
import statusCodeMessages from '@ttn-lw/lib/errors/status-code-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './error.styl'

const reload = () => location.reload()

const FullViewErrorInner = function({ error, env }) {
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
          <Col md={6} sm={12}>
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

const FullViewErrorInnerWithEnv = withEnv(FullViewErrorInner)

const FullViewError = function({ error }) {
  return (
    <div className={style.wrapper}>
      <Header className={style.header} anchored />
      <FullViewErrorInnerWithEnv error={error} />
      <Footer />
    </div>
  )
}

FullViewErrorInner.propTypes = {
  env: PropTypes.env,
  error: PropTypes.error.isRequired,
}

FullViewError.propTypes = {
  error: PropTypes.error.isRequired,
}

FullViewErrorInner.defaultProps = {
  env: undefined,
}

export { FullViewError as default, FullViewErrorInnerWithEnv as FullViewErrorInner }
