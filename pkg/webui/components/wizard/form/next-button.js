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
import { defineMessages } from 'react-intl'

import Button from '@ttn-lw/components/button'
import { useFormContext } from '@ttn-lw/components/form'
import { useWizardContext } from '@ttn-lw/components/wizard'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './form.styl'

const m = defineMessages({
  next: 'Next',
  complete: 'Complete',
})

const WizardNextButton = props => {
  const { isLastStep, completeMessage, validationContext, validationSchema } = props
  const { currentStepId, steps, onStepComplete } = useWizardContext()
  const { disabled, isSubmitting, isValidating, values } = useFormContext()

  const stepIndex = steps.findIndex(({ id }) => id === currentStepId)
  const { title: nextStepTitle } = steps[Math.min(stepIndex + 1, steps.length - 1)] || {
    title: m.next,
  }

  const nextMessage = isLastStep
    ? Boolean(completeMessage)
      ? completeMessage
      : m.complete
    : nextStepTitle

  const handleClick = React.useCallback(() => {
    onStepComplete(validationSchema.cast(values, { context: validationContext }))
  }, [onStepComplete, validationContext, validationSchema, values])

  return (
    <Button
      className={style.button}
      type="submit"
      primary
      onClick={handleClick}
      disabled={disabled}
      busy={isSubmitting || isValidating}
    >
      <Message content={nextMessage} />
      <Button.Icon icon={isLastStep ? '' : 'keyboard_arrow_right'} type="right" />
    </Button>
  )
}

WizardNextButton.propTypes = {
  completeMessage: PropTypes.message,
  isLastStep: PropTypes.bool.isRequired,
  validationContext: PropTypes.shape({}).isRequired,
  validationSchema: PropTypes.shape({
    cast: PropTypes.func.isRequired,
  }).isRequired,
}

WizardNextButton.defaultProps = {
  completeMessage: undefined,
}

export default WizardNextButton
