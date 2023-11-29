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
import { useSelector } from 'react-redux'

import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { SELECT_OTHER_OPTION } from '@console/lib/device-utils'

import {
  selectDeviceBrands,
  selectDeviceBrandsError,
  selectDeviceBrandsFetching,
} from '@console/store/selectors/device-repository'

const m = defineMessages({
  title: 'End device brand',
  noOptionsMessage: 'No matching brand found',
})

const formatOptions = (brands = []) =>
  brands
    .map(brand => ({
      value: brand.brand_id,
      label: brand.name || brand.brand_id,
      profileID: brand.brand_id,
    }))
    .concat([{ value: SELECT_OTHER_OPTION, label: sharedMessages.otherOption }])

const BrandSelect = props => {
  const { name, onChange, ...rest } = props
  const { formatMessage } = useIntl()
  const brands = useSelector(selectDeviceBrands)
  const error = useSelector(selectDeviceBrandsError)
  const fetching = useSelector(selectDeviceBrandsFetching)

  const options = React.useMemo(() => formatOptions(brands), [brands])
  const handleNoOptions = React.useCallback(
    () => formatMessage(m.noOptionsMessage),
    [formatMessage],
  )

  return (
    <Field
      {...rest}
      options={options}
      name={name}
      title={m.title}
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

BrandSelect.propTypes = {
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
}

BrandSelect.defaultProps = {
  onChange: () => null,
}

export default BrandSelect
