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
import ReactSelect from 'react-select'

import style from './select.styl'

// Map value to a plain string, instead of value object.
// See: https://github.com/JedWatson/react-select/issues/2841
const getValue = (opts, val) => opts.find(o => o.value === val)

const Select = function ({ value, className, ...rest }) {
  const classNames = className ? [ className, style.container ].join(' ') : style.container
  return (
    <ReactSelect
      className={classNames}
      classNamePrefix="select"
      value={getValue(rest.options, value)}
      {...rest}
    />
  )
}

Select.propTypes = {}

export default Select
