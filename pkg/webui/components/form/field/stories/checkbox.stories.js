// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

/* eslint-disable react/prop-types */

import React from 'react'

import Checkbox from '@ttn-lw/components/checkbox'
import Form from '@ttn-lw/components/form'

import FieldsWrapperExample from './shared'

export default {
  title: 'Fields/Checkbox',
}

export const Default = () => (
  <FieldsWrapperExample
    initialValues={{
      default: true,
      description: false,
      warning: false,
      error: false,
      disabled: false,
    }}
  >
    <Form.Field name="default" title="Default" component={Checkbox} />
    <Form.Field name="without-title" label="Without title" component={Checkbox} />
    <Form.Field
      name="description"
      title="With Description"
      description="A select field."
      component={Checkbox}
    />
    <Form.Field name="error" title="With Error" component={Checkbox} />
    <Form.Field
      name="warning"
      title="With Warning"
      warning="A select field."
      component={Checkbox}
    />
    <Form.Field name="disabled" title="Disabled" disabled component={Checkbox} />
  </FieldsWrapperExample>
)

export const HorizontalGroup = () => (
  <FieldsWrapperExample
    initialValues={{
      default: {
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
    <Form.Field name="default" title="Default" component={Checkbox.Group} horizontal>
      <Checkbox name="default1" label="Checkbox 1" />
      <Checkbox name="default2" label="Checkbox 2" />
      <Checkbox name="default3" label="Checkbox 3" />
    </Form.Field>
    <Form.Field
      name="description"
      title="With Description"
      description="A select field."
      component={Checkbox.Group}
      horizontal
    >
      <Checkbox name="description1" label="Checkbox 1" />
      <Checkbox name="description2" label="Checkbox 2" />
      <Checkbox name="description3" label="Checkbox 3" />
    </Form.Field>
    <Form.Field name="error" title="With Error" component={Checkbox.Group} horizontal>
      <Checkbox name="error1" label="Checkbox 1" />
      <Checkbox name="error2" label="Checkbox 2" />
      <Checkbox name="error3" label="Checkbox 3" />
    </Form.Field>
    <Form.Field
      name="warning"
      title="With Warning"
      warning="A select field."
      component={Checkbox.Group}
      horizontal
    >
      <Checkbox name="warning1" label="Checkbox 1" />
      <Checkbox name="warning2" label="Checkbox 2" />
      <Checkbox name="warning3" label="Checkbox 3" />
    </Form.Field>
    <Form.Field name="disabled" title="Disabled" disabled component={Checkbox.Group} horizontal>
      <Checkbox name="disabled1" label="Checkbox 1" />
      <Checkbox name="disabled2" label="Checkbox 2" />
      <Checkbox name="disabled3" label="Checkbox 3" />
    </Form.Field>
  </FieldsWrapperExample>
)

export const RowGroup = () => (
  <FieldsWrapperExample
    initialValues={{
      default: {
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
    <Form.Field name="default" title="Default" component={Checkbox.Group}>
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
    <Form.Field name="error" title="With Error" component={Checkbox.Group}>
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
    <Form.Field name="disabled" title="Disabled" disabled component={Checkbox.Group}>
      <Checkbox name="disabled1" label="Checkbox 1" />
      <Checkbox name="disabled2" label="Checkbox 2" />
      <Checkbox name="disabled3" label="Checkbox 3" />
    </Form.Field>
  </FieldsWrapperExample>
)
