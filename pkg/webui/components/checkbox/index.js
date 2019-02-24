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
import PropTypes from 'prop-types'
import bind from 'autobind-decorator'
import classnames from 'classnames'

import style from './checkbox.styl'

@bind
class Checkbox extends React.PureComponent {

  onChange (evt) {
    const { onChange, name } = this.props

    onChange(name, evt.target.checked)
  }

  render () {
    const { value, onChange, indeterminate, ...rest } = this.props
    const checkedClassNames = classnames(style.checkbox, {
      [style.checkmark]: !indeterminate,
      [style.indeterminate]: indeterminate,
    })

    return (
      <label className={style.container}>
        <input
          className={style.input}
          type="checkbox"
          onChange={this.onChange}
          checked={value}
          {...rest}
        />
        <span className={checkedClassNames} />
      </label>
    )
  }
}


Checkbox.propTypes = {
  value: PropTypes.bool,
  onFocus: PropTypes.func,
  onBlur: PropTypes.func,
  onChange: PropTypes.func,
  disabled: PropTypes.bool,
  indeterminate: PropTypes.bool,
}

Checkbox.defaultProps = {
  onChange: () => null,
}

export default Checkbox
