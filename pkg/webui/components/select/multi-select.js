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

import React, { useCallback, useState } from 'react'
import { components } from 'react-select'
import AsyncSelect from 'react-select/async'
import { defineMessage, useIntl } from 'react-intl'
import classnames from 'classnames'
import { debounce } from 'lodash'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import Icon from '../icon'
import Button from '../button'

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

const m = defineMessage({
  remove: 'Remove',
})

const SuggestedMultiSelect = props => {
  const {
    value,
    name,
    onBlur,
    onChange,
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
    selectedValue => {
      if (!Boolean(value)) {
        setInputValue(selectedValue)
      }

      onChange(selectedValue)
    },
    [setInputValue, value, onChange],
  )

  const handleRemoveSelected = useCallback(
    (e, option) => {
      const newValue = inputValue.filter(o => o.value !== option.value)

      setInputValue(newValue)
    },
    [inputValue, setInputValue],
  )

  const debouncedFetch = debounce((query, callback) => {
    loadOptions(query).then(result => callback(result))
  }, 500)

  const cls = classnames(className, style.container, style[`input-width-${inputWidth}`], {
    [style.error]: error,
    [style.warning]: warning,
  })
  const selectedOptionsClasses = classnames(
    style.container,
    style.selectedOptionsContainer,
    style[`input-width-${inputWidth}`],
    'mt-cs-xs',
  )

  return (
    <>
      <AsyncSelect
        {...rest}
        isMulti
        controlShouldRenderValue={false}
        isClearable={false}
        isDisabled={disabled}
        loadOptions={debouncedFetch}
        className={cls}
        inputId={id}
        classNamePrefix="select"
        onChange={handleChange}
        onFocus={onFocus}
        value={inputValue}
        name={name}
        components={{ Input, Option: customOption, ...customComponents }}
        aria-describedby={rest['aria-describedby']}
        placeholder={Boolean(placeholder) ? formatMessage(placeholder) : undefined}
      />
      {inputValue?.map(option => (
        <div key={option.value} className={selectedOptionsClasses}>
          <Icon icon={option.icon} className="mr-cs-s" />
          <Message content={option.description ?? option.label} />
          <Button
            type="button"
            naked
            message={m.remove}
            value={option}
            onClick={handleRemoveSelected}
            className={style.removeOptionButton}
          />
        </div>
      ))}
    </>
  )
}

SuggestedMultiSelect.propTypes = {
  className: PropTypes.string,
  customComponents: PropTypes.shape({
    Option: PropTypes.func,
    SingleValue: PropTypes.func,
  }),
  disabled: PropTypes.bool,
  error: PropTypes.bool,
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

SuggestedMultiSelect.defaultProps = {
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
  loadOptions: () => null,
  showOptionIcon: false,
  customComponents: {},
}

export default SuggestedMultiSelect
