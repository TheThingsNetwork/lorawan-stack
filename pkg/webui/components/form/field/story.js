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
import * as Yup from 'yup'
import { storiesOf } from '@storybook/react'
import { action } from '@storybook/addon-actions'
import { withInfo } from '@storybook/addon-info'

import Input from '../../input'
import Checkbox from '../../checkbox'
import Radio from '../../radio-button'
import Select from '../../select'
import Form from '..'

const handleSubmit = function (data, { resetForm }) {
  action('Submit')(data)
  setTimeout(() => resetForm(data), 1000)
}

const info = {
  inline: true,
  header: false,
  source: false,
  propTables: [ Form.Field ],
}

const errorSchema = Yup.string().test('error', 'Something went wrong.', () => false)
const validationSchema = Yup.object().shape({
  error: errorSchema,
})

class FieldsWrapperExample extends React.Component {

  form = React.createRef()

  static defaultProps = {
    horizontal: true,
  }

  componentDidMount () {
    if (this.form.current) {
      this.form.current.setFieldError('error', 'Something went wrong.')
      this.form.current.setFieldTouched('error')
    }
  }

  render () {
    return (
      <Form
        onSubmit={handleSubmit}
        initialValues={this.props.initialValues}
        formikRef={this.form}
        horizontal={this.props.horizontal}
        validationSchema={validationSchema}
      >
        {this.props.children}
      </Form>
    )
  }
}

storiesOf('Fields/Select', module)
  .addDecorator((story, context) => withInfo(info)(story)(context))
  .add('Horizontal', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': 'amsterdam',
        description: 'amsterdam',
        error: 'amsterdam',
        warning: 'amsterdam',
        disabled: 'amsterdam',
      }}
    >
      <Form.Field
        name="default"
        title="Default"
        component={Select}
        options={[
          { value: 'amsterdam', label: 'Amsterdam' },
          { value: 'berlin', label: 'Berlin' },
          { value: 'dusseldorf', label: 'Düsseldorf' },
        ]}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        component={Select}
        options={[
          { value: 'amsterdam', label: 'Amsterdam' },
          { value: 'berlin', label: 'Berlin' },
          { value: 'dusseldorf', label: 'Düsseldorf' },
        ]}
      />
      <Form.Field
        name="error"
        title="With Error"
        component={Select}
        options={[
          { value: 'amsterdam', label: 'Amsterdam' },
          { value: 'berlin', label: 'Berlin' },
          { value: 'dusseldorf', label: 'Düsseldorf' },
        ]}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        component={Select}
        options={[
          { value: 'amsterdam', label: 'Amsterdam' },
          { value: 'berlin', label: 'Berlin' },
          { value: 'dusseldorf', label: 'Düsseldorf' },
        ]}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        component={Select}
        options={[
          { value: 'amsterdam', label: 'Amsterdam' },
          { value: 'berlin', label: 'Berlin' },
          { value: 'dusseldorf', label: 'Düsseldorf' },
        ]}
      />
    </FieldsWrapperExample>
  ))
  .add('Vertical', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': 'amsterdam',
        description: 'amsterdam',
        warning: 'amsterdam',
        error: 'amsterdam',
        disabled: 'amsterdam',
      }}
      horizontal={false}
    >
      <Form.Field
        name="default"
        title="Default"
        horizontal={false}
        component={Select}
        options={[
          { value: 'amsterdam', label: 'Amsterdam' },
          { value: 'berlin', label: 'Berlin' },
          { value: 'dusseldorf', label: 'Düsseldorf' },
        ]}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        horizontal={false}
        component={Select}
        options={[
          { value: 'amsterdam', label: 'Amsterdam' },
          { value: 'berlin', label: 'Berlin' },
          { value: 'dusseldorf', label: 'Düsseldorf' },
        ]}
      />
      <Form.Field
        name="error"
        title="With Error"
        horizontal={false}
        component={Select}
        options={[
          { value: 'amsterdam', label: 'Amsterdam' },
          { value: 'berlin', label: 'Berlin' },
          { value: 'dusseldorf', label: 'Düsseldorf' },
        ]}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        horizontal={false}
        component={Select}
        options={[
          { value: 'amsterdam', label: 'Amsterdam' },
          { value: 'berlin', label: 'Berlin' },
          { value: 'dusseldorf', label: 'Düsseldorf' },
        ]}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        horizontal={false}
        component={Select}
        options={[
          { value: 'amsterdam', label: 'Amsterdam' },
          { value: 'berlin', label: 'Berlin' },
          { value: 'dusseldorf', label: 'Düsseldorf' },
        ]}
      />
    </FieldsWrapperExample>
  ))

