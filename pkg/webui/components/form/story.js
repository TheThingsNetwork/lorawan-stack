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
import { action } from '@storybook/addon-actions'

import Button from '@ttn-lw/components/button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Input from '@ttn-lw/components/input'
import Checkbox from '@ttn-lw/components/checkbox'
import Radio from '@ttn-lw/components/radio-button'

import Yup from '@ttn-lw/lib/yup'

import Form from '.'

const handleSubmit = (data, { resetForm }) => {
  action('Submit')(data)
  setTimeout(() => resetForm({ values: data }), 1000)
}

const containerLoginStyles = {
  maxWidth: '400px',
}

const containerDefaultStyles = {
  maxWidth: '600px',
}

export default {
  title: 'Form',
  component: Form,
}

export const Login = () => (
  <div style={containerLoginStyles}>
    <Form
      onSubmit={handleSubmit}
      initialValues={{
        user_id: '',
        password: '',
      }}
    >
      <Form.Field title="Username or Email" name="user_id" type="text" component={Input} />
      <Form.Field title="Password" name="password" type="password" component={Input} />
      <SubmitBar>
        <Form.Submit message="Login" component={SubmitButton} />
      </SubmitBar>
      <Button naked message="Create an account" />
    </Form>
  </div>
)

export const FieldGroups = () => (
  <div style={containerDefaultStyles}>
    <Form
      onSubmit={handleSubmit}
      initialValues={{
        'radio-story': 'foo',
        'checkbox-story': { foo: true },
      }}
      submitEnabledWhenInvalid
    >
      <Form.Field name="radio-story" title="Radio Buttons" component={Radio.Group}>
        <Radio label="Foo" value="foo" name="foo" />
        <Radio label="Bar" value="bar" name="bar" />
        <Radio label="Baz" value="baz" name="baz" />
      </Form.Field>
      <Form.Field name="checkbox-story" title="Checkboxes" component={Checkbox.Group}>
        <Checkbox label="Foo" name="foo" />
        <Checkbox label="Bar" name="bar" />
        <Checkbox label="Baz" name="baz" />
      </Form.Field>
      <SubmitBar>
        <Form.Submit message="Save" component={SubmitButton} />
      </SubmitBar>
    </Form>
  </div>
)

export const Mixed = () => (
  <div style={containerDefaultStyles}>
    <Form
      validateOnBlur
      validateOnChange
      validate
      onSubmit={handleSubmit}
      validationSchema={Yup.object().shape({
        name: Yup.string().min(5, 'Too Short').max(25, 'Too Long').required('Required'),
        description: Yup.string().min(5, 'Too Short').max(50, 'Too Long'),
        checkboxes: Yup.object().test('checkboxes', 'Cannot be empty', values =>
          Object.values(values).reduce((acc, curr) => acc || curr, false),
        ),
        about: Yup.string().max(2000),
      })}
      initialValues={{
        name: '',
        description: '',
        radio: 'radio1',
        checkboxes: {},
        about: '',
      }}
    >
      <Form.SubTitle title="General information" />
      <Form.Field
        component={Input}
        type="text"
        name="name"
        placeholder="Name"
        title="Name"
        required
      />
      <Form.Field
        component={Checkbox.Group}
        name="checkboxes"
        title="Checkboxes"
        description="Choose at least one"
        required
      >
        <Checkbox name="cb1" label="Checkbox 1" />
        <Checkbox name="cb2" label="Checkbox 2" />
        <Checkbox name="cb3" label="Checkbox 3" />
      </Form.Field>
      <Form.Field component={Radio.Group} name="radio" title="Radio" required>
        <Radio label="Radio 1" value="radio1" />
        <Radio label="Radio 2" value="radio2" />
        <Radio label="Radio 3" value="radio3" />
      </Form.Field>
      <Form.CollapseSection title="Optional information" id="optional-section">
        <Form.Field
          component={Input}
          type="textarea"
          name="about"
          title="About"
          description="Tell us about yourself"
        />
      </Form.CollapseSection>
      <SubmitBar>
        <Form.Submit message="Submit" component={SubmitButton} />
      </SubmitBar>
    </Form>
  </div>
)
