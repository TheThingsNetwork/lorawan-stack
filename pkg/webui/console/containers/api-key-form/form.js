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

import React, { useState, useCallback } from 'react'

import Form from '@ttn-lw/components/form'

import PropTypes from '@ttn-lw/lib/prop-types'

const ApiKeyForm = ({
  children,
  formError,
  initialValues,
  onSubmit,
  onSubmitFailure,
  onSubmitSuccess,
  validationSchema,
}) => {
  const [error, setError] = useState('')

  const handleSubmit = useCallback(
    async (values, { resetForm }) => {
      setError('')

      try {
        const result = await onSubmit(values)

        resetForm({ values })
        await onSubmitSuccess(result)
      } catch (error) {
        resetForm({ values })

        setError(error)
        await onSubmitFailure(error)
      }
    },
    [onSubmit, onSubmitSuccess, onSubmitFailure],
  )

  const displayError = error || formError || ''

  return (
    <Form
      error={displayError}
      onSubmit={handleSubmit}
      validationSchema={validationSchema}
      initialValues={initialValues}
    >
      {children}
    </Form>
  )
}

ApiKeyForm.propTypes = {
  children: PropTypes.node,
  formError: PropTypes.error,
  initialValues: PropTypes.shape({}),
  onSubmit: PropTypes.func.isRequired,
  onSubmitFailure: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
  validationSchema: PropTypes.shape({}).isRequired,
}

ApiKeyForm.defaultProps = {
  children: undefined,
  initialValues: {},
  formError: null,
}

export default ApiKeyForm
