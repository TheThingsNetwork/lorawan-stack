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

import Icon, { IconCicleCheck, IconX } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './step.styl'

const Step = props => {
  const { className, title, description, status, stepNumber, active, transitionFailed, vertical } =
    props

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
    label = <Icon icon={IconCicleCheck} nudgeDown />
  } else if (isFailure) {
    label = <Icon icon={IconX} nudgeDown />
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
}

Step.defaultProps = {
  active: false,
  className: undefined,
  description: undefined,
  transitionFailed: false,
  vertical: false,
  status: 'wait',
  stepNumber: 1,
}

Step.propTypes = {
  active: PropTypes.bool,
  className: PropTypes.string,
  description: PropTypes.message,
  status: PropTypes.oneOf(['success', 'failure', 'current', 'wait']),
  stepNumber: PropTypes.number,
  title: PropTypes.message.isRequired,
  transitionFailed: PropTypes.bool,
  vertical: PropTypes.bool,
}

const MemoizedStep = React.memo(Step)
MemoizedStep.displayName = 'Stepper.Step'

export default MemoizedStep
