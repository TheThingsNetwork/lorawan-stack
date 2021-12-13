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
import { withInfo } from '@storybook/addon-info'

import UnitInput from '@ttn-lw/components/unit-input'
import Form from '@ttn-lw/components/form'

import { FieldsWrapperExample, info } from './shared'

export default {
  title: 'Fields/UnitInput',
  decorators: [withInfo(info)],
}

export const Default = () => (
  <FieldsWrapperExample
    initialValues={{
      default: '530ms',
      description: '530ms',
      warning: '530ms',
      error: '530ms',
      disabled: '530ms',
    }}
  >
    <Form.Field
      name="default"
      title="Default"
      units={[
        { label: 'miliseconds', value: 'ms' },
        { label: 'seconds', value: 's' },
        { label: 'minutes', value: 'm' },
        { label: 'hours', value: 'h' },
      ]}
      component={UnitInput}
    />
    <Form.Field
      name="description"
      title="Description"
      units={[
        { label: 'miliseconds', value: 'ms' },
        { label: 'seconds', value: 's' },
        { label: 'minutes', value: 'm' },
        { label: 'hours', value: 'h' },
      ]}
      component={UnitInput}
      description="The unit input"
    />
    <Form.Field
      name="warning"
      title="Warning"
      units={[
        { label: 'miliseconds', value: 'ms' },
        { label: 'seconds', value: 's' },
        { label: 'minutes', value: 'm' },
        { label: 'hours', value: 'h' },
      ]}
      component={UnitInput}
      warning="The unit input"
    />
    <Form.Field
      name="error"
      title="Error"
      units={[
        { label: 'miliseconds', value: 'ms' },
        { label: 'seconds', value: 's' },
        { label: 'minutes', value: 'm' },
        { label: 'hours', value: 'h' },
      ]}
      component={UnitInput}
    />
  </FieldsWrapperExample>
)
