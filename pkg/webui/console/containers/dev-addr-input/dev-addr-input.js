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

import Input from '../../../components/input'

import PropTypes from '../../../lib/prop-types'

const m = defineMessages({
  generate: 'Generate Device Address',
})

const DevAddrInput = function(props) {
  const {
    className,
    name,
    onFocus,
    onChange,
    onBlur,
    value,
    fetching,
    disabled,
    autoFocus,
    error,
    warning,
    onDevAddrGenerate,
    generatedDevAddr,
  } = props

  React.useEffect(
    function() {
      if (Boolean(generatedDevAddr)) {
        onChange(generatedDevAddr)
        onBlur({ target: { value: generatedDevAddr } })
      }
    },
    [generatedDevAddr, onChange, onBlur],
  )

  const action = React.useMemo(
    function() {
      return {
        icon: 'autorenew',
        title: m.generate,
        type: 'button',
        disabled: fetching || disabled,
        onClick: onDevAddrGenerate,
        raw: true,
      }
    },
    [disabled, fetching, onDevAddrGenerate],
  )

  return (
    <Input
      type="byte"
      min={4}
      max={4}
      action={action}
      className={className}
      name={name}
      onChange={onChange}
      onBlur={onBlur}
      onFocus={onFocus}
      value={value}
      defaultValue={generatedDevAddr}
      error={error}
      warning={warning}
      loading={fetching}
      disabled={disabled}
      autoFocus={autoFocus}
    />
  )
}

DevAddrInput.propTypes = {
  className: PropTypes.string,
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func.isRequired,
  onBlur: PropTypes.func.isRequired,
  onFocus: PropTypes.func,
  fetching: PropTypes.bool,
  error: PropTypes.bool,
  warning: PropTypes.bool,
  onDevAddrGenerate: PropTypes.func.isRequired,
  generatedDevAddr: PropTypes.string,
  value: PropTypes.string,
  disabled: PropTypes.bool,
  autoFocus: PropTypes.bool,
}

DevAddrInput.defaultProps = {
  onFocus: () => null,
  fetching: false,
  disabled: false,
  error: false,
  warning: false,
  autoFocus: false,
}

export default DevAddrInput
