// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'
import { useFormContext } from '@ttn-lw/components/form'

import PropTypes from '@ttn-lw/lib/prop-types'

import { SELECT_OTHER_OPTION } from '../../../../utils'
import messages from '../../../../messages'

const m = defineMessages({
  title: 'Firmware Ver.',
})

const formatOptions = (versions = []) =>
  versions
    .map(version => ({
      value: version.version,
      label: version.version,
    }))
    .concat([{ value: SELECT_OTHER_OPTION, label: messages.otherOption }])

const FirmwareVersionSelect = props => {
  const { name, versions, onChange, ...rest } = props
  const { setFieldValue } = useFormContext()

  const options = React.useMemo(() => formatOptions(versions), [versions])

  React.useEffect(() => {
    if (options.length > 0 && options.length <= 2) {
      setFieldValue('version_ids.firmware_version', options[0].value)
    }
  }, [setFieldValue, options])

  return <Field {...rest} options={options} name={name} title={m.title} component={Select} />
}

FirmwareVersionSelect.propTypes = {
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  versions: PropTypes.arrayOf(
    PropTypes.shape({
      version: PropTypes.string.isRequired,
    }),
  ),
}

FirmwareVersionSelect.defaultProps = {
  versions: [],
  onChange: () => null,
}

export default FirmwareVersionSelect
