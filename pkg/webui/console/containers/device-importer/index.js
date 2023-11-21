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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { isObject } from 'lodash'
import { useParams } from 'react-router-dom'

import tts from '@console/api/tts'

import { createFrontendError, isFrontend } from '@ttn-lw/lib/errors/utils'
import { getDeviceId } from '@ttn-lw/lib/selectors/id'
import { selectAsConfig, selectJsConfig, selectNsConfig } from '@ttn-lw/lib/selectors/env'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import randomByteString from '@console/lib/random-bytes'

import { convertTemplate } from '@console/store/actions/device-template-formats'

import { selectDeviceTemplate } from '@console/store/selectors/device-repository'

import Form from './form'
import Processor from './processor'
import m from './messages'

const conversionError = createFrontendError(m.conversionErrorTitle, m.conversionErrorMessage)

const DeviceImporter = () => {
  const { appId } = useParams()
  const deviceRepoTemplate = useSelector(selectDeviceTemplate)
  const asConfig = useSelector(selectAsConfig)
  const nsConfig = useSelector(selectNsConfig)
  const jsConfig = useSelector(selectJsConfig)
  const availableComponents = ['is']
  if (nsConfig.enabled) availableComponents.push('ns')
  if (jsConfig.enabled) availableComponents.push('js')
  if (asConfig.enabled) availableComponents.push('as')
  const editorRef = useRef()
  const [log, setLog] = useState('')
  const [currentDeviceIndex, setCurrentDeviceIndex] = useState(0)
  const [convertedDevices, setConvertedDevices] = useState([])
  const [deviceErrors, setDeviceErrors] = useState([])
  const [status, setStatus] = useState('initial')
  const prevStatus = useRef()
  const [step, setStep] = useState('initial')
  const [error, setError] = useState(undefined)
  const [aborted, setAborted] = useState(false)
  const createStream = useRef(null)
  const dispatch = useDispatch()

  useEffect(() => {
    // Disable undo manager of the code editor to release old logs from the heap.
    // Without this fix the browser can run out of memory when importing many end devices.
    prevStatus.current = status
    if (prevStatus !== 'initial' && status !== 'initial') {
      editorRef.current.editor.session.setUndoManager(null)
    }
  }, [status])

  const appendToLog = useCallback(
    message => {
      const text = typeof message !== 'string' ? JSON.stringify(message, null, 2) : message
      setLog(log => `${log}\n${text}`)
    },
    [setLog],
  )

  const handleCreationSuccess = useCallback(
    device => {
      appendToLog(device)
      setCurrentDeviceIndex(currentDeviceIndex => currentDeviceIndex + 1)
    },
    [appendToLog],
  )

  const logError = useCallback(
    error => {
      if (isObject(error)) {
        if (!isFrontend(error)) {
          const json = JSON.stringify(error, null, 2)
          setLog(log => `${log}\n${json}`)
        }
      }
    },
    [setLog],
  )

  const handleCreationError = useCallback(
    error => {
      logError(error)
      const currentDevice =
        convertedDevices.length > currentDeviceIndex ? convertedDevices[currentDeviceIndex] : {}
      const currentDeviceId =
        'end_device' in currentDevice
          ? getDeviceId(currentDevice.end_device)
          : `unknown device ID ${Date.now()}`
      setCurrentDeviceIndex(currentDeviceIndex => currentDeviceIndex + 1)
      setDeviceErrors(errors => [...errors, { deviceId: currentDeviceId, error }])
    },
    [convertedDevices, currentDeviceIndex, logError],
  )

  const handleFatalError = useCallback(
    error => {
      logError(error)

      const logAppend = '\n\nImport process cancelled due to error.'
      setStatus('error')
      setError(error)
      setLog(log => `${log}\n${logAppend}`)
    },
    [logError],
  )

  const handleAbort = useCallback(() => {
    if (createStream.current !== null) {
      createStream.current.abort()
      setAborted(true)
    }
  }, [createStream])

  const handleSubmit = useCallback(
    async values => {
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
        setStep('conversion')
        setStatus('processing')
        appendToLog('Converting end device templates…')
        const templateStream = await dispatch(attachPromise(convertTemplate(format_id, data)))

        devices = await new Promise((resolve, reject) => {
          const chunks = []

          templateStream.on('message', message => {
            appendToLog(message)
            chunks.push(message)
          })
          templateStream.on('error', reject)
          templateStream.on('close', () => resolve(chunks))

          templateStream.open()
        })

        if (devices.length === 0) {
          throw conversionError
        }

        setConvertedDevices(devices)
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

          // Fallback values
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
            field_mask.paths.push('version_ids')

            if (!device.lorawan_version && deviceRepoTemplate) {
              device.lorawan_version = deviceRepoTemplate.end_device.lorawan_version
              field_mask.paths.push('lorawan_version')
            }
            if (!device.lorawan_phy_version && deviceRepoTemplate) {
              device.lorawan_phy_version = deviceRepoTemplate.end_device.lorawan_phy_version
              field_mask.paths.push('lorawan_phy_version')
            }
            if (!device.supports_join && deviceRepoTemplate) {
              device.supports_join = deviceRepoTemplate.end_device.supports_join
              field_mask.paths.push('supports_join')
            }
            if (!device.mac_settings && deviceRepoTemplate) {
              device.mac_settings = deviceRepoTemplate.end_device.mac_settings
              field_mask.paths.push('mac_settings')
            }
            if (!device.frequency_plan_id && Boolean(frequency_plan_id)) {
              device.frequency_plan_id = frequency_plan_id
              field_mask.paths.push('frequency_plan_id')
            }
          }
        }
      } catch (error) {
        handleFatalError(error)
        return
      }

      // Start batch device creation.
      setStep('creation')
      appendToLog('Creating end devices…')

      try {
        createStream.current = tts.Applications.Devices.bulkCreate(appId, devices)

        await new Promise(resolve => {
          createStream.current.on('message', handleCreationSuccess)
          createStream.current.on('error', handleCreationError)
          createStream.current.on('close', resolve)

          createStream.current.start()
        })

        if (!aborted) {
          appendToLog('\nImport operation complete')
        } else {
          appendToLog('\nImport operation aborted')
        }
        setStatus('finished')
      } catch (error) {
        handleCreationError(error)
      }
    },
    [
      appId,
      asConfig,
      dispatch,
      handleCreationError,
      handleCreationSuccess,
      handleFatalError,
      jsConfig,
      nsConfig,
      setConvertedDevices,
      setStatus,
      setStep,
      aborted,
      appendToLog,
      deviceRepoTemplate,
      createStream,
    ],
  )

  const handleReset = useCallback(() => {
    setLog('')
    setCurrentDeviceIndex(0)
    setConvertedDevices([])
    setDeviceErrors([])
    setStatus('initial')
    setStep('initial')
    setError(undefined)
    setAborted(false)
  }, [])

  switch (step) {
    case 'conversion':
    case 'creation':
      return (
        <Processor
          log={log}
          currentDeviceIndex={currentDeviceIndex}
          deviceErrors={deviceErrors}
          status={status}
          step={step}
          error={error}
          convertedDevices={convertedDevices}
          aborted={aborted}
          handleAbort={handleAbort}
          handleReset={handleReset}
          editorRef={editorRef}
        />
      )
    case 'initial':
    default:
      return <Form handleSubmit={handleSubmit} jsEnabled={availableComponents.includes('js')} />
  }
}

export default DeviceImporter
