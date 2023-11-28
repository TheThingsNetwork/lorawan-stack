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
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'
import { useFormContext } from '@ttn-lw/components/form'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { SELECT_OTHER_OPTION, SELECT_UNKNOWN_HW_OPTION } from '@console/lib/device-utils'

import { selectDeviceModelHardwareVersions } from '@console/store/selectors/device-repository'

const m = defineMessages({
  title: 'Hardware Ver.',
})

const formatOptions = (versions = []) =>
  versions
    .map(version => ({
      value: version.version,
      label: version.version,
    }))
    .concat([{ value: SELECT_OTHER_OPTION, label: sharedMessages.otherOption }])

const HardwareVersionSelect = props => {
  const { name, brandId, modelId, onChange, ...rest } = props
  const { setFieldValue, values } = useFormContext()
  const versions = useSelector(state => selectDeviceModelHardwareVersions(state, brandId, modelId))

  const options = React.useMemo(() => {
    const opts = formatOptions(versions)
    // When only the `Other...` option is available (so end device model has no hw versions defined
    // in the device repository) add another pseudo option that represents absence of hw versions.
    if (opts.length === 1) {
      opts.unshift({ value: SELECT_UNKNOWN_HW_OPTION, label: sharedMessages.unknownHwOption })
    }

    return opts
  }, [versions])

  React.useEffect(() => {
    if (options.length > 0 && options.length <= 2 && !values.version_ids.hardware_version.length) {
      setFieldValue('version_ids.hardware_version', options[0].value)
    }
  }, [options, setFieldValue, values])

  return (
    <Field {...rest} options={options} name={name} title={m.title} component={Select} autoFocus />
  )
}

HardwareVersionSelect.propTypes = {
  brandId: PropTypes.string.isRequired,
  modelId: PropTypes.string.isRequired,
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
}

HardwareVersionSelect.defaultProps = {
  onChange: () => null,
}

export default HardwareVersionSelect
