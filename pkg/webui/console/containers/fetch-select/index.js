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

import React, { useCallback, useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'

import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'

import PropTypes from '@ttn-lw/lib/prop-types'

const formatOptions = options =>
  Object.keys(options).map(key => ({ value: key, label: options[key] }))

const { component, ...fieldPropTypes } = Field.propTypes

export default ({
  optionsSelector,
  errorSelector,
  fetchingSelector,
  fetchOptions,
  defaultWarning,
  defaultTitle,
  optionsFormatter = formatOptions,
  defaultDescription,
  additionalOptions = [],
}) => {
  const FetchSelect = props => {
    const { warning, onChange, ...rest } = props
    const options = [...optionsFormatter(useSelector(optionsSelector)), ...additionalOptions]
    const error = useSelector(errorSelector)
    const fetching = useSelector(fetchingSelector)
    const dispatch = useDispatch()

    useEffect(() => {
      dispatch(fetchOptions())
    }, [dispatch])

    const handleChange = useCallback(
      value => {
        const selectedOption = options.find(option => option.value === value)
        onChange(selectedOption)
      },
      [onChange, options],
    )

    return (
      <Field
        {...rest}
        options={options}
        component={Select}
        isLoading={fetching}
        warning={Boolean(error) ? defaultWarning : warning}
        onChange={handleChange}
      />
    )
  }

  FetchSelect.propTypes = {
    ...fieldPropTypes,
    ...Select.propTypes,
    defaultWarning: PropTypes.message,
    description: PropTypes.message,
    fetchOptions: PropTypes.func.isRequired,
    menuPlacement: PropTypes.oneOf(['top', 'bottom', 'auto']),
    onChange: PropTypes.func,
    options: PropTypes.arrayOf(
      PropTypes.shape({ value: PropTypes.string, label: PropTypes.message }),
    ),
    title: PropTypes.message,
    warning: PropTypes.message,
  }

  FetchSelect.defaultProps = {
    description: defaultDescription,
    menuPlacement: 'auto',
    onChange: () => null,
    options: [],
    title: defaultTitle,
    warning: undefined,
    defaultWarning,
  }

  return FetchSelect
}
