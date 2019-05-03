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

import FieldGroup from './group'
import { Field } from '.'

const info = {
  inline: true,
  header: false,
  source: false,
  propTables: [ Field ],
}

const checkboxFields = [
  <Field
    type="radio"
    value="foo"
    title="Foo"
    name="radio-story"
    form={false}
    key="one"
  />,
  <Field
    type="radio"
    value="bar"
    title="Bar"
    name="radio-story"
    form={false}
    key="two"
  />,
]

storiesOf('Fields', module)
  .addDecorator((story, context) => withInfo(info)(story)(context))
  .add('Default', () => (
    <Field
      title="Foo"
      name="foo"
      description="Foo is a required field"
      required
      form={false}
    />
  ))
  .add('Horizontal', () => (
    <div style={{ width: '60%' }}>
      <Field
        title="Foo"
        name="foo"
        required
        horizontal
        form={false}
      />
      <Field
        title="Bar"
        name="bar"
        required
        horizontal
        form={false}
      />
    </div>

  ))
  .add('Number', () => (
    <Field
      title="Foo"
      name="foo"
      description="Foo is a field containing a number"
      type="number"
      form={false}
    />
  ))
  .add('Boolean', () => (
    <Field
      type="checkbox"
      name="foo"
      title="Foo"
      description="Foo field."
      form={false}
    />
  ))
  .add('Select', () => (
    <Field
      type="select"
      name="foo"
      title="Foo"
      description="A Select field."
      options={[
        { value: 'amsterdam', label: 'Amsterdam' },
        { value: 'berlin', label: 'Berlin' },
        { value: 'dusseldorf', label: 'Düsseldorf' },
      ]}
      form={false}
    />
  ))
  .add('Error', () => (
    <Field
      name="email"
      title="Email Address"
      description="Your email address"
      value="foo?@bar.com"
      type="text"
      error="Invalid email address"
      touched
      form={false}
    />
  ))
  .add('Boolean Disabled', () => (
    <Field
      type="checkbox"
      name="foo"
      title="Foo"
      description="Foo field."
      form={false}
      disabled
    />
  ))
  .add('Boolean Error', () => (
    <Field
      type="checkbox"
      name="foo"
      title="Foo"
      description="Foo field."
      error="Are you sure?"
      touched
      form={false}
    />
  ))
  .add('Warning', () => (
    <Field
      name="password"
      title="Password"
      description="Create a new password."
      value="123456"
      type="password"
      warning="Insecure password"
      touched
      form={false}
    />
  ))

storiesOf('Fields/Radio Buttons', module)
  .addDecorator((story, context) => withInfo(info)(story)(context))
  .add('Default', () => (
    <FieldGroup
      name="radio-story"
      title="Radio Buttons"
      children={checkboxFields}
    />
  ))
  .add('Disabled', () => (
    <FieldGroup
      name="radio-story"
      title="Radio Buttons"
      children={checkboxFields}
      disabled
    />
  ))
  .add('Horizontal', () => (
    <FieldGroup
      name="radio-story"
      title="Radio Buttons"
      horizontal
      children={checkboxFields}
    />
  ))
  .add('Columns', () => (
    <FieldGroup
      name="radio-story"
      title="Radio Buttons"
      columns
      children={checkboxFields}
    />
  ))
  .add('Horizontal & Columns', () => (
    <FieldGroup
      name="radio-story"
      title="Radio Buttons"
      horizontal
      columns
      children={checkboxFields}
    />
  ))
  .add('Horizontal & Error', () => (
    <FieldGroup
      name="radio-story"
      title="Radio Buttons"
      error="Test Error!"
      touched
      horizontal
      children={checkboxFields}
    />
  ))
