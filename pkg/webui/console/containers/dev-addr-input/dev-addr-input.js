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

import Input from '@ttn-lw/components/input'

import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  generate: 'Generate end device address',
})

const DevAddrInput = props => {
  const {
    className,
    id,
    name,
    onFocus,
    onChange,
    onBlur,
    value,
    disabled,
    autoFocus,
    warning,
    error: fieldError,
    loading: fieldLoading,
    onGenerate,
    generatedError,
    generatedLoading,
    generatedValue,
    ...rest
  } = props

  const action = {
    icon: 'autorenew',
    title: m.generate,
    type: 'button',
    disabled: fieldLoading || disabled || generatedLoading,
    onClick: onGenerate,
    raw: true,
  }

  const showLoading = fieldLoading || generatedLoading
  const showError = fieldError
  // Always show field validation error first.
  const showWarning = fieldError ? false : Boolean(warning) && generatedError

  React.useEffect(() => {
    if (Boolean(generatedValue)) {
      onChange(generatedValue, true)
    }
  }, [generatedValue, onChange])

  return (
    <Input
      type="byte"
      id={id}
      min={4}
      max={4}
      action={action}
      className={className}
      name={name}
      onChange={onChange}
      onBlur={onBlur}
      onFocus={onFocus}
      value={value}
      error={showError}
      warning={showWarning}
      loading={showLoading}
      disabled={disabled}
      autoFocus={autoFocus}
      {...rest}
    />
  )
}

DevAddrInput.propTypes = {
  autoFocus: PropTypes.bool,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  error: PropTypes.bool,
  generatedError: PropTypes.bool.isRequired,
  generatedLoading: PropTypes.bool.isRequired,
  generatedValue: PropTypes.string.isRequired,
  id: PropTypes.string.isRequired,
  loading: PropTypes.bool,
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func.isRequired,
  onChange: PropTypes.func.isRequired,
  onFocus: PropTypes.func,
  onGenerate: PropTypes.func.isRequired,
  value: PropTypes.string,
  warning: PropTypes.bool,
}

DevAddrInput.defaultProps = {
  className: undefined,
  onFocus: () => null,
  disabled: false,
  error: false,
  warning: false,
  autoFocus: false,
  value: undefined,
  loading: false,
}

export default DevAddrInput
