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

import PropTypes from '../../../lib/prop-types'

import Field from '../../../components/form/field'
import DevAddrInput from './dev-addr-input'
import connect from './connect'

const m = defineMessages({
  devAddrFetchingFailure: 'Could not generate device address',
})

const DevAddrField = function(props) {
  const {
    className,
    title,
    description,
    placeholder,
    name,
    fetching,
    disabled,
    required,
    autoFocus,
    horizontal,
    error,
    onDevAddrGenerate,
    generatedDevAddr,
  } = props

  return (
    <Field
      className={className}
      title={title}
      description={description}
      placeholder={placeholder}
      name={name}
      fetching={fetching}
      disabled={disabled}
      required={required}
      autoFocus={autoFocus}
      horizontal={horizontal}
      warning={Boolean(error) ? m.devAddrFetchingFailure : undefined}
      component={DevAddrInput}
      onDevAddrGenerate={onDevAddrGenerate}
      generatedDevAddr={generatedDevAddr}
    />
  )
}

DevAddrField.propTypes = {
  autoFocus: PropTypes.bool,
  className: PropTypes.string,
  description: PropTypes.message,
  disabled: PropTypes.bool,
  error: PropTypes.error,
  fetching: PropTypes.bool.isRequired,
  generatedDevAddr: PropTypes.string.isRequired,
  horizontal: PropTypes.bool,
  name: PropTypes.string.isRequired,
  onDevAddrGenerate: PropTypes.func.isRequired,
  placeholder: PropTypes.message,
  required: PropTypes.bool,
  title: PropTypes.message.isRequired,
}

DevAddrField.defaultProps = {
  className: undefined,
  description: undefined,
  error: undefined,
  placeholder: undefined,
  disabled: false,
  required: false,
  autoFocus: false,
  horizontal: false,
}

export default connect(DevAddrField)
