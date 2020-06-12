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

import { useFormContext } from '@ttn-lw/components/form'
import Button from '@ttn-lw/components/button'
import { useWizardContext } from '@ttn-lw/components/wizard'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  prev: 'Previous',
})

const WizardPrevButton = props => {
  const { isFirstStep, validationSchema, validationContext } = props
  const { onPrevStep, currentStep, steps } = useWizardContext()
  const { values } = useFormContext()

  const handlePrevStep = React.useCallback(() => {
    onPrevStep(
      validationSchema.cast(values, {
        context: validationContext,
      }),
    )
  }, [onPrevStep, validationContext, validationSchema, values])

  if (isFirstStep) {
    return null
  }

  const { title: prevMessage } = steps.find(({ stepNumber }) => stepNumber === currentStep - 1) || {
    title: m.next,
  }

  return (
    <Button secondary onClick={handlePrevStep} type="button">
      <Button.Icon icon="keyboard_arrow_left" type="left" />
      <Message content={prevMessage} />
    </Button>
  )
}

WizardPrevButton.propTypes = {
  isFirstStep: PropTypes.bool.isRequired,
  validationContext: PropTypes.shape({}).isRequired,
  validationSchema: PropTypes.shape({
    cast: PropTypes.func.isRequired,
  }).isRequired,
}

export default WizardPrevButton
