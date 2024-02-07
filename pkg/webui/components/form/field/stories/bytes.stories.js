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

import FieldsWrapperExample from './shared'

export default {
  title: 'Fields/Byte',
}

export const Default = () => (
  <FieldsWrapperExample
    initialValues={{
      default: 'ADADADAD',
      'xxs-size': 'ADAD',
      'xs-size': 'ADADADAD',
      's-size': 'ADADADADADADADAD',
      'm-size': 'ADADADADADADADADADADADADADADAD',
      'l-size': 'ADADADADADADADADADADADADADADADADADADADAD',
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
      name="xxs-size"
      title="XXS Size"
      type="byte"
      placeholder="default"
      min={2}
      max={2}
      component={Input}
      inputWidth="xxs"
    />
    <Form.Field
      name="xs-size"
      title="XS Size"
      type="byte"
      placeholder="default"
      min={4}
      max={4}
      component={Input}
      inputWidth="xs"
    />
    <Form.Field
      name="s-size"
      title="S Size"
      type="byte"
      placeholder="default"
      min={8}
      max={8}
      component={Input}
      inputWidth="s"
    />
    <Form.Field
      name="m-size"
      title="M Size"
      type="byte"
      placeholder="default"
      min={15}
      max={15}
      component={Input}
      inputWidth="m"
    />
    <Form.Field
      name="l-size"
      title="L Size"
      type="byte"
      placeholder="default"
      min={20}
      max={20}
      component={Input}
      inputWidth="l"
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
)
