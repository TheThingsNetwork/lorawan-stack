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

/* eslint-disable react/prop-types, import/prefer-default-export */

import React, { useCallback, useState } from 'react'

import Input from '..'

const Example = props => {
  const { value: initialValue, type, component: Component, ...rest } = props

  const [value, setValue] = useState(initialValue || '')

  const handleChangeInput = useCallback(newValue => {
    setValue(newValue)
  }, [])

  const InputComponent = Component || Input

  return <InputComponent {...rest} type={type} onChange={handleChangeInput} value={value} />
}

export { Example }
