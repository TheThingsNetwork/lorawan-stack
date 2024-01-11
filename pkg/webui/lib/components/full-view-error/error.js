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

import Footer from '@ttn-lw/components/footer'
import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import errorMessages from '@ttn-lw/lib/errors/error-messages'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import * as cache from '@ttn-lw/lib/cache'
import {
  isUnknown as isUnknownError,
  isNotFoundError,
  isFrontend as isFrontendError,
  isBackend as isBackendError,
  getCorrelationId,
  getBackendErrorId,
  isOAuthClientRefusedError,
  isOAuthInvalidStateError,
} from '@ttn-lw/lib/errors/utils'
import statusCodeMessages from '@ttn-lw/lib/errors/status-code-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import {
  selectApplicationRootPath,
  selectSupportLinkConfig,
  selectApplicationSiteName,
  selectApplicationSiteTitle,
  selectDocumentationUrlConfig,
} from '@ttn-lw/lib/selectors/env'

import style from './error.styl'

const appRoot = selectApplicationRootPath()
const siteName = selectApplicationSiteName()
const siteTitle = selectApplicationSiteTitle()
const supportLink = selectSupportLinkConfig()
const documentationLink = selectDocumentationUrlConfig()
const hasSupportLink = Boolean(supportLink)
const lastRefreshAttemptCacheKey = 'last-refresh-attempt'

// Mind any rendering that is dependant on context, since the errors
// can be rendered before such context is injected. Use the `safe`
// prop to conditionally render any context-dependant nodes.
const FullViewError = ({ error, header, onlineStatus, safe, action, unexpected }) => (
  <div className={style.wrapper}>
    {Boolean(header) && header}
    <div className={style.flexWrapper}>
      <FullViewErrorInner error={error} safe={safe} action={action} unexpected={unexpected} />
    </div>
    <Footer
      onlineStatus={onlineStatus}
      documentationLink={documentationLink}
      supportLink={supportLink}
      safe={safe}
    />
  </div>
)

