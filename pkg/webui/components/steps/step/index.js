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

import Icon from '../../icon'
import Message from '../../../lib/components/message'

import style from './step.styl'

const Step = React.memo(props => {
  const {
    className,
    title,
    description,
    status,
    stepNumber,
    active,
    transitionFailed,
    vertical,
  } = props

  const isSuccess = status === 'success'
  const isFailure = status === 'failure'

  const cls = classnames(className, style.step, {
    [style.current]: status === 'current',
    [style.wait]: status === 'wait',
    [style.success]: isSuccess,
    [style.failure]: isFailure,
    [style.active]: active,
    [style.transitionFailed]: transitionFailed,
    [style.vertical]: vertical,
  })
  const tailCls = classnames(style.tail, {
    [style.line]: vertical,
  })
  const titleCls = classnames(style.title, {
    [style.line]: !vertical,
  })

  let label
  if (isSuccess) {
    label = <Icon icon="done" nudgeDown />
  } else if (isFailure) {
    label = <Icon icon="close" nudgeDown />
  } else {
    label = <span className={style.label}>{stepNumber}</span>
  }

  return (
    <li className={cls}>
      <div className={tailCls} />
      <div className={style.status}>{label}</div>
      <div className={style.content}>
        <Message className={titleCls} content={title} component="span" />
        <Message className={style.description} content={description} component="p" />
      </div>
    </li>
  )
})

Step.displayName = 'Step'

Step.defaultProps = {
  active: false,
  className: undefined,
  description: undefined,
  status: 'wait',
  transitionFailed: false,
  vertical: PropTypes.bool,
}

Step.propTypes = {
  active: PropTypes.bool,
  className: PropTypes.string,
  description: PropTypes.message,
  status: PropTypes.oneOf(['success', 'failure', 'current', 'wait']),
  stepNumber: PropTypes.number.isRequired,
  title: PropTypes.message.isRequired,
  transitionFailed: PropTypes.bool,
  vertical: false,
}

export default Step
