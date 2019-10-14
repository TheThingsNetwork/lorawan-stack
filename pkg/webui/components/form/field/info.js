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
import classnames from 'classnames'

import PropTypes from '../../../lib/prop-types'
import Message from '../../../lib/components/message'
import from from '../../../lib/from'
import FormContext from '../context'

import style from './field.styl'

class InfoField extends React.Component {
  static contextType = FormContext

  static propTypes = {
    children: PropTypes.node.isRequired,
    className: PropTypes.string,
    title: PropTypes.message,
  }

  static defaultProps = {
    className: undefined,
    title: undefined,
  }

  render() {
    const { children, className, title, disabled: fieldDisabled } = this.props
    const { horizontal, disabled: formDisabled } = this.context
    const disabled = formDisabled || fieldDisabled
    const cls = classnames(className, style.field, from(style, { horizontal, disabled }))

    return (
      <div className={cls}>
        <label className={style.label}>
          <Message content={title} className={style.title} />
          <span className={style.reqicon}>&middot;</span>
        </label>
        <div className={classnames(style.componentArea, style.infoArea)}>{children}</div>
      </div>
    )
  }
}

export default InfoField
