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

import FileInput from '@ttn-lw/components/file-input'
import Form from '@ttn-lw/components/form'

import { FieldsWrapperExample } from './shared'

export default {
  title: 'Fields/FileInput',
  component: Form.Field,
}

export const Default = () => (
  <FieldsWrapperExample
    initialValues={{
      default: '',
      withValue: 'base64-value-goes-here',
      error: '',
    }}
  >
    <Form.Field name="default" title="Default" component={FileInput} />
    <Form.Field
      name="description"
      title="With Description"
      description="A file input field."
      component={FileInput}
    />
    <Form.Field name="withValue" title="With initially attached file" component={FileInput} />
    <Form.Field name="error" title="With error" component={FileInput} />
    <Form.Field
      name="warning"
      title="With warning"
      component={FileInput}
      warning="A file input field."
    />
    <Form.Field name="disabled" title="Disabled" component={FileInput} disabled />
  </FieldsWrapperExample>
)
