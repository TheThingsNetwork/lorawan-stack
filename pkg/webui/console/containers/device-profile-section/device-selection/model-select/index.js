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
import { defineMessages, useIntl } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'

import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { SELECT_OTHER_OPTION } from '@console/lib/device-utils'

import { listModels } from '@console/store/actions/device-repository'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import {
  selectDeviceModelsByBrandId,
  selectDeviceModelsError,
  selectDeviceModelsFetching,
} from '@console/store/selectors/device-repository'

const m = defineMessages({
  noOptionsMessage: 'No matching model found',
})

const formatOptions = (models = []) =>
  models
    .map(model => ({
      value: model.model_id,
      label: model.name,
    }))
    .concat([{ value: SELECT_OTHER_OPTION, label: sharedMessages.otherOption }])

const ModelSelect = props => {
  const { brandId, name, onChange, ...rest } = props
  const { formatMessage } = useIntl()
  const dispatch = useDispatch()
  const appId = useSelector(selectSelectedApplicationId)
  const models = useSelector(state => selectDeviceModelsByBrandId(state, brandId))
  const error = useSelector(selectDeviceModelsError)
  const fetching = useSelector(selectDeviceModelsFetching)

  React.useEffect(() => {
    dispatch(
      listModels(appId, brandId, {}, [
        'name',
        'description',
        'firmware_versions',
        'hardware_versions',
        'key_provisioning',
        'photos',
        'product_url',
        'datasheet_url',
      ]),
    )
  }, [appId, brandId, dispatch])

  const options = React.useMemo(() => formatOptions(models), [models])
  const handleNoOptions = React.useCallback(
    () => formatMessage(m.noOptionsMessage),
    [formatMessage],
  )

  return (
    <Field
      {...rest}
      options={options}
      name={name}
      title={sharedMessages.model}
      component={Select}
      isLoading={fetching}
      warning={Boolean(error) ? sharedMessages.endDeviceModelsUnavailable : undefined}
      onChange={onChange}
      noOptionsMessage={handleNoOptions}
      placeholder={sharedMessages.typeToSearch}
      autoFocus
    />
  )
}

ModelSelect.propTypes = {
  appId: PropTypes.string.isRequired,
  brandId: PropTypes.string.isRequired,
  error: PropTypes.error,
  fetching: PropTypes.bool,
  models: PropTypes.arrayOf(
    PropTypes.shape({
      model_id: PropTypes.string.isRequired,
      name: PropTypes.string.isRequired,
    }),
  ),
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
}

ModelSelect.defaultProps = {
  error: undefined,
  fetching: false,
  models: [],
  onChange: () => null,
}

export default ModelSelect
