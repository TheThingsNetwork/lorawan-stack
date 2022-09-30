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

import React, { Component } from 'react'
import { connect } from 'react-redux'
import { push } from 'connected-react-router'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { Col, Row } from 'react-grid-system'
import { isObject } from 'lodash'

import tts from '@console/api/tts'

import CodeEditor from '@ttn-lw/components/code-editor'
import ProgressBar from '@ttn-lw/components/progress-bar'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Button from '@ttn-lw/components/button'
import ErrorNotification from '@ttn-lw/components/error-notification'
import Notification from '@ttn-lw/components/notification'
import Status from '@ttn-lw/components/status'
import Link from '@ttn-lw/components/link'
import ButtonGroup from '@ttn-lw/components/button/group'

import ErrorMessage from '@ttn-lw/lib/components/error-message'
import Message from '@ttn-lw/lib/components/message'

import DeviceImportForm from '@console/components/device-import-form'

import PropTypes from '@ttn-lw/lib/prop-types'
import { createFrontendError, isFrontend } from '@ttn-lw/lib/errors/utils'
import { selectNsConfig, selectJsConfig, selectAsConfig } from '@ttn-lw/lib/selectors/env'
import { getDeviceId } from '@ttn-lw/lib/selectors/id'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import randomByteString from '@console/lib/random-bytes'

import { convertTemplate } from '@console/store/actions/device-template-formats'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import style from './device-importer.styl'

const m = defineMessages({
  proceed: 'Proceed to end device list',
  retry: 'Retry from scratch',
  abort: 'Abort',
  converting: 'Converting templates…',
  creating: 'Creating end devices…',
  operationInProgress: 'Operation in progress',
  operationHalted: 'Operation halted',
  operationFinished: 'Operation finished',
  operationAborted: 'Operation aborted',
  errorTitle: 'There was an error and the operation could not be completed',
  conversionErrorTitle: 'Could not import devices',
  conversionErrorMessage:
    'An error occurred while processing the provided end device template. This could be due to invalid format, syntax or file encoding. Please check the provided template file and try again. See also our documentation on <DocLink>Importing End Devices</DocLink> for more information.',
  incompleteWarningTitle: 'Not all devices imported successfully',
  incompleteWarningMessage:
    '{count} {count, plural, one {end device} other {end devices}} could not be imported successfully, because {count, plural, one {its} other {their}} registration attempt resulted in an error',
  incompleteStatus:
    'The registration of the following {count, plural, one {end device} other {end devices}} failed:',
  noneWarningTitle: 'No end device was created',
  noneWarningMessage:
    'None of your specified end devices was imported, because each registration attempt resulted in an error',
  processLog: 'Process log',
  progress:
    'Successfully converted {errorCount} of {deviceCount} {deviceCount, plural, one {end device} other {end devices}}',
  successInfoTitle: 'All end devices imported successfully',
  successInfoMessage:
    'All of the specified end devices have been converted and imported successfully',
  documentationHint:
    'Please also see our documentation on <DocLink>Importing End Devices</DocLink> for more information and possible resolutions.',
  abortWarningTitle: 'Device import aborted',
  abortWarningMessage:
    'The end device import was aborted and the remaining {count} {count, plural, one {end device} other {end devices}} have not been imported',
  largeFileWarningMessage:
    'Providing files larger than {warningThreshold} can cause issues during the import process. We recommend you to split such files up into multiple smaller files and importing them one by one.',
})

const initialState = {
  log: '',
  currentDeviceIndex: 0,
  convertedDevices: [],
  deviceErrors: [],
  status: 'initial',
  step: 'inital',
  error: undefined,
  aborted: false,
}

const statusMap = {
  processing: 'good',
  error: 'bad',
  finished: 'good',
}

const conversionError = createFrontendError(m.conversionErrorTitle, m.conversionErrorMessage)
const docLinkValue = msg => (
  <Link.DocLink secondary path="/getting-started/migrating/import-devices/">
    {msg}
  </Link.DocLink>
)

