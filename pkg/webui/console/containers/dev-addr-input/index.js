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

import React, { useEffect } from 'react'
import { defineMessages } from 'react-intl'
import { merge } from 'lodash'

import Field from '@ttn-lw/components/form/field'
import { useFormContext } from '@ttn-lw/components/form'

import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import PropTypes from '@ttn-lw/lib/prop-types'

import DevAddrInput from './dev-addr-input'
import connect from './connect'

const m = defineMessages({
  devAddrFetchingFailure: 'There was an error and the end device address could not be generated',
})

const DevAddrField = props => {
  const {
    className,
    title,
    description,
    placeholder,
    name,
    disabled,
    required,
    autoFocus,
    onGenerate,
    generatedValue,
    generatedError,
    generatedLoading,
    encode,
    decode,
  } = props

  const { setValues } = useFormContext()

  useEffect(() => {
    setValues(values => merge({}, values, encode(generatedValue)))
  }, [encode, generatedValue, name, setValues])

  return (
    <Field
      className={className}
      title={title}
      description={description}
      placeholder={placeholder}
      name={name}
      disabled={disabled}
      required={required}
      autoFocus={autoFocus}
      warning={generatedError ? m.devAddrFetchingFailure : undefined}
      component={DevAddrInput}
      onGenerate={onGenerate}
      generatedError={generatedError}
      generatedLoading={generatedLoading}
      tooltipId={tooltipIds.DEVICE_ADDRESS}
      encode={encode}
      decode={decode}
    />
  )
}

DevAddrField.propTypes = {
  autoFocus: PropTypes.bool,
  className: PropTypes.string,
  decode: PropTypes.func,
  description: PropTypes.message,
  disabled: PropTypes.bool,
  encode: PropTypes.func,
  generatedError: PropTypes.bool,
  generatedLoading: PropTypes.bool,
  generatedValue: PropTypes.string,
  name: PropTypes.string.isRequired,
  onGenerate: PropTypes.func.isRequired,
  placeholder: PropTypes.message,
  required: PropTypes.bool,
  title: PropTypes.message.isRequired,
}

DevAddrField.defaultProps = {
  className: undefined,
  description: undefined,
  decode: val => val,
  encode: val => val,
  placeholder: undefined,
  disabled: false,
  required: false,
  autoFocus: false,
  generatedValue: '',
  generatedError: false,
  generatedLoading: false,
}

export default connect(DevAddrField)
