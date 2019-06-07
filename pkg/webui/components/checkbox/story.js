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

import CheckboxGroup from './group'
import Checkbox from '.'

const info = {
  inline: true,
  header: false,
  source: false,
  propTables: [ Checkbox ],
}

storiesOf('Checkbox', module)
  .addDecorator((story, context) => withInfo(info)(story)(context))
  .add('Default', () => (
    <Checkbox
      label="Checkbox"
      name="checkbox"
    />
  ))
  .add('Disabled', () => (
    <div style={{ padding: '20px' }}>
      <Checkbox
        name="checkbox"
        label="Checkbox"
        value
        disabled
      />
      <br />
      <Checkbox
        name="checkbox"
        label="Checkbox"
        disabled
      />
    </div>
  ))
  .add('Group (horizontal)', () => (
    <div>
      <div style={{ padding: '20px' }}>
        <CheckboxGroup
          name="checkbox1"
          initialValue={{ cb1: true, cb2: true }}
          horizontal
        >
          <Checkbox
            label="Checkbox 1"
            name="cb1"
          />
          <Checkbox
            label="Checkbox 2"
            name="cb2"
          />
          <Checkbox
            label="Checkbox 3"
            name="cb3"
          />
          <Checkbox
            label="Checkbox 4"
            name="cb4"
          />
        </CheckboxGroup>
      </div>
      <div style={{ padding: '20px' }}>
        <CheckboxGroup
          name="checkbox2"
          initialValue={{}}
          horizontal
        >
          <Checkbox
            label="Checkbox 1"
            name="cb1"
          />
          <Checkbox
            label="Checkbox 2"
            name="cb2"
            disabled
          />
          <Checkbox
            label="Checkbox 3"
            name="cb3"
            disabled
          />
          <Checkbox
            label="Checkbox 4"
            name="cb4"
          />
        </CheckboxGroup>
      </div>
      <div style={{ padding: '20px' }}>
        <CheckboxGroup
          name="checkbox3"
          initialValue={{ cb1: true }}
          disabled
          horizontal
        >
          <Checkbox
            label="Checkbox 1"
            name="cb1"
          />
          <Checkbox
            label="Checkbox 2"
            name="cb2"
          />
          <Checkbox
            label="Checkbox 3"
            name="cb3"
          />
          <Checkbox
            label="Checkbox 4"
            name="cb4"
          />
        </CheckboxGroup>
      </div>
    </div>
  ))
  .add('Group (vertical)', () => (
    <div>
      <div style={{ padding: '20px' }}>
        <CheckboxGroup
          name="checkbox1"
          initialValue={{ cb1: true, cb2: true }}
          horizontal={false}
        >
          <Checkbox
            label="Checkbox 1"
            name="cb1"
          />
          <Checkbox
            label="Checkbox 2"
            name="cb2"
          />
          <Checkbox
            label="Checkbox 3"
            name="cb3"
          />
          <Checkbox
            label="Checkbox 4"
            name="cb4"
          />
        </CheckboxGroup>
      </div>
      <div style={{ padding: '20px' }}>
        <CheckboxGroup
          name="checkbox2"
          initialValue={{}}
          horizontal={false}
        >
          <Checkbox
            label="Checkbox 1"
            name="cb1"
          />
          <Checkbox
            label="Checkbox 2"
            name="cb2"
            disabled
          />
          <Checkbox
            label="Checkbox 3"
            name="cb3"
            disabled
          />
          <Checkbox
            label="Checkbox 4"
            name="cb4"
          />
        </CheckboxGroup>
      </div>
      <div style={{ padding: '20px' }}>
        <CheckboxGroup
          name="checkbox3"
          initialValue={{ cb1: true }}
          disabled
          horizontal={false}
        >
          <Checkbox
            label="Checkbox 1"
            name="cb1"
          />
          <Checkbox
            label="Checkbox 2"
            name="cb2"
          />
          <Checkbox
            label="Checkbox 3"
            name="cb3"
          />
          <Checkbox
            label="Checkbox 4"
            name="cb4"
          />
        </CheckboxGroup>
      </div>
    </div>
  ))
