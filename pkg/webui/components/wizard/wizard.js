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
import { merge } from 'lodash'

import PropTypes from '../../lib/prop-types'
import renderCallback from '../../lib/render-callback'

import { WizardContext } from './context'

const FIRST_STEP = 1

// Action types.
const INIT = 'INIT'
const GO_TO_STEP = 'GO_TO_STEP'

const reducer = (state, action) => {
  switch (action.type) {
    case INIT:
      return {
        ...state,
        steps: action.steps,
      }
    case GO_TO_STEP:
      const { snapshots: oldSnapshots, currentStep: oldStep } = state
      const { values, step: currentStep } = action

      // Replace current step values when navigating between wizard steps.
      const snapshots = [
        ...oldSnapshots.slice(0, oldStep - 1),
        values,
        ...oldSnapshots.slice(oldStep),
      ]

      return {
        ...state,
        currentStep,
        snapshots,
      }
    default:
      return state
  }
}

const Wizard = props => {
  const { initialStep, onComplete, initialValues, completeMessage } = props

  const [state, dispatch] = React.useReducer(reducer, {
    // Active step in the wizard.
    currentStep: initialStep,
    // A list of all steps in the wizard.
    steps: [],
    // A list of form values for each step in the wizard.
    // For example, `snapshots[0]` - has form values after submitting the first step.
    snapshots: [],
  })

  const { currentStep, steps, snapshots } = state

  const stepsCount = steps.length

  const stepsInit = React.useCallback(steps => {
    dispatch({ type: INIT, steps })
  }, [])
  const prevStep = React.useCallback(
    values => dispatch({ type: GO_TO_STEP, step: Math.max(currentStep - 1, FIRST_STEP), values }),
    [currentStep],
  )
  const nextStep = React.useCallback(
    values => dispatch({ type: GO_TO_STEP, step: Math.min(currentStep + 1, stepsCount), values }),
    [currentStep, stepsCount],
  )

  const snapshot = React.useMemo(() => merge({}, initialValues, ...snapshots), [
    initialValues,
    snapshots,
  ])

  const context = {
    completeMessage,
    onComplete,
    onNextStep: nextStep,
    onPrevStep: prevStep,
    onStepsInit: stepsInit,
    currentStep,
    snapshot,
    steps,
  }

  return (
    <WizardContext.Provider value={context}>
      {renderCallback(props, context)}
    </WizardContext.Provider>
  )
}

Wizard.propTypes = {
  completeMessage: PropTypes.message,
  initialStep: PropTypes.number,
  initialValues: PropTypes.shape({}),
  onComplete: PropTypes.func.isRequired,
}

Wizard.defaultProps = {
  initialValues: {},
  initialStep: 1,
  completeMessage: undefined,
}

export default Wizard
