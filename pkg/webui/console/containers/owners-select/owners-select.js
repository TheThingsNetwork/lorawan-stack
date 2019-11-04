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

import Select from '../../../components/select'
import Field from '../../../components/form/field'

import { getOrganizationId, getUserId } from '../../../lib/selectors/id'
import PropTypes from '../../../lib/prop-types'

const m = defineMessages({
  title: 'Owner',
  warning: 'Cannot load user oganizations',
})

const OwnersSelect = props => {
  const {
    autoFocus,
    error,
    fetching,
    getOrganizationsList,
    menuPlacement,
    name,
    onChange,
    organizations,
    required,
    user,
  } = props

  React.useEffect(() => {
    getOrganizationsList()
  }, [getOrganizationsList])

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
  error: PropTypes.error,
  fetching: PropTypes.bool,
  getOrganizationsList: PropTypes.func.isRequired,
  menuPlacement: PropTypes.oneOf(['top', 'bottom', 'auto']),
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  organizations: PropTypes.arrayOf(PropTypes.organization).isRequired,
  required: PropTypes.bool,
  user: PropTypes.user.isRequired,
}

OwnersSelect.defaultProps = {
  autoFocus: false,
  error: undefined,
  fetching: false,
  onChange: () => null,
  menuPlacement: 'auto',
  required: false,
}

export default OwnersSelect
