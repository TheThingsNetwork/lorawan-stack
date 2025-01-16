// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector } from 'react-redux'

import CodeEditor from '@ttn-lw/components/code-editor'
import ProgressBar from '@ttn-lw/components/progress-bar'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Button from '@ttn-lw/components/button'
import ErrorNotification from '@ttn-lw/components/error-notification'
import Notification from '@ttn-lw/components/notification'
import Status from '@ttn-lw/components/status'
import ButtonGroup from '@ttn-lw/components/button/group'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'
import ErrorMessage from '@ttn-lw/lib/components/error-message'

import { isFrontend } from '@ttn-lw/lib/errors/utils'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectConsolePreferences } from '@console/store/selectors/user-preferences'

import m from './messages'

const statusMap = {
  processing: 'good',
  error: 'bad',
  finished: 'good',
}

const docLinkValue = msg => (
  <Link.DocLink secondary path="/devices/adding-devices/adding-devices-in-bulk">
    {msg}
  </Link.DocLink>
)

const Processor = ({
  log,
  currentDeviceIndex,
  deviceErrors,
  status,
  step,
  error,
  convertedDevices,
  aborted,
  handleAbort,
  handleReset,
  editorRef,
}) => {
  const consolePreferences = useSelector(selectConsolePreferences)
  const darkTheme = consolePreferences.console_theme === 'CONSOLE_THEME_DARK'
  const appId = useSelector(selectSelectedApplicationId)
  const hasErrored = status === 'error'
  const operationMessage = step === 'conversion' ? m.converting : m.creating
  let statusMessage
  if (!aborted) {
    if (status === 'error') {
      statusMessage = m.operationHalted
    } else if (status === 'finished') {
      statusMessage = m.operationFinished
    } else if (status === 'processing') {
      statusMessage = m.operationInProgress
    }
  } else {
    statusMessage = m.operationAborted
  }

  return (
    <div>
      {!hasErrored ? (
        <>
          <Status
            label={statusMessage}
            pulse={status === 'processing'}
            status={aborted ? 'mediocre' : statusMap[status] || 'unknown'}
          />
          <ProgressBar
            current={currentDeviceIndex}
            target={convertedDevices.length}
            showStatus
            showEstimation={!hasErrored && !aborted}
            className="mb-cs-m"
          >
            <Message
              content={m.progress}
              values={{
                errorCount: currentDeviceIndex - deviceErrors.length,
                deviceCount: convertedDevices.length,
              }}
            />
          </ProgressBar>
          {status === 'processing' && (
            <Message className="m-vert-0 ml-0 mr-cs-s" component="h4" content={operationMessage} />
          )}
        </>
      ) : (
        <ErrorNotification
          small
          content={error}
          title={!isFrontend(error) ? m.errorTitle : undefined}
          messageValues={
            !isFrontend(error)
              ? undefined
              : {
                  DocLink: docLinkValue,
                }
          }
        />
      )}
      {status === 'finished' && (
        <>
          {aborted && convertedDevices.length - currentDeviceIndex !== 0 && (
            <Notification
              small
              warning
              content={m.abortWarningMessage}
              title={m.abortWarningTitle}
              messageValues={{ count: convertedDevices.length - currentDeviceIndex }}
            />
          )}
          {deviceErrors.length !== 0 ? (
            <div>
              {deviceErrors.length >= currentDeviceIndex ? (
                <Notification
                  small
                  warning
                  content={m.noneWarningMessage}
                  title={m.noneWarningTitle}
                />
              ) : (
                <Notification
                  small
                  warning
                  content={m.incompleteWarningMessage}
                  title={m.incompleteWarningTitle}
                  messageValues={{ count: deviceErrors.length }}
                />
              )}
              <Message
                component="span"
                content={m.incompleteStatus}
                values={{ count: deviceErrors.length }}
              />
              <ul>
                {deviceErrors.map(({ deviceId, error }) => (
                  <li key={deviceId}>
                    <pre className="mb-cs-xs">{deviceId}</pre>
                    <ErrorMessage useTopmost content={error} />
                  </li>
                ))}
              </ul>
              <hr />
              <Message content={m.documentationHint} values={{ DocLink: docLinkValue }} />
            </div>
          ) : (
            <Notification small info title={m.successInfoTitle} content={m.successInfoMessage} />
          )}
        </>
      )}
      <Message content={m.processLog} component="h3" className="mb-cs-xxs" />
      <CodeEditor
        className="mb-cs-m"
        minLines={20}
        maxLines={20}
        mode="json"
        name="process_log"
        readOnly
        value={log}
        editorOptions={{ useWorker: false }}
        showGutter={false}
        scrollToBottom
        editorRef={editorRef}
        darkTheme={darkTheme}
      />
      <SubmitBar align="start">
        <ButtonGroup>
          {status === 'finished' && (
            <>
              {!hasErrored ? (
                <Button.Link
                  busy={status !== 'finished' && !hasErrored}
                  to={`/applications/${appId}/devices`}
                  message={m.proceed}
                  primary
                />
              ) : (
                <Button
                  busy={status !== 'finished' && !hasErrored}
                  message={m.retry}
                  onClick={handleReset}
                  primary
                />
              )}
            </>
          )}
          {status === 'processing' && step === 'creation' && (
            <Button danger message={m.abort} onClick={handleAbort} secondary />
          )}
        </ButtonGroup>
      </SubmitBar>
    </div>
  )
}

Processor.propTypes = {
  aborted: PropTypes.bool.isRequired,
  convertedDevices: PropTypes.arrayOf(
    PropTypes.shape({
      deviceId: PropTypes.string,
      device: PropTypes.shape({}),
    }),
  ).isRequired,
  currentDeviceIndex: PropTypes.number.isRequired,
  deviceErrors: PropTypes.arrayOf(
    PropTypes.shape({
      deviceId: PropTypes.string.isRequired,
      error: PropTypes.string.isRequired,
    }),
  ).isRequired,
  editorRef: PropTypes.shape({}).isRequired,
  error: PropTypes.error,
  handleAbort: PropTypes.func.isRequired,
  handleReset: PropTypes.func.isRequired,
  log: PropTypes.string.isRequired,
  status: PropTypes.string.isRequired,
  step: PropTypes.string.isRequired,
}

Processor.defaultProps = {
  error: undefined,
}

export default Processor