storiesOf('Fields/Checkbox', module)
  .addDecorator((story, context) => withInfo(info)(story)(context))
  .add('Horizontal', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': true,
        description: false,
        warning: false,
        error: false,
        disabled: false,
      }}
    >
      <Form.Field
        name="default"
        title="Default"
        component={Checkbox}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        component={Checkbox}
      />
      <Form.Field
        name="error"
        title="With Error"
        component={Checkbox}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        component={Checkbox}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        component={Checkbox}
      />
    </FieldsWrapperExample>
  ))
  .add('Horizontal Group', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': {
          default1: false,
          default2: false,
          default3: false,
        },
        description: {
          description1: false,
          description2: false,
          description3: false,
        },
        warning: {
          warning1: false,
          warning2: false,
          warning3: false,
        },
        error: {
          error1: false,
          error2: false,
          error3: false,
        },
        disabled: {
          disabled1: false,
          disabled2: false,
          disabled3: false,
        },
      }}
    >
      <Form.Field
        name="default"
        title="Default"
        component={Checkbox.Group}
      >
        <Checkbox name="default1" label="Checkbox 1" />
        <Checkbox name="default2" label="Checkbox 2" />
        <Checkbox name="default3" label="Checkbox 3" />
      </Form.Field>
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        component={Checkbox.Group}
      >
        <Checkbox name="description1" label="Checkbox 1" />
        <Checkbox name="description2" label="Checkbox 2" />
        <Checkbox name="description3" label="Checkbox 3" />
      </Form.Field>
      <Form.Field
        name="error"
        title="With Error"
        component={Checkbox.Group}
      >
        <Checkbox name="error1" label="Checkbox 1" />
        <Checkbox name="error2" label="Checkbox 2" />
        <Checkbox name="error3" label="Checkbox 3" />
      </Form.Field>
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        component={Checkbox.Group}
      >
        <Checkbox name="warning1" label="Checkbox 1" />
        <Checkbox name="warning2" label="Checkbox 2" />
        <Checkbox name="warning3" label="Checkbox 3" />
      </Form.Field>
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        component={Checkbox.Group}
      >
        <Checkbox name="disabled1" label="Checkbox 1" />
        <Checkbox name="disabled2" label="Checkbox 2" />
        <Checkbox name="disabled3" label="Checkbox 3" />
      </Form.Field>
    </FieldsWrapperExample>
  ))
  .add('Columns', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': false,
        description: false,
        warning: false,
        error: false,
        disabled: false,
      }}
      horizontal={false}
    >
      <Form.Field
        name="default"
        title="Default"
        component={Checkbox}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        component={Checkbox}
      />
      <Form.Field
        name="error"
        title="With Error"
        component={Checkbox}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        component={Checkbox}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        component={Checkbox}
      />
    </FieldsWrapperExample>
  ))
  .add('Columns Group', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': {
          default1: false,
          default2: false,
          default3: false,
        },
        description: {
          description1: false,
          description2: false,
          description3: false,
        },
        warning: {
          warning1: false,
          warning2: false,
          warning3: false,
        },
        error: {
          error1: false,
          error2: false,
          error3: false,
        },
        disabled: {
          disabled1: false,
          disabled2: false,
          disabled3: false,
        },
      }}
      horizontal={false}
    >
      <Form.Field
        name="default"
        title="Default"
        component={Checkbox.Group}
      >
        <Checkbox name="default1" label="Checkbox 1" />
        <Checkbox name="default2" label="Checkbox 2" />
        <Checkbox name="default3" label="Checkbox 3" />
      </Form.Field>
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        component={Checkbox.Group}
      >
        <Checkbox name="description1" label="Checkbox 1" />
        <Checkbox name="description2" label="Checkbox 2" />
        <Checkbox name="description3" label="Checkbox 3" />
      </Form.Field>
      <Form.Field
        name="error"
        title="With Error"
        component={Checkbox.Group}
      >
        <Checkbox name="error1" label="Checkbox 1" />
        <Checkbox name="error2" label="Checkbox 2" />
        <Checkbox name="error3" label="Checkbox 3" />
      </Form.Field>
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        component={Checkbox.Group}
      >
        <Checkbox name="warning1" label="Checkbox 1" />
        <Checkbox name="warning2" label="Checkbox 2" />
        <Checkbox name="warning3" label="Checkbox 3" />
      </Form.Field>
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        component={Checkbox.Group}
      >
        <Checkbox name="disabled1" label="Checkbox 1" />
        <Checkbox name="disabled2" label="Checkbox 2" />
        <Checkbox name="disabled3" label="Checkbox 3" />
      </Form.Field>
    </FieldsWrapperExample>
  ))

