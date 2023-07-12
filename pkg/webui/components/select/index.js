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

import React, { useCallback, useState } from 'react'
import ReactSelect, { components } from 'react-select'
import { useIntl } from 'react-intl'
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'

import Icon from '../icon'

import SuggestedSelect from './suggested-select'
import SuggestedMultiSelect from './multi-select'

import style from './select.styl'

const customOption = props => {
  const { showOptionIcon } = props.selectProps

  return (
    <components.Option {...props}>
      {showOptionIcon && <Icon icon={props.data.icon} className="mr-cs-xs" />}
      <b>{props.label}</b>
    </components.Option>
  )
}

customOption.propTypes = {
  data: PropTypes.shape({
    icon: PropTypes.string,
  }).isRequired,
  label: PropTypes.string.isRequired,
}

const Input = props => {
  const { selectProps } = props

  return <components.Input {...props} aria-describedby={selectProps['aria-describedby']} />
}

Input.propTypes = {
  selectProps: PropTypes.shape({
    'aria-describedby': PropTypes.string,
  }).isRequired,
}

// Map value to a plain string, instead of value object.
// See: https://github.com/JedWatson/react-select/issues/2841
const getValue = (opts, val) => opts.find(o => o.value === val)

const Select = props => {
  const {
    value,
    name,
    onBlur,
    onChange,
    hasAutosuggest,
    loadOptions,
    className,
    options,
    inputWidth,
    onFocus,
    disabled,
    error,
    warning,
    id,
    placeholder,
    showOptionIcon,
    customComponents,
    ...rest
  } = props

  const { formatMessage } = useIntl()
  const [inputValue, setInputValue] = useState(value)

  const handleChange = useCallback(
    value => {
      if (!('value' in props)) {
        setInputValue(value?.value)
      }

      onChange(value?.value, true)
    },
    [onChange, props],
  )

  const handleBlur = useCallback(
    event => {
      // https://github.com/JedWatson/react-select/issues/3523
      // Make sure the input name is always present in the event object.
      event.target.name = name

      if (typeof inputValue !== 'undefined') {
        // https://github.com/JedWatson/react-select/issues/3175
        event.target.value = inputValue
      }

      onBlur(event)
    },
    [onBlur, name, inputValue],
  )

  const cls = classnames(className, style.container, style[`input-width-${inputWidth}`], {
    [style.error]: error,
    [style.warning]: warning,
  })

  const translatedOptions = options?.map(option => {
    const { label, labelValues = {} } = option
    if (typeof label === 'object' && label.id && label.defaultMessage) {
      return { ...option, label: formatMessage(label, labelValues) }
    }

    return option
  })

  return (
    <ReactSelect
      className={cls}
      inputId={id}
      classNamePrefix="select"
      value={getValue(translatedOptions, value) || null}
      options={translatedOptions}
      onChange={handleChange}
      onBlur={handleBlur}
      onFocus={onFocus}
      isDisabled={disabled}
      name={name}
      components={{ Input }}
      aria-describedby={rest['aria-describedby']}
      placeholder={Boolean(placeholder) ? formatMessage(placeholder) : undefined}
      {...rest}
    />
  )
}

Select.propTypes = {
  className: PropTypes.string,
  customComponents: PropTypes.shape({
    Option: PropTypes.func,
    SingleValue: PropTypes.func,
  }),
  disabled: PropTypes.bool,
  error: PropTypes.bool,
  hasAutosuggest: PropTypes.bool,
  id: PropTypes.string,
  inputWidth: PropTypes.inputWidth,
  loadOptions: PropTypes.func,
  menuPlacement: PropTypes.string,
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func,
  onChange: PropTypes.func,
  onFocus: PropTypes.func,
  options: PropTypes.arrayOf(
    PropTypes.shape({
      value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
      label: PropTypes.message,
    }),
  ),
  placeholder: PropTypes.message,
  showOptionIcon: PropTypes.bool,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.shape({})]),
  warning: PropTypes.bool,
}

Select.defaultProps = {
  className: undefined,
  onChange: () => null,
  onBlur: () => null,
  onFocus: () => null,
  options: [],
  disabled: false,
  error: false,
  warning: false,
  value: undefined,
  id: undefined,
  inputWidth: 'm',
  placeholder: undefined,
  menuPlacement: 'auto',
  hasAutosuggest: false,
  loadOptions: () => null,
  showOptionIcon: false,
  customComponents: {},
}

Select.Suggested = SuggestedSelect
Select.Suggested.displayName = 'Select.Suggested'

Select.Multi = SuggestedMultiSelect
Select.Multi.displayName = 'Select.Multi'

export default Select
