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

import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'

import { FieldsWrapperExample } from './shared'

export default {
  title: 'Fields/TextArea',
  component: Form.Field,
}

export const Default = () => (
  <FieldsWrapperExample
    initialValues={{
      default: 'something...',
      description: 'something...',
      warning: 'something...',
      error: 'something...',
      disabled: 'something...',
    }}
  >
    <Form.Field name="default" title="Default" type="textarea" component={Input} />
    <Form.Field
      name="description"
      title="With Description"
      description="A select field."
      type="textarea"
      component={Input}
    />
    <Form.Field name="error" title="With Error" type="textarea" component={Input} />
    <Form.Field
      name="warning"
      title="With Warning"
      warning="A select field."
      type="textarea"
      component={Input}
    />
    <Form.Field name="disabled" title="Disabled" disabled type="textarea" component={Input} />
  </FieldsWrapperExample>
)
