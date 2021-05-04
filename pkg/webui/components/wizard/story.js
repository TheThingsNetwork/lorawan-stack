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
import { storiesOf } from '@storybook/react'
import { action } from '@storybook/addon-actions'

import Wizard, { WizardContext } from '@ttn-lw/components/wizard'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Select from '@ttn-lw/components/select'

import Yup from '@ttn-lw/lib/yup'

const Debug = () => (
  <div>
    <div
      style={{
        textTransform: 'uppercase',
        fontWeight: 600,
        padding: '.5rem',
        background: 'gray',
        color: 'white',
      }}
    >
      wizard state
    </div>
    <WizardContext.Consumer>
      {({ onNextStep, onPrevStep, onStepsInit, steps, ...rest }) => {
        const formattedSteps = steps.map(({ title, id }) => ({ title, id }))

        return (
          <pre
            style={{
              fontSize: '.85rem',
              padding: '.25rem .5rem',
              overflowX: 'scroll',
            }}
          >
            {JSON.stringify({ ...rest, steps: formattedSteps }, null, 2)}
          </pre>
        )
      }}
    </WizardContext.Consumer>
  </div>
)

const stepSubmit = desc => data => action(desc)(data)
const onComplete = data => action('OnComplete')(data)

storiesOf('Wizard', module).add('Basic', () => (
  <Wizard onComplete={onComplete} completeMessage="Create account" initialStepId="1">
    <Wizard.Stepper>
      <Wizard.Stepper.Step title="Account settings" description="E-mail and password" />
      <Wizard.Stepper.Step title="Personal information" description="Year of birth and gender" />
      <Wizard.Stepper.Step title="Details" description="Account comment" />
    </Wizard.Stepper>
    <Wizard.Steps>
      <Wizard.Step title="Account settings" id="1">
        <Wizard.Form
          validationSchema={Yup.object({
            email: Yup.string().email().required(),
            password: Yup.string().required(),
          }).noUnknown()}
          initialValues={{ email: '', password: '' }}
          onSubmit={stepSubmit('step1')}
        >
          <Form.Field component={Input} type="text" name="email" title="E-mail" required />
          <Form.Field component={Input} type="password" name="password" title="Password" required />
        </Wizard.Form>
      </Wizard.Step>
      <Wizard.Step title="Personal information" id="2">
        <Wizard.Form
          onSubmit={stepSubmit('step2')}
          validationSchema={Yup.object({
            year: Yup.number().min(1900).max(1999).required(),
            gender: Yup.string().required(),
          }).noUnknown()}
          initialValues={{ year: 0, gender: '' }}
        >
          <Form.Field component={Input} type="number" name="year" title="Year of birth" required />
          <Form.Field
            required
            component={Select}
            name="gender"
            title="Gender"
            options={[
              { value: 'male', label: 'Male' },
              { value: 'female', label: 'Female' },
              { value: 'other', label: 'Other' },
            ]}
          />
        </Wizard.Form>
      </Wizard.Step>
      <Wizard.Step title="Details" id="3">
        <Wizard.Form
          onSubmit={stepSubmit('step3')}
          validationSchema={Yup.object({
            comment: Yup.string().max(2000),
          }).noUnknown()}
          initialValues={{ comment: '' }}
        >
          <Form.Field component={Input} type="textarea" name="comment" title="Comment" />
        </Wizard.Form>
      </Wizard.Step>
    </Wizard.Steps>
    <Debug />
  </Wizard>
))
