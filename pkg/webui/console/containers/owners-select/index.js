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

import React from 'react'
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import Select from '@ttn-lw/components/select'
import Field from '@ttn-lw/components/form/field'

import { getOrganizationId, getUserId } from '@ttn-lw/lib/selectors/id'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectUser } from '@console/store/selectors/logout'
import {
  selectOrganizations,
  selectOrganizationsError,
  selectOrganizationsFetching,
} from '@console/store/selectors/organizations'

const m = defineMessages({
  title: 'Owner',
  warning: 'There was an error and the list of organizations could not be displayed',
})

const OwnersSelect = props => {
  const { autoFocus, menuPlacement, name, onChange, required } = props

  const user = useSelector(selectUser)
  const organizations = useSelector(selectOrganizations)
  const error = useSelector(selectOrganizationsError)
  const fetching = useSelector(selectOrganizationsFetching)

  const options = React.useMemo(() => {
    const usrOption = { label: getUserId(user), value: getUserId(user) }
    const orgsOptions = organizations.map(org => ({
      label: getOrganizationId(org),
      value: getOrganizationId(org),
    }))

    return [usrOption, ...orgsOptions]
  }, [user, organizations])
  const handleChange = React.useCallback(
    value => {
      onChange(options.find(option => option.value === value))
    },
    [onChange, options],
  )

  // Do not show the input when there are no alternative options.
  if (options.length === 1) {
    return null
  }

  return (
    <Field
      component={Select}
      options={options}
      name={name}
      required={required}
      title={m.title}
      autoFocus={autoFocus}
      isLoading={fetching}
      warning={Boolean(error) ? m.warning : undefined}
      menuPlacement={menuPlacement}
      onChange={handleChange}
      defaultValue={options[0]}
    />
  )
}

OwnersSelect.propTypes = {
  autoFocus: PropTypes.bool,
  menuPlacement: PropTypes.oneOf(['top', 'bottom', 'auto']),
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  required: PropTypes.bool,
}

OwnersSelect.defaultProps = {
  autoFocus: false,
  onChange: () => null,
  menuPlacement: 'auto',
  required: false,
}

export default OwnersSelect
