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

const INIT = 'INIT'
const GO_TO_STEP = 'GO_TO_STEP'
const SET_SNAPSHOT = 'SET_SNAPSHOT'

const reducer = (state, action) => {
  switch (action.type) {
    case INIT:
      return {
        ...state,
        steps: action.steps,
      }
    case GO_TO_STEP:
      return {
        ...state,
        currentStep: action.step,
      }
    case SET_SNAPSHOT:
      return {
        ...state,
        snapshot: merge({}, state.snapshot, action.snapshot),
      }
    default:
      return state
  }
}

const Wizard = props => {
  const { initialStep, onComplete, initialValues, completeMessage } = props

  const [state, dispatch] = React.useReducer(reducer, {
    currentStep: initialStep,
    steps: [],
    snapshot: initialValues,
  })

  const { currentStep, steps, snapshot } = state

  const stepsCount = steps.length

  const stepsInit = React.useCallback(steps => {
    dispatch({ type: INIT, steps })
  }, [])
  const prevStep = React.useCallback(
    values => {
      dispatch({ type: SET_SNAPSHOT, snapshot: values })
      dispatch({ type: GO_TO_STEP, step: Math.min(currentStep - 1, stepsCount) })
    },
    [currentStep, stepsCount],
  )
  const nextStep = React.useCallback(
    values => {
      dispatch({ type: SET_SNAPSHOT, snapshot: values })
      dispatch({ type: GO_TO_STEP, step: Math.min(currentStep + 1, stepsCount) })
    },
    [currentStep, stepsCount],
  )

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
