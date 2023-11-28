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
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'
import { useFormContext } from '@ttn-lw/components/form'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { isUnknownHwVersion, SELECT_OTHER_OPTION } from '@console/lib/device-utils'

import { selectDeviceModelFirmwareVersions } from '@console/store/selectors/device-repository'

const m = defineMessages({
  title: 'Firmware Ver.',
})

const formatOptions = (versions = []) =>
  versions
    .map(version => ({
      value: version.version,
      label: version.version,
    }))
    .concat([{ value: SELECT_OTHER_OPTION, label: sharedMessages.otherOption }])

const FirmwareVersionSelect = props => {
  const { name, onChange, brandId, modelId, hwVersion, ...rest } = props
  const { setFieldValue, values } = useFormContext()

  const versions = useSelector(state =>
    selectDeviceModelFirmwareVersions(state, brandId, modelId).filter(
      ({ supported_hardware_versions = [] }) =>
        (Boolean(hwVersion) && supported_hardware_versions.includes(hwVersion)) ||
        // Include firmware versions when there are no hardware versions configured in device repository
        // for selected end device model.
        isUnknownHwVersion(hwVersion),
    ),
  )

  const options = React.useMemo(() => formatOptions(versions), [versions])

  React.useEffect(() => {
    if (options.length > 0 && options.length <= 2 && !values.version_ids.firmware_version.length) {
      setFieldValue('version_ids.firmware_version', options[0].value)
    }
  }, [setFieldValue, options, values.version_ids.firmware_version.length])

  return (
    <Field {...rest} options={options} name={name} title={m.title} component={Select} autoFocus />
  )
}

FirmwareVersionSelect.propTypes = {
  brandId: PropTypes.string.isRequired,
  hwVersion: PropTypes.string.isRequired,
  modelId: PropTypes.string.isRequired,
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
}

FirmwareVersionSelect.defaultProps = {
  onChange: () => null,
}

export default FirmwareVersionSelect
