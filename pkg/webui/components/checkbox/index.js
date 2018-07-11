// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import PropTypes from 'prop-types'

import style from './checkbox.styl'

const Checkbox = function ({ value, onChange, type, error, warning, ...rest }) {
  return (
    <label className={style.container}>
      <input
        className={style.input}
        type="checkbox"
        onChange={onChange}
        {...rest}
      />
      <span className={style.checkmark} />
    </label>
  )
}

Checkbox.propTypes = {
  value: PropTypes.bool,
  onFocus: PropTypes.func,
  onBlur: PropTypes.func,
  onChange: PropTypes.func,
  disabled: PropTypes.bool,
  readOnly: PropTypes.bool,
  error: PropTypes.bool,
  warning: PropTypes.bool,
}

Checkbox.defaultProps = {
  onChange: () => null,
}

export default Checkbox
