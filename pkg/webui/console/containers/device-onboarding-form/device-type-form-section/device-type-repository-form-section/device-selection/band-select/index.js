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

import { selectDeviceModelFirmwareVersions } from '@console/store/selectors/device-repository'

import { SELECT_OTHER_OPTION } from '../../../../utils'
import messages from '../../../../messages'

const m = defineMessages({
  title: 'Profile (Region)',
})

const formatOptions = (profiles = []) =>
  profiles
    .map(profile => ({
      value: profile,
      label: profile,
    }))
    .concat([{ value: SELECT_OTHER_OPTION, label: messages.otherOption }])

const BandSelect = props => {
  const { name, onChange, brandId, modelId, fwVersion, ...rest } = props
  const { setFieldValue } = useFormContext()
  const versions = useSelector(state => selectDeviceModelFirmwareVersions(state, brandId, modelId))
  const version = versions.find(v => v.version === fwVersion) || { profiles: [] }
  const profiles = Object.keys(version.profiles)

  const options = React.useMemo(() => formatOptions(profiles), [profiles])
  const onlyOption = options.length > 0 && options.length <= 2 ? options[0].value : undefined

  React.useEffect(() => {
    if (onlyOption) {
      setFieldValue('version_ids.band_id', onlyOption)
    }
  }, [onlyOption, setFieldValue])

  return (
    <Field {...rest} options={options} name={name} title={m.title} component={Select} autoFocus />
  )
}

BandSelect.propTypes = {
  brandId: PropTypes.string.isRequired,
  fwVersion: PropTypes.string.isRequired,
  modelId: PropTypes.string.isRequired,
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
}

BandSelect.defaultProps = {
  onChange: () => null,
}

export default BandSelect