const FullViewErrorInner = ({ error, safe, action, unexpected }) => {
  const isUnknown = isUnknownError(error)
  const isNotFound = isNotFoundError(error)
  const isFrontend = isFrontendError(error)
  const isBackend = isBackendError(error)
  const hasAction = Boolean(action)
  const isErrorObject = error instanceof Error
  const isOAuthCallback = /oauth.*\/callback$/.test(window.location.pathname)

  const errorId = getBackendErrorId(error) || 'n/a'
  const correlationId = getCorrelationId(error) || 'n/a'

  const [copied, setCopied] = useState(false)

  const lastRefreshAttempt = cache.get(lastRefreshAttemptCacheKey)
  let doBlankRender = false

  let errorMessage
  let errorTitle
  if (isNotFound) {
    errorTitle = statusCodeMessages['404']
    errorMessage = errorMessages.genericNotFound
  } else if (isOAuthCallback) {
    errorTitle = sharedMessages.loginFailed
    errorMessage = errorMessages.loginFailedDescription
    if (isOAuthClientRefusedError(error) || error.error === 'access_denied') {
      errorMessage = errorMessages.loginFailedAbortDescription
    } else if (isOAuthInvalidStateError(error)) {
      // Usually in case of state errors, the state has expired or otherwise
      // invalidated since initiating the OAuth request. This can usually be
      // resolved by re-initiating the OAuth flow. We need to keep track of
      // refresh attempts though to avoid infinite refresh loops.
      if (!lastRefreshAttempt || Date.now() - lastRefreshAttempt > 6 * 1000) {
        cache.set(lastRefreshAttemptCacheKey, Date.now())
        doBlankRender = true
        window.location = appRoot
      }
    }
  } else if (isFrontend) {
    errorMessage = error.errorMessage
    if (Boolean(error.errorTitle)) {
      errorTitle = error.errorTitle
    }
  } else if (!isUnknown) {
    errorTitle = errorMessages.error
    errorMessage = errorMessages.errorOccurred
  } else {
    errorTitle = errorMessages.error
    errorMessage = errorMessages.genericError
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
  const hasErrorDetails =
    (!isNotFound && Boolean(error) && errorDetails.length > 2) || (isFrontend && error.errorCode)

  // Do not render anything when performing a redirect.
  if (doBlankRender) {
    return null
  }

  return (
    <div className={style.fullViewError} data-test-id="full-error-view">
      <Container>
        <Row>
          <Col xl={7} lg={8} md={10} sm={12}>
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
              <Message component="span" content={errorMessage} />
              {!isNotFound && unexpected && (
                <>
                  {' '}
                  <Message
                    component="span"
                    content={
                      hasSupportLink
                        ? errorMessages.contactSupport
                        : errorMessages.contactAdministrator
                    }
                  />
                  <br />
                  <Message component="span" content={errorMessages.inconvenience} />
                </>
              )}
            </div>
            <div className={style.errorActions}>
              {isNotFound && (
                <Button.AnchorLink
                  href={appRoot}
                  className={style.actionButton}
                  icon="keyboard_arrow_left"
                  message={sharedMessages.backToOverview}
                  secondary
                />
              )}
              {isOAuthCallback && (
                <Button.AnchorLink
                  href={appRoot}
                  className={style.actionButton}
                  icon="keyboard_arrow_left"
                  message={sharedMessages.backToLogin}
                  secondary
                />
              )}
              {hasAction && (
                <Button
                  type="button"
                  primary
                  className={style.actionButton}
                  onClick={action.action}
                  icon={action.icon}
                  message={action.message}
                  secondary
                />
              )}
              {hasSupportLink && !isNotFound && (
                <>
                  <Button.AnchorLink
                    href={supportLink}
                    target="_blank"
                    className={style.actionButton}
                    icon="contact_support"
                    message={sharedMessages.getSupport}
                    secondary
                  />
                  {hasErrorDetails && (
                    <Message component="span" content={errorMessages.attachToSupportInquiries} />
                  )}
                </>
              )}
            </div>
            {hasErrorDetails && (
              <>
                {isErrorObject && (
                  <>
                    <hr />
                    <div className={style.detailColophon}>
                      <span>
                        Error Type: <code>{error.name}</code>
                      </span>
                    </div>
                  </>
                )}
                {isFrontend && (
                  <>
                    <hr />
                    <div className={style.detailColophon}>
                      <span>
                        Frontend Error ID: <code>{error.errorCode}</code>
                      </span>
                    </div>
                  </>
                )}
                {isBackend && (
                  <>
                    <hr />
                    <div className={style.detailColophon}>
                      <span>
                        Error ID: <code>{errorId}</code>
                      </span>
                      <span>
                        Correlation ID: <code>{correlationId}</code>
                      </span>
                    </div>
                  </>
                )}
                <hr />
                <details>
                  <summary>
                    <Message content={errorMessages.technicalDetails} />
                  </summary>
                  <pre>{errorDetails}</pre>
                  <Button
                    onClick={handleCopyClick}
                    className={style.actionButton}
                    data-clipboard-text={errorDetails}
                    ref={copyButton}
                    icon={copied ? 'done' : 'file_copy'}
                    message={
                      copied ? sharedMessages.copiedToClipboard : sharedMessages.copyToClipboard
                    }
                    secondary
                  />
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
  action: PropTypes.shape({
    action: PropTypes.func.isRequired,
    icon: PropTypes.string.isRequired,
    message: PropTypes.message.isRequired,
  }),
  error: PropTypes.error.isRequired,
  safe: PropTypes.bool,
  unexpected: PropTypes.bool,
}

FullViewErrorInner.defaultProps = {
  action: undefined,
  safe: false,
  unexpected: true,
}

FullViewError.propTypes = {
  action: PropTypes.shape({
    action: PropTypes.func.isRequired,
    icon: PropTypes.string.isRequired,
    message: PropTypes.message.isRequired,
  }),
  error: PropTypes.error.isRequired,
  header: PropTypes.node,
  onlineStatus: PropTypes.onlineStatus,
  safe: PropTypes.bool,
  unexpected: PropTypes.bool,
}

FullViewError.defaultProps = {
  action: undefined,
  header: undefined,
  onlineStatus: undefined,
  safe: false,
  unexpected: true,
}

export { FullViewError, FullViewErrorInner }