storiesOf('Fields/Radio', module)
  .addDecorator((story, context) => withInfo(info)(story)(context))
  .add('Horizontal', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': '1',
        description: '2',
        warning: '3',
        error: '4',
        disabled: '5',
      }}
    >
      <Form.Field
        name="default"
        title="Default"
        label="Radio"
        checked
        component={Radio}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        label="Radio"
        checked
        component={Radio}
      />
      <Form.Field
        name="error"
        title="With Error"
        label="Radio"
        checked
        component={Radio}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        label="Radio"
        checked
        component={Radio}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        label="Radio"
        disabled
        checked
        component={Radio}
      />
    </FieldsWrapperExample>
  ))
  .add('Horizontal Group', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': '1',
        description: '1',
        warning: '1',
        error: '1',
        disabled: '1',
      }}
    >
      <Form.Field
        name="default"
        title="Default"
        component={Radio.Group}
      >
        <Radio label="Radio 1" value="1" />
        <Radio label="Radio 2" value="2" />
        <Radio label="Radio 3" value="3" />
      </Form.Field>
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        component={Radio.Group}
      >
        <Radio label="Radio 1" value="1" />
        <Radio label="Radio 2" value="2" />
        <Radio label="Radio 3" value="3" />
      </Form.Field>
      <Form.Field
        name="error"
        title="With Error"
        component={Radio.Group}
      >
        <Radio label="Radio 1" value="1" />
        <Radio label="Radio 2" value="2" />
        <Radio label="Radio 3" value="3" />
      </Form.Field>
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        component={Radio.Group}
      >
        <Radio label="Radio 1" value="1" />
        <Radio label="Radio 2" value="2" />
        <Radio label="Radio 3" value="3" />
      </Form.Field>
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        component={Radio.Group}
      >
        <Radio label="Radio 1" value="1" />
        <Radio label="Radio 2" value="2" />
        <Radio label="Radio 3" value="3" />
      </Form.Field>
    </FieldsWrapperExample>
  ))
  .add('Columns', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': '1',
        description: '2',
        warning: '3',
        error: '4',
        disabled: '5',
      }}
      horizontal={false}
    >
      <Form.Field
        name="default"
        title="Default"
        label="Radio"
        component={Radio}
      />
      <Form.Field
        name="description"
        title="With Description"
        label="Radio"
        description="A select field."
        component={Radio}
      />
      <Form.Field
        name="error"
        title="With Error"
        label="Radio"
        component={Radio}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        label="Radio"
        component={Radio}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        label="Radio"
        component={Radio}
      />
    </FieldsWrapperExample>
  ))
  .add('Columns Group', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': '1',
        description: '1',
        warning: '1',
        error: '1',
        disabled: '1',
      }}
      horizontal={false}
    >
      <Form.Field
        name="default"
        title="Default"
        component={Radio.Group}
      >
        <Radio label="Radio 1" value="1" />
        <Radio label="Radio 2" value="2" />
        <Radio label="Radio 3" value="3" />
      </Form.Field>
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        component={Radio.Group}
      >
        <Radio label="Radio 1" value="1" />
        <Radio label="Radio 2" value="2" />
        <Radio label="Radio 3" value="3" />
      </Form.Field>
      <Form.Field
        name="error"
        title="With Error"
        component={Radio.Group}
      >
        <Radio label="Radio 1" value="1" />
        <Radio label="Radio 2" value="2" />
        <Radio label="Radio 3" value="3" />
      </Form.Field>
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        component={Radio.Group}
      >
        <Radio label="Radio 1" value="1" />
        <Radio label="Radio 2" value="2" />
        <Radio label="Radio 3" value="3" />
      </Form.Field>
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        component={Radio.Group}
      >
        <Radio label="Radio 1" value="1" />
        <Radio label="Radio 2" value="2" />
        <Radio label="Radio 3" value="3" />
      </Form.Field>
    </FieldsWrapperExample>
  ))