@connect(
  state => {
    const asConfig = selectAsConfig()
    const nsConfig = selectNsConfig()
    const jsConfig = selectJsConfig()
    const availableComponents = ['is']
    if (nsConfig.enabled) availableComponents.push('ns')
    if (jsConfig.enabled) availableComponents.push('js')
    if (asConfig.enabled) availableComponents.push('as')

    return {
      appId: selectSelectedApplicationId(state),
      nsConfig,
      jsConfig,
      asConfig,
      availableComponents,
    }
  },
  dispatch => ({
    redirectToList: appId => dispatch(push(`/applications/${appId}/devices`)),
    convertTemplate: (format_id, data) => dispatch(attachPromise(convertTemplate(format_id, data))),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    redirectToList: () => dispatchProps.redirectToList(stateProps.appId),
  }),
)
export default class DeviceImporter extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    asConfig: PropTypes.stackComponent.isRequired,
    availableComponents: PropTypes.components.isRequired,
    convertTemplate: PropTypes.func.isRequired,
    jsConfig: PropTypes.stackComponent.isRequired,
    nsConfig: PropTypes.stackComponent.isRequired,
    redirectToList: PropTypes.func.isRequired,
  }

  constructor(props) {
    super(props)

    this.state = { ...initialState }
    this.editorRef = React.createRef()
    this.createStream = null
  }

  componentDidUpdate(prevProps, prevState) {
    const { status } = this.state

    if (prevState.status === 'initial' && status !== 'initial') {
      // Disable undo manager of the code editor to release old logs from the heap.
      // Without this fix the browser can run out of memory when importing many end devices.
      this.editorRef.current.editor.session.setUndoManager(null)
    }
  }

  @bind
  appendToLog(message) {
    const text = typeof message !== 'string' ? JSON.stringify(message, null, 2) : message
    this.setState(({ log }) => ({ log: `${log}\n${text}` }))
  }

  @bind
  handleCreationSuccess(device) {
    this.appendToLog(device)
    this.setState(({ currentDeviceIndex }) => ({
      currentDeviceIndex: currentDeviceIndex + 1,
    }))
  }

  @bind
  handleCreationError(error) {
    this.logError(error)
    const { convertedDevices, currentDeviceIndex } = this.state
    const currentDevice =
      convertedDevices.length > currentDeviceIndex ? convertedDevices[currentDeviceIndex] : {}
    const currentDeviceId =
      'end_device' in currentDevice
        ? getDeviceId(currentDevice.end_device)
        : `unknown device ID ${Date.now()}`
    this.setState(({ currentDeviceIndex, deviceErrors }) => ({
      currentDeviceIndex: currentDeviceIndex + 1,
      deviceErrors: [...deviceErrors, { deviceId: currentDeviceId, error }],
    }))
  }

  @bind
  logError(error) {
    if (isObject(error)) {
      if (!isFrontend(error)) {
        const json = JSON.stringify(error, null, 2)
        this.setState(({ log }) => ({ log: `${log}\n${json}` }))
      }
    }
  }

  @bind
  handleFatalError(error) {
    this.logError(error)

    const logAppend = '\n\nImport process cancelled due to error.'
    this.setState(({ log }) => ({ error, status: 'error', log: `${log}\n${logAppend}` }))
  }

  @bind
  async handleSubmit(values) {
    const { appId, jsConfig, nsConfig, asConfig, convertTemplate } = this.props
    const {
      format_id,
      data,
      set_claim_auth_code,
      frequency_plan_id,
      lorawan_version,
      lorawan_phy_version,
      version_ids,
      _inputMethod,
    } = values

    let devices = []

    try {
      // Start template conversion.
      this.setState({ step: 'conversion', status: 'processing' })
      this.appendToLog('Converting end device templates…')
      const templateStream = await convertTemplate(format_id, data)
      devices = await new Promise((resolve, reject) => {
        const chunks = []

        templateStream.on('chunk', message => {
          this.appendToLog(message)
          chunks.push(message)
        })
        templateStream.on('error', reject)
        templateStream.on('close', () => resolve(chunks))

        templateStream.open()
      })

      if (devices.length === 0) {
        throw conversionError
      }

      this.setState({ convertedDevices: devices })
      // Apply default values.
      for (const deviceAndFieldMask of devices) {
        const { end_device: device, field_mask } = deviceAndFieldMask

        if (set_claim_auth_code && jsConfig.enabled) {
          device.claim_authentication_code = { value: randomByteString(4 * 2) }
          field_mask.paths.push('claim_authentication_code')
        }
        if (device.supports_join && !device.join_server_address && jsConfig.enabled) {
          device.join_server_address = new URL(jsConfig.base_url).hostname
          field_mask.paths.push('join_server_address')
        }
        if (!device.application_server_address && asConfig.enabled) {
          device.application_server_address = new URL(asConfig.base_url).hostname
          field_mask.paths.push('application_server_address')
        }
        if (!device.network_server_address && nsConfig.enabled) {
          device.network_server_address = new URL(nsConfig.base_url).hostname
          field_mask.paths.push('network_server_address')
        }
        if (
          !device.frequency_plan_id &&
          Boolean(frequency_plan_id) &&
          nsConfig.enabled &&
          _inputMethod === 'manual'
        ) {
          device.frequency_plan_id = frequency_plan_id
          field_mask.paths.push('frequency_plan_id')
        }
        if (!device.lorawan_version && Boolean(lorawan_version) && _inputMethod === 'manual') {
          device.lorawan_version = lorawan_version
          field_mask.paths.push('lorawan_version')
        }
        if (
          !device.lorawan_phy_version &&
          Boolean(lorawan_phy_version) &&
          _inputMethod === 'manual'
        ) {
          device.lorawan_phy_version = lorawan_phy_version
          field_mask.paths.push('lorawan_phy_version')
        }
        if (!device.version_ids && Boolean(version_ids) && _inputMethod === 'device-repository') {
          device.version_ids = version_ids
          field_mask = `${field_mask},version_ids`
        }
      }
    } catch (error) {
      this.handleFatalError(error)
      return
    }

    // Start batch device creation.
    this.setState({
      step: 'creation',
    })
    this.appendToLog('Creating end devices…')

    try {
      this.createStream = tts.Applications.Devices.bulkCreate(appId, devices)

      await new Promise(resolve => {
        this.createStream.on('chunk', this.handleCreationSuccess)
        this.createStream.on('error', this.handleCreationError)
        this.createStream.on('close', resolve)

        this.createStream.start()
      })

      if (!this.state.aborted) {
        this.appendToLog('\nImport operation complete')
      } else {
        this.appendToLog('\nImport operation aborted')
      }
      this.setState({ status: 'finished' })
    } catch (error) {
      this.handleCreationError(error)
    }
  }

  @bind
  handleAbort() {
    if (this.createStream !== null) {
      this.createStream.abort()
      this.setState({ aborted: true })
    }
  }

  @bind
  handleReset() {
    this.setState(initialState)
  }

  get processor() {
    const {
      log,
      currentDeviceIndex,
      deviceErrors,
      status,
      step,
      error,
      convertedDevices,
      aborted,
    } = this.state
    const hasErrored = status === 'error'
    const { redirectToList } = this.props
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
              className={style.progressBar}
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
              <Message className={style.title} component="h4" content={operationMessage} />
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
                    <li key={deviceId} className={style.deviceErrorEntry}>
                      <pre>{deviceId}</pre>
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
        <Message content={m.processLog} component="h3" className={style.processLogTitle} />
        <CodeEditor
          className={style.logOutput}
          minLines={20}
          maxLines={20}
          mode="json"
          name="process_log"
          readOnly
          value={log}
          editorOptions={{ useWorker: false }}
          showGutter={false}
          scrollToBottom
          editorRef={this.editorRef}
        />
        <SubmitBar align="start">
          <ButtonGroup>
            <Button
              busy={status !== 'finished' && !hasErrored}
              message={hasErrored ? m.retry : m.proceed}
              onClick={hasErrored ? this.handleReset : redirectToList}
              primary
            />
            {status === 'processing' && step === 'creation' && (
              <Button danger message={m.abort} onClick={this.handleAbort} />
            )}
          </ButtonGroup>
        </SubmitBar>
      </div>
    )
  }

  get form() {
    const { availableComponents } = this.props
    const initialValues = {
      format_id: '',
      data: '',
      set_claim_auth_code: true,
      _inputMethod: 'manual',
    }
    const largeFile = 10 * 1024 * 1024
    return (
      <Row>
        <Col md={8}>
          <DeviceImportForm
            jsEnabled={availableComponents.includes('js')}
            largeFileWarningMessage={m.largeFileWarningMessage}
            warningSize={largeFile}
            initialValues={initialValues}
            onSubmit={this.handleSubmit}
          />
        </Col>
      </Row>
    )
  }

  render() {
    const { step } = this.state

    switch (step) {
      case 'conversion':
      case 'creation':
        return this.processor
      case 'initial':
      default:
        return this.form
    }
  }
}
