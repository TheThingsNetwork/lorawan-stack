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
import { defineMessages, useIntl } from 'react-intl'

import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'

import PropTypes from '@ttn-lw/lib/prop-types'

import { SELECT_OTHER_OPTION } from '../../../utils'
import messages from '../../../messages'

const m = defineMessages({
  title: 'Brand',
  warning: 'End device models unavailable',
  noOptionsMessage: 'No matching brand found',
})

const formatOptions = (brands = []) =>
  brands
    .map(brand => ({
      value: brand.brand_id,
      label: brand.name || brand.brand_id,
    }))
    .concat([{ value: SELECT_OTHER_OPTION, label: messages.otherOption }])

const BrandSelect = props => {
  const { appId, name, error, fetching, brands, listBrands, onChange, ...rest } = props
  const { formatMessage } = useIntl()

  React.useEffect(() => {
    listBrands(appId, {}, ['name'])
  }, [appId, listBrands])

  const options = React.useMemo(() => formatOptions(brands), [brands])
  const handleChange = React.useCallback(
    value => {
      onChange(options.find(e => e.value === value))
    },
    [onChange, options],
  )
  const handleNoOptions = React.useCallback(() => formatMessage(m.noOptionsMessage), [
    formatMessage,
  ])

  return (
    <Field
      {...rest}
      options={options}
      name={name}
      title={m.title}
      component={Select}
      isLoading={fetching}
      warning={Boolean(error) ? m.warning : undefined}
      onChange={handleChange}
      noOptionsMessage={handleNoOptions}
      placeholder={messages.typeToSearch}
    />
  )
}

BrandSelect.propTypes = {
  appId: PropTypes.string.isRequired,
  brands: PropTypes.arrayOf(
    PropTypes.shape({
      brand_id: PropTypes.string.isRequired,
    }),
  ),
  error: PropTypes.error,
  fetching: PropTypes.bool,
  listBrands: PropTypes.func.isRequired,
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func,
}

BrandSelect.defaultProps = {
  error: undefined,
  fetching: false,
  brands: [],
  onChange: () => null,
}

export default BrandSelect
