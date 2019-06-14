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
import { action } from '@storybook/addon-actions'
import { withInfo } from '@storybook/addon-info'
import * as Yup from 'yup'

import Button from '../button'
import SubmitBar from '../submit-bar'
import SubmitButton from '../submit-button'
import Input from '../input'
import Form from '../form'
import Checkbox from '../checkbox'
import Radio from '../radio-button'

const handleSubmit = function (data, { resetForm }) {
  action('Submit')(data)
  setTimeout(() => resetForm(data), 1000)
}

const containerStyles = {
  maxWidth: '400px',
}

const containerHorizontalStyles = {
  maxWidth: '600px',
}

storiesOf('Form', module)
  .addDecorator((story, context) => withInfo({
    inline: true,
    header: false,
    source: true,
    propTables: [ Form ],
  })(story)(context))
  .add('Login', () => (
    <div style={containerStyles}>
      <Form
        onSubmit={handleSubmit}
        initialValues={{
          user_id: '',
          password: '',
        }}
      >
        <Form.Field
          title="Username or Email"
          name="user_id"
          type="text"
          component={Input}
        />
        <Form.Field
          title="Password"
          name="password"
          type="password"
          component={Input}
        />
        <SubmitBar>
          <Form.Submit message="Login" component={SubmitButton} />
        </SubmitBar>
        <Button naked message="Create an account" />
      </Form>
    </div>
  ))
  .add('Field Groups', () => (
    <div style={containerHorizontalStyles}>
      <Form
        onSubmit={handleSubmit}
        initialValues={{
          'radio-story': 'foo',
          'checkbox-story': { foo: true },
        }}
        submitEnabledWhenInvalid
        horizontal
      >
        <Form.Field
          name="radio-story"
          title="Radio Buttons"
          component={Radio.Group}
        >
          <Radio
            label="Foo"
            value="foo"
            name="foo"
          />
          <Radio
            label="Bar"
            value="bar"
            name="bar"
          />
          <Radio
            label="Baz"
            value="baz"
            name="baz"
          />
        </Form.Field>
        <Form.Field
          name="checkbox-story"
          title="Checkboxes"
          component={Checkbox.Group}
        >
          <Checkbox
            label="Foo"
            name="foo"
          />
          <Checkbox
            label="Bar"
            name="bar"
          />
          <Checkbox
            label="Baz"
            name="baz"
          />
        </Form.Field>
        <SubmitBar>
          <Form.Submit message="Save" component={SubmitButton} />
        </SubmitBar>
      </Form>
    </div>
  ))
  .add('Mixed', () => (
    <div style={containerHorizontalStyles}>
      <Form
        validateOnBlur
        validateOnChange
        validate
        horizontal
        onSubmit={handleSubmit}
        validationSchema={Yup.object().shape({
          name: Yup.string()
            .min(5, 'Too Short')
            .max(25, 'Too Long')
            .required('Required'),
          description: Yup.string()
            .min(5, 'Too Short')
            .max(50, 'Too Long'),
          checkboxes: Yup.object().test(
            'checkboxes',
            'Cannot be empty',
            values => Object.values(values).reduce((acc, curr) => acc || curr, false)
          ),
        })}
        initialValues={{
          name: '',
          description: '',
          radio: 'radio1',
          checkboxes: {},
        }}
      >
        <Form.Field
          component={Input}
          type="text"
          name="name"
          placeholder="Name"
          title="Name"
          required
        />
        <Form.Field
          component={Input}
          type="text"
          name="description"
          placeholder="Description"
          title="Description"
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
        <Form.Field
          component={Radio.Group}
          name="radio"
          title="Radio"
          required
        >
          <Radio label="Radio 1" value="radio1" />
          <Radio label="Radio 2" value="radio2" />
          <Radio label="Radio 3" value="radio3" />
        </Form.Field>
        <SubmitBar>
          <Form.Submit message="Submit" component={SubmitButton} />
        </SubmitBar>
      </Form>
    </div>
  ))