storiesOf('Fields/Input', module)
  .addDecorator((story, context) => withInfo(info)(story)(context))
  .add('Horizontal', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': 'something...',
        description: 'something...',
        warning: 'something...',
        error: 'something...',
        disabled: 'something...',
      }}
    >
      <Form.Field
        name="default"
        title="Default"
        component={Input}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        component={Input}
      />
      <Form.Field
        name="error"
        title="With Error"
        component={Input}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        component={Input}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        component={Input}
      />
    </FieldsWrapperExample>
  ))
  .add('Vertical', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': 'something...',
        description: 'something...',
        warning: 'something...',
        error: 'something...',
        disabled: 'something...',
      }}
      horizontal={false}
    >
      <Form.Field
        name="default"
        title="Default"
        component={Input}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        component={Input}
      />
      <Form.Field
        name="error"
        title="With Error"
        component={Input}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        component={Input}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        horizontal={false}
        component={Input}
      />
    </FieldsWrapperExample>
  ))

storiesOf('Fields/Byte', module)
  .addDecorator((story, context) => withInfo(info)(story)(context))
  .add('Horizontal', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': 'ADADADAD',
        description: 'ADADADAD',
        warning: 'ADADADAD',
        error: 'ADADADAD',
        disabled: 'ADADADAD',
      }}
    >
      <Form.Field
        name="default"
        title="Default"
        type="byte"
        placeholder="default"
        min={4}
        max={4}
        component={Input}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        type="byte"
        placeholder="description"
        min={4}
        max={4}
        component={Input}
      />
      <Form.Field
        name="error"
        title="With Error"
        type="byte"
        placeholder="error"
        min={4}
        max={4}
        component={Input}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        type="byte"
        placeholder="warning"
        min={4}
        max={4}
        component={Input}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        placeholder="disabled"
        type="byte"
        min={4}
        max={4}
        component={Input}
      />
    </FieldsWrapperExample>
  ))
  .add('Vertical', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': 'ADADADAD',
        description: 'ADADADAD',
        warning: 'ADADADAD',
        error: 'ADADADAD',
        disabled: 'ADADADAD',
      }}
      horizontal={false}
    >
      <Form.Field
        name="default"
        title="Default"
        type="byte"
        placeholder="default"
        min={4}
        max={4}
        component={Input}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        type="byte"
        placeholder="description"
        min={4}
        max={4}
        component={Input}
      />
      <Form.Field
        name="error"
        title="With Error"
        type="byte"
        placeholder="error"
        min={4}
        max={4}
        component={Input}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        type="byte"
        placeholder="warning"
        min={4}
        max={4}
        component={Input}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        type="byte"
        placeholder="disabled"
        min={4}
        max={4}
        component={Input}
      />
    </FieldsWrapperExample>
  ))

storiesOf('Fields/TextArea', module)
  .addDecorator((story, context) => withInfo(info)(story)(context))
  .add('Horizontal', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': 'something...',
        description: 'something...',
        warning: 'something...',
        error: 'something...',
        disabled: 'something...',
      }}
    >
      <Form.Field
        name="default"
        title="Default"
        type="textarea"
        component={Input}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        type="textarea"
        component={Input}
      />
      <Form.Field
        name="error"
        title="With Error"
        type="textarea"
        component={Input}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        type="textarea"
        component={Input}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        type="textarea"
        component={Input}
      />
    </FieldsWrapperExample>
  ))
  .add('Vertical', () => (
    <FieldsWrapperExample
      initialValues={{
        'default': 'something...',
        description: 'something...',
        warning: 'something...',
        error: 'something...',
        disabled: 'something...',
      }}
      horizontal={false}
    >
      <Form.Field
        name="default"
        title="Default"
        type="textarea"
        component={Input}
      />
      <Form.Field
        name="description"
        title="With Description"
        description="A select field."
        type="textarea"
        component={Input}
      />
      <Form.Field
        name="error"
        title="With Error"
        type="textarea"
        component={Input}
      />
      <Form.Field
        name="warning"
        title="With Warning"
        warning="A select field."
        type="textarea"
        component={Input}
      />
      <Form.Field
        name="disabled"
        title="Disabled"
        disabled
        type="textarea"
        component={Input}
      />
    </FieldsWrapperExample>
  ))
