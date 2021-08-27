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
import clipboard from 'clipboard'
import { Helmet } from 'react-helmet'
import classnames from 'classnames'

import Footer from '@ttn-lw/components/footer'
import buttonStyle from '@ttn-lw/components/button/button.styl'
import Icon from '@ttn-lw/components/icon'

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

const appRoot = selectApplicationRootPath()
const siteName = selectApplicationSiteName()
const siteTitle = selectApplicationSiteTitle()
const supportLink = selectSupportLinkConfig()

// Mind any rendering that is dependant on context, since the errors
// can be rendered before such context is injected. Use the `safe`
// prop to conditionally render any context-dependant nodes.
const FullViewError = ({ error, header, onlineStatus, safe }) => (
  <div className={style.wrapper}>
    {Boolean(header) && header}
    <div className={style.flexWrapper}>
      <FullViewErrorInner error={error} safe={safe} />
    </div>
    <Footer onlineStatus={onlineStatus} safe={safe} />
  </div>
)

const FullViewErrorInner = ({ error, safe }) => {
  const isUnknown = isUnknownError(error)
  const statusCode = httpStatusCode(error)
  const isNotFound = isNotFoundError(error)
  const isFrontend = isFrontendError(error)

  const [copied, setCopied] = useState(false)

  let errorTitle = errorMessages.unknownErrorTitle
  let errorMessage = errorMessages.contactAdministrator
  if (!isUnknown) {
    errorMessage = error
  } else if (isNotFound) {
    errorMessage = errorMessages.genericNotFound
  }
  if (statusCode) {
    errorTitle = statusCodeMessages[statusCode]
  }
  if (isFrontend) {
    errorMessage = error.errorMessage
    if (Boolean(error.errorTitle)) {
      errorTitle = error.errorTitle
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

  const errorDetails = JSON.stringify(error, undefined, 2)
  const hasErrorDetails = !isNotFound && Boolean(error) && errorDetails.length > 2

  return (
    <div className={style.fullViewError} data-test-id="full-error-view">
      <Container>
        <Row>
          <Col xl={6} lg={8} md={10} sm={12}>
            {safe ? (
              <Helmet titleTemplate={`%s - ${siteTitle ? `${siteTitle} - ` : ''}${siteName}`}>
                <title>Error</title>
              </Helmet>
            ) : (
              <IntlHelmet title={errorMessages.error} />
            )}
            <h1>
              <Icon className={style.icon} textPaddedRight icon="error_outline" />
              <Message content={errorTitle} />
            </h1>
            <div className={style.fullViewErrorSub}>
              <ErrorMessage component="span" content={errorMessage} />
              {!isNotFound && (
                <>
                  <br />
                  <Message component="span" content={errorMessages.inconvenience} />
                </>
              )}
            </div>
            {Boolean(supportLink && !isNotFound) && (
              <div className={style.errorActions}>
                <a
                  href={supportLink}
                  target="_blank"
                  className={classnames(buttonStyle.button, style.supportButton)}
                >
                  <Message content={sharedMessages.getSupport} />
                </a>
                {hasErrorDetails && <Message content={errorMessages.attachToSupportInquiries} />}
              </div>
            )}
            {isNotFound && (
              <a
                icon="keyboard_arrow_left"
                message={sharedMessages.backToOverview}
                href={appRoot}
                className={buttonStyle.button}
              />
            )}
            {hasErrorDetails && (
              <>
                <hr />
                <details>
                  <summary>
                    <Message content={errorMessages.additionalInformation} />
                  </summary>
                  <pre>{errorDetails}</pre>
                  <button
                    onClick={handleCopyClick}
                    className={classnames(
                      buttonStyle.button,
                      buttonStyle.secondary,
                      style.supportButton,
                    )}
                    data-clipboard-text={errorDetails}
                    ref={copyButton}
                  >
                    <Icon icon={copied ? 'done' : 'file_copy'} textPaddedRight />
                    <Message
                      content={
                        copied ? sharedMessages.copiedToClipboard : sharedMessages.copyToClipboard
                      }
                    />
                  </button>
                </details>
              </>
            )}
          </Col>
        </Row>
      </Container>
    </div>
  )
}

FullViewErrorInner.propTypes = {
  error: PropTypes.error.isRequired,
  safe: PropTypes.bool,
}

FullViewErrorInner.defaultProps = {
  safe: false,
}

FullViewError.propTypes = {
  error: PropTypes.error.isRequired,
  header: PropTypes.node,
  onlineStatus: PropTypes.onlineStatus,
  safe: PropTypes.bool,
}

FullViewError.defaultProps = {
  header: undefined,
  onlineStatus: undefined,
  safe: false,
}

export { FullViewError, FullViewErrorInner }
