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

import SubmitButton from '@ttn-lw/components/submit-button'
import Button from '@ttn-lw/components/button'
import Form from '@ttn-lw/components/form'
import { useWizardContext } from '@ttn-lw/components/wizard'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  next: 'Next',
  complete: 'Complete',
})

const WizardNextButton = props => {
  const { isLastStep, completeMessage } = props
  const { currentStep, steps } = useWizardContext()

  const { title: nextStepTitle } = steps.find(
    ({ stepNumber }) => stepNumber === currentStep + 1,
  ) || { title: m.next }

  const nextMessage = isLastStep
    ? Boolean(completeMessage)
      ? completeMessage
      : m.complete
    : nextStepTitle

  return (
    <Form.Submit component={SubmitButton}>
      <Message content={nextMessage} />
      {!isLastStep && <Button.Icon icon="keyboard_arrow_right" type="right" />}
    </Form.Submit>
  )
}

WizardNextButton.propTypes = {
  completeMessage: PropTypes.message,
  isLastStep: PropTypes.bool.isRequired,
}

WizardNextButton.defaultProps = {
  completeMessage: undefined,
}

export default WizardNextButton
