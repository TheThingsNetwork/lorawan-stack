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

import PropTypes from '../../lib/prop-types'

import style from './stepper.styl'

const Stepper = props => {
  const { className, children, stepCountStart, currentStep, status, vertical } = props

  const cls = classnames(className, style.stepper, {
    [style.vertical]: vertical,
  })
  const steps = React.Children.map(children, (child, index) => {
    if (!Boolean(child)) {
      return null
    }

    if (React.isValidElement(child) && child.type.displayName === 'Step') {
      const stepNumber = index + stepCountStart
      let stepStatus = status
      if (stepNumber < currentStep) {
        stepStatus = 'success'
      } else if (stepNumber > currentStep) {
        stepStatus = 'wait'
      }

      const props = {
        ...child.props,
        stepNumber,
        vertical,
        transitionFailed: status === 'failure' && stepNumber === currentStep - 1,
        active: stepNumber === currentStep,
        status: stepStatus,
      }

      return React.cloneElement(child, props)
    }

    return child
  })

  return <ol className={cls}>{steps}</ol>
}

Stepper.propTypes = {
  children: PropTypes.oneOfType([PropTypes.node, PropTypes.arrayOf(PropTypes.node)]),
  className: PropTypes.string,
  currentStep: PropTypes.number.isRequired,
  status: PropTypes.oneOf(['success', 'failure', 'current', 'wait']),
  stepCountStart: PropTypes.number,
  vertical: PropTypes.bool,
}

Stepper.defaultProps = {
  className: undefined,
  stepCountStart: 1,
  status: 'current',
  children: [],
  vertical: false,
}

export default Stepper
