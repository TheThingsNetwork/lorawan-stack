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
import { storiesOf } from '@storybook/react'
import { withInfo } from '@storybook/addon-info'

import RadioGroup from './group'
import Radio from '.'

const info = {
  inline: true,
  header: false,
  source: false,
  propTables: [ Radio ],
}

storiesOf('Radio', module)
  .addDecorator((story, context) => withInfo(info)(story)(context))
  .add('Default', () => (
    <Radio
      label="Radio"
      name="radio"
      value="1"
    />
  ))
  .add('Disabled', () => (
    <div style={{ padding: '20px' }}>
      <Radio
        name="radio"
        label="Radio 1"
        value="1"
        checked
        disabled
      />
      <br />
      <Radio
        name="radio"
        label="Radio 2"
        value="2"
        disabled
      />
    </div>
  ))
  .add('Group (horizontal)', () => (
    <div>
      <div style={{ padding: '20px' }}>
        <RadioGroup name="radio" initialValue="1" horizontal>
          <Radio
            label="Radio 1"
            value="1"
          />
          <Radio
            label="Radio 2"
            value="2"
          />
          <Radio
            label="Radio 3"
            value="3"
          />
          <Radio
            label="Radio 4"
            value="4"
          />
        </RadioGroup>
      </div>
      <div style={{ padding: '20px' }}>
        <RadioGroup name="radio-with-disabled" initialValue="1" horizontal>
          <Radio
            label="Radio 1"
            value="1"
          />
          <Radio
            label="Radio 2"
            value="2"
            disabled
          />
          <Radio
            label="Radio 3"
            value="3"
            disabled
          />
          <Radio
            label="Radio 4"
            value="4"
          />
        </RadioGroup>
      </div>
      <div style={{ padding: '20px' }}>
        <RadioGroup name="radio-all-disabled" initialValue="1" disabled horizontal>
          <Radio
            label="Radio 1"
            value="1"
          />
          <Radio
            label="Radio 2"
            value="2"
          />
          <Radio
            label="Radio 3"
            value="3"
          />
          <Radio
            label="Radio 4"
            value="4"
          />
        </RadioGroup>
      </div>
    </div>
  ))
  .add('Group (vertical)', () => (
    <div>
      <div style={{ padding: '20px' }}>
        <RadioGroup name="radio" initialValue="1" horizontal={false}>
          <Radio
            label="Radio 1"
            value="1"
          />
          <Radio
            label="Radio 2"
            value="2"
          />
          <Radio
            label="Radio 3"
            value="3"
          />
          <Radio
            label="Radio 4"
            value="4"
          />
        </RadioGroup>
      </div>
      <div style={{ padding: '20px' }}>
        <RadioGroup name="radio-with-disabled" initialValue="1" horizontal={false}>
          <Radio
            label="Radio 1"
            value="1"
          />
          <Radio
            label="Radio 2"
            value="2"
            disabled
          />
          <Radio
            label="Radio 3"
            value="3"
            disabled
          />
          <Radio
            label="Radio 4"
            value="4"
          />
        </RadioGroup>
      </div>
      <div style={{ padding: '20px' }}>
        <RadioGroup name="radio-all-disabled" initialValue="1" disabled horizontal={false}>
          <Radio
            label="Radio 1"
            value="1"
          />
          <Radio
            label="Radio 2"
            value="2"
          />
          <Radio
            label="Radio 3"
            value="3"
          />
          <Radio
            label="Radio 4"
            value="4"
          />
        </RadioGroup>
      </div>
    </div>
  ))
