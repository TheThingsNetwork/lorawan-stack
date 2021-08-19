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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Container, Row, Col } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import clipboard from 'clipboard'
import { Helmet } from 'react-helmet'
import classnames from 'classnames'

import Link from '@ttn-lw/components/link'
import Footer from '@ttn-lw/components/footer'
import Button from '@ttn-lw/components/button'
import buttonStyle from '@ttn-lw/components/button/button.styl'

import Message from '@ttn-lw/lib/components/message'
import ErrorMessage from '@ttn-lw/lib/components/error-message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import errorMessages from '@ttn-lw/lib/errors/error-messages'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import {
  httpStatusCode,
  isUnknown as isUnknownError,
  isNotFoundError,
  isFrontend as isFrontendError,
} from '@ttn-lw/lib/errors/utils'
import statusCodeMessages from '@ttn-lw/lib/errors/status-code-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import {
  selectApplicationRootPath,
  selectSupportLinkConfig,
  selectApplicationSiteName,
  selectApplicationSiteTitle,
} from '@ttn-lw/lib/selectors/env'

import style from './error.styl'

const m = defineMessages({
  errorDetails: 'Error details',
})

const appRoot = selectApplicationRootPath()
const siteName = selectApplicationSiteName()
const siteTitle = selectApplicationSiteTitle()
const supportLink = selectSupportLinkConfig()

const FullViewErrorInner = ({ error }) => {
  const isUnknown = isUnknownError(error)
  const statusCode = httpStatusCode(error)
  const isNotFound = isNotFoundError(error)
  const isFrontend = isFrontendError(error)

  const [copied, setCopied] = useState(false)

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
  if (isFrontend) {
    errorMessageMessage = error.errorMessage
    if (Boolean(error.errorTitle)) {
      errorTitleMessage = error.errorTitle
    }
  }

  const copiedTimer = useRef(undefined)
  const handleCopyClick = useCallback(() => {
    if (!copied) {
      setCopied(true)
      copiedTimer.current = setTimeout(() => setCopied(false), 3000)
    }
  }, [setCopied, copied])

  const copyButton = useRef(null)
  useEffect(() => {
    if (copyButton.current) {
      new clipboard(copyButton.current)
    }
    return () => {
      // Clear timer on unmount.
      if (copiedTimer.current) {
        clearTimeout(copiedTimer.current)
      }
    }
  }, [])

  let action = undefined
  if (isNotFound) {
    action = (
      <Link.Anchor icon="keyboard_arrow_left" href={appRoot} primary>
        <Message content={sharedMessages.backToOverview} />
      </Link.Anchor>
    )
  } else {
    const errorDetails = JSON.stringify(error, undefined, 2)

    action = (
      <details>
        <summary>
          <Message content={m.errorDetails} />
        </summary>
        <pre>{errorDetails}</pre>
        <Button
          onClick={handleCopyClick}
          ref={copyButton}
          data-clipboard-text={errorDetails}
          message={copied ? sharedMessages.copiedToClipboard : sharedMessages.copyToClipboard}
          icon={copied ? 'done' : 'file_copy'}
          secondary
        />
        {Boolean(supportLink) && (
          <Button.AnchorLink
            href={supportLink}
            message={sharedMessages.getSupport}
            icon="contact_support"
            target="_blank"
            secondary
          />
        )}
      </details>
    )
  }

  return (
    <div className={style.fullViewError} data-test-id="full-error-view">
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
            {action}
          </Col>
        </Row>
      </Container>
    </div>
  )
}

const FullViewError = ({ error, header, onlineStatus }) => (
  <div className={style.wrapper}>
    {Boolean(header) && header}
    <div className={style.flexWrapper}>
      <FullViewErrorInner error={error} />
    </div>
    <Footer onlineStatus={onlineStatus} />
  </div>
)

const FullViewErrorRaw = ({ error }) => {
  const isUnknown = isUnknownError(error)
  const statusCode = httpStatusCode(error)
  const isNotFound = isNotFoundError(error)

  let errorTitle = errorMessages.unknownErrorTitle.defaultMessage
  let errorMessage = errorMessages.contactAdministrator.defaultMessage
  if (!isUnknown) {
    errorMessage = error
  } else if (isNotFound) {
    errorMessage = errorMessages.genericNotFound.defaultMessage
  }
  if (statusCode) {
    errorTitle = statusCodeMessages[statusCode].defaultMessage
  }

  const errorDetails = JSON.stringify(error, undefined, 2)

  return (
    <div className={style.fullViewError} data-test-id="full-error-view">
      <Container>
        <Row>
          <Col sm={12} md={6}>
            <Helmet titleTemplate={`%s - ${siteTitle ? `${siteTitle} - ` : ''}${siteName}`}>
              <title>Error</title>
            </Helmet>
            <h2>{errorTitle}</h2>
            <span className={style.fullViewErrorSub}>{errorMessage}</span>
            <details>
              <summary>{m.errorDetails.defaultMessage}</summary>
              <pre>{errorDetails}</pre>
              {Boolean(supportLink) && (
                <a
                  href={supportLink}
                  target="_blank"
                  className={classnames(
                    buttonStyle.button,
                    buttonStyle.secondary,
                    style.supportButton,
                  )}
                >
                  {sharedMessages.getSupport.defaultMessage}
                </a>
              )}
            </details>
          </Col>
        </Row>
      </Container>
    </div>
  )
}

FullViewErrorInner.propTypes = FullViewErrorRaw.propTypes = {
  error: PropTypes.error.isRequired,
}

FullViewError.propTypes = {
  error: PropTypes.error.isRequired,
  header: PropTypes.node,
  onlineStatus: PropTypes.onlineStatus.isRequired,
}

FullViewError.defaultProps = {
  header: undefined,
}

export { FullViewError, FullViewErrorInner, FullViewErrorRaw }
