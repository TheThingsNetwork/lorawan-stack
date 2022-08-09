// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { useDispatch, useSelector } from 'react-redux'
import classnames from 'classnames'

import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'

import Message from '@ttn-lw/lib/components/message'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import env from '@ttn-lw/lib/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { getApplicationDevEUICount, issueDevEUI } from '@console/store/actions/applications'

import {
  selectApplicationDevEUICount,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'

import style from './dev-eui.styl'

const DevEUIComponent = props => {
  const { values, setFieldValue, initialValues, devEUISchema } = props
  const dispatch = useDispatch()
  const appId = useSelector(selectSelectedApplicationId)
  const promisifiedIssueDevEUI = attachPromise(issueDevEUI)
  const fetchDevEUICounter = attachPromise(getApplicationDevEUICount)
  const euiInputRef = React.useRef(null)
  const [devEUIGenerated, setDevEUIGenerated] = React.useState(false)
  const [errorMessage, setErrorMessage] = React.useState(undefined)
  const applicationDevEUICounter = useSelector(selectApplicationDevEUICount)

  const indicatorContent = Boolean(errorMessage)
    ? errorMessage
    : {
        ...sharedMessages.used,
        values: {
          currentValue: applicationDevEUICounter,
          maxValue: env.devEUIConfig.applicationLimit,
        },
      }

  const indicatorCls = classnames(style.indicator, {
    [style.error]:
      applicationDevEUICounter === env.devEUIConfig.applicationLimit || Boolean(errorMessage),
  })

  const handleDevEUIRequest = React.useCallback(async () => {
    const result = await dispatch(promisifiedIssueDevEUI(appId))
    await dispatch(fetchDevEUICounter(appId))
    return result.dev_eui
  }, [appId, dispatch, fetchDevEUICounter, promisifiedIssueDevEUI])

  const handleGenerate = React.useCallback(async () => {
    try {
      const result = await handleDevEUIRequest()
      setDevEUIGenerated(true)
      euiInputRef.current.focus()
      setErrorMessage(undefined)
      return result
    } catch (error) {
      if (error.details[0].name === 'global_eui_limit_reached') {
        setErrorMessage(sharedMessages.devEUIBlockLimitReached)
      } else setErrorMessage(sharedMessages.unknownError)
      setDevEUIGenerated(true)
    }
  }, [handleDevEUIRequest])

  const generateDeviceId = React.useCallback(
    (device = {}) => {
      const { ids: idsValues = {} } = device

      try {
        devEUISchema.validateSync(idsValues.dev_eui)
        return `eui-${idsValues.dev_eui.toLowerCase()}`
      } catch (e) {
        // We dont want to use invalid `dev_eui` as `device_id`.
      }

      return initialValues.ids.device_id || undefined
    },
    [initialValues.ids.device_id, devEUISchema],
  )

  const handleIdPrefill = React.useCallback(() => {
    if (values) {
      // Do not overwrite a value that the user has already set.
      if (values.ids.device_id === initialValues.ids.device_id) {
        const generatedId = generateDeviceId(values)
        setFieldValue('ids.device_id', generatedId)
      }
    }
  }, [values, setFieldValue, initialValues, generateDeviceId])

  const devEUIGenerateDisabled =
    applicationDevEUICounter === env.devEUIConfig.applicationLimit ||
    !env.devEUIConfig.devEUIIssuingEnabled ||
    devEUIGenerated

  return env.devEUIConfig.devEUIIssuingEnabled ? (
    <Form.Field
      title={sharedMessages.devEUI}
      name="ids.dev_eui"
      type="byte"
      min={8}
      max={8}
      required
      component={Input.Generate}
      tooltipId={tooltipIds.DEV_EUI}
      onBlur={handleIdPrefill}
      onGenerateValue={handleGenerate}
      actionDisable={devEUIGenerateDisabled}
      inputRef={euiInputRef}
    >
      <Message className={indicatorCls} component="label" content={indicatorContent} />
    </Form.Field>
  ) : (
    <Form.Field
      title={sharedMessages.devEUI}
      name="ids.dev_eui"
      type="byte"
      min={8}
      max={8}
      required
      component={Input}
      tooltipId={tooltipIds.DEV_EUI}
      onBlur={handleIdPrefill}
    />
  )
}

DevEUIComponent.propTypes = {
  devEUISchema: PropTypes.shape({
    validateSync: PropTypes.func,
  }),
  initialValues: PropTypes.shape({
    ids: PropTypes.shape({
      device_id: PropTypes.string,
    }),
  }),
  setFieldValue: PropTypes.func,
  values: PropTypes.shape({
    ids: PropTypes.shape({
      device_id: PropTypes.string,
    }),
  }),
}

DevEUIComponent.defaultProps = {
  devEUISchema: {
    validateSync: () => null,
  },
  setFieldValue: () => null,
  values: {
    ids: {
      device_id: undefined,
    },
  },
  initialValues: {
    ids: {
      device_id: undefined,
    },
  },
}

export default DevEUIComponent
