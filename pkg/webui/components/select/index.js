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
import { injectIntl } from 'react-intl'
import bind from 'autobind-decorator'

import style from './select.styl'

// Map value to a plain string, instead of value object.
// See: https://github.com/JedWatson/react-select/issues/2841
const getValue = (opts, val) => opts.find(o => o.value === val)

@bind
class Select extends React.PureComponent {

  onChange ({ value }) {
    const { onChange } = this.props

    if (onChange) {
      onChange(value)
    }
  }

  render () {
    const {
      className,
      options,
      intl,
      value,
      onChange,
      ...rest
    } = this.props

    const formatMessage = (label, values) => intl ? intl.formatMessage(label, values) : label
    const classNames = className ? [ className, style.container ].join(' ') : style.container
    const translatedOptions = options.map(function (option) {
      const { label, labelValues = {}} = option
      if (typeof label === 'object' && label.id && label.defaultMessage) {
        return { ...option, label: formatMessage(label, labelValues) }
      }

      return option
    })

    return (
      <ReactSelect
        className={classNames}
        classNamePrefix="select"
        value={getValue(translatedOptions, value)}
        options={translatedOptions}
        onChange={this.onChange}
        {...rest}
      />
    )
  }
}

export default injectIntl(Select)
export { Select }
