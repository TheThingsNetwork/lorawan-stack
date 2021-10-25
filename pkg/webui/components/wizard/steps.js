// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import PropTypes from '@ttn-lw/lib/prop-types'

import { useWizardContext } from '.'

const Steps = props => {
  const { children } = props
  const { onStepsInit, currentStepId } = useWizardContext()

  const childrenRef = React.useRef(children)
  React.useEffect(() => {
    onStepsInit(
      React.Children.toArray(childrenRef.current)
        .filter(child => React.isValidElement(child) && child.type.displayName === 'Wizard.Step')
        .map(step => ({
          title: step.props.title,
          id: step.props.id,
        })),
    )
  }, [onStepsInit])

  return React.Children.toArray(children)
    .filter(child => React.isValidElement(child) && child.type.displayName === 'Wizard.Step')
    .reduce((acc, child) => {
      if (child.props.id === currentStepId) {
        return child
      }

      return acc
    }, null)
}

Steps.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
}
export default Steps
