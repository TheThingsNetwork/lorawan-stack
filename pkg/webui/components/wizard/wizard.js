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

const FIRST_STEP = 0

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
      if (values) {
      }
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

const Wizard = React.forwardRef((props, ref) => {
  const { initialStepId, onComplete, initialValues, completeMessage } = props

  const [state, dispatch] = React.useReducer(reducer, {
    // Active step in the wizard.
    currentStepId: initialStepId,
    // A list of all steps in the wizard.
    steps: [],
    // A list of form values for each step in the wizard.
    // For example, `snapshots[0]` - has form values after submitting the first step.
    snapshots: [],
  })

  const { currentStepId, steps, snapshots } = state

  const stepsInit = React.useCallback(steps => {
    dispatch({ type: INIT, steps })
  }, [])
  const goToStep = React.useCallback(stepId => {
    dispatch({ type: GO_TO_STEP, step: stepId })
  }, [])
  const prevStep = React.useCallback(
    values => {
      const currentStepIndex = steps.findIndex(({ id }) => id === currentStepId)
      const prevStep = steps[Math.max(currentStepIndex - 1, FIRST_STEP)] || {}

      dispatch({ type: GO_TO_STEP, step: prevStep.id || currentStepId, values })
    },
    [currentStepId, steps],
  )
  const nextStep = React.useCallback(
    values => {
      const currentStepIndex = steps.findIndex(({ id }) => id === currentStepId)
      const nextStep = steps[Math.min(currentStepIndex + 1, steps.length - 1)] || {}

      dispatch({ type: GO_TO_STEP, step: nextStep.id || currentStepId, values })
    },
    [currentStepId, steps],
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
    onStepChange: goToStep,
    currentStepId,
    snapshot,
    steps,
  }

  React.useImperativeHandle(ref, () => context)

  return (
    <WizardContext.Provider value={context}>
      {renderCallback(props, context)}
    </WizardContext.Provider>
  )
})

Wizard.propTypes = {
  completeMessage: PropTypes.message,
  initialStepId: PropTypes.string.isRequired,
  initialValues: PropTypes.shape({}),
  onComplete: PropTypes.func.isRequired,
}

Wizard.defaultProps = {
  initialValues: {},
  completeMessage: undefined,
}

export default Wizard
