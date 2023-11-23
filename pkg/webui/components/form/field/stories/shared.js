// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useEffect, useRef } from 'react'
import { action } from '@storybook/addon-actions'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'

import Form from '../..'

const handleSubmit = (data, { resetForm }) => {
  action('Submit')(data)
  setTimeout(() => resetForm({ values: data }), 1000)
}

const info = {
  inline: true,
  header: false,
  source: false,
  propTables: [Form.Field],
}

const errorSchema = Yup.string().test('error', 'Something went wrong.', () => false)
const validationSchema = Yup.object().shape({
  error: errorSchema,
})

const FieldsWrapperExample = props => {
  const formRef = useRef(null)
  const { initialValues, children } = props
  useEffect(() => {
    if (formRef.current) {
      formRef.current.setFieldError('error', 'Something went wrong.')
      formRef.current.setFieldTouched('error')
    }
  }, [])

  return (
    <Form
      onSubmit={handleSubmit}
      initialValues={initialValues}
      formikRef={formRef}
      validationSchema={validationSchema}
    >
      {children}
    </Form>
  )
}

FieldsWrapperExample.propTypes = {
  children: PropTypes.node.isRequired,
  initialValues: PropTypes.shape({}).isRequired,
}

export { info, FieldsWrapperExample }
