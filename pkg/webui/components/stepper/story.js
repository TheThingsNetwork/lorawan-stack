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
import bind from 'autobind-decorator'

import Stepper from '.'

const steps = [
  { title: 'Step 1', description: 'Here is a description.' },
  { title: 'Step 2', description: 'Here is a description.' },
  { title: 'Step 3', description: 'Here is a description.' },
  { title: 'Step 4', description: 'Here is a description.' },
  { title: 'Step 5', description: 'Here is a description.' },
]

class InteractiveExample extends React.Component {
  state = { step: 1, error: false }

  @bind
  nextStep() {
    this.setState(prev => ({
      step: prev.step + 1,
    }))
  }

  @bind
  previousStep() {
    this.setState(prev => ({
      step: prev.step - 1,
    }))
  }

  @bind
  triggerError() {
    this.setState(prev => ({
      error: prev.error ? false : prev.step,
    }))
  }

  render() {
    const { step, error } = this.state

    const isNextDisabled = step >= steps.length
    const isPrevDisabled = step <= 1

    return (
      <div style={{ width: '100%' }}>
        <Stepper currentStep={step} status={error ? 'failure' : 'current'} {...this.props}>
          {steps.map(({ title, description }) => (
            <Stepper.Step title={title} description={description} key={title} />
          ))}
        </Stepper>
        <div style={{ marginTop: '50px', textAlign: 'center' }}>
          <button onClick={this.previousStep} disabled={isPrevDisabled}>
            Previous
          </button>
          <button onClick={this.nextStep} disabled={isNextDisabled}>
            Next
          </button>
        </div>
        <div style={{ marginTop: '20px', textAlign: 'center' }}>
          <button onClick={this.triggerError}>{error ? 'Reset error' : 'Trigger error'}</button>
        </div>
      </div>
    )
  }
}

export default {
  title: 'Stepper',
}

export const Default = () => (
  <Stepper currentStep={2}>
    <Stepper.Step stepNumber={1} title="Step 1" description="Here is a description." />
    <Stepper.Step stepNumber={2} title="Step 2" description="Here is a description." />
    <Stepper.Step stepNumber={3} title="Step 3" description="Here is a description." />
  </Stepper>
)

export const DefaultError = () => (
  <Stepper status="failure" currentStep={2}>
    <Stepper.Step stepNumber={1} title="Step 1" description="Here is a description." />
    <Stepper.Step stepNumber={2} title="Step 2" description="Here is a description." />
    <Stepper.Step stepNumber={3} title="Step 3" description="Here is a description." />
  </Stepper>
)

DefaultError.story = {
  name: 'Default (Error)',
}

export const DefaultInteractive = () => <InteractiveExample />

DefaultInteractive.story = {
  name: 'Default (Interactive)',
}

export const Vertical = () => (
  <Stepper currentStep={2} vertical>
    <Stepper.Step stepNumber={1} title="Step 1" description="Here is a description." />
    <Stepper.Step stepNumber={2} title="Step 2" description="Here is a description." />
    <Stepper.Step stepNumber={3} title="Step 3" description="Here is a description." />
  </Stepper>
)

export const VerticalError = () => (
  <Stepper status="failure" currentStep={2} vertical>
    <Stepper.Step stepNumber={1} title="Step 1" description="Here is a description." />
    <Stepper.Step stepNumber={2} title="Step 2" description="Here is a description." />
    <Stepper.Step stepNumber={3} title="Step 3" description="Here is a description." />
  </Stepper>
)

VerticalError.story = {
  name: 'Vertical (Error)',
}

export const VerticalInteractive = () => <InteractiveExample vertical />

VerticalInteractive.story = {
  name: 'Vertical (Interactive)',
}
