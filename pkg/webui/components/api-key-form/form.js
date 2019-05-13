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
import bind from 'autobind-decorator'

import Form from '../form'
import PropTypes from '../../lib/prop-types'

@bind
class ApiKeyForm extends React.Component {
  state = {
    error: '',
  }

  async handleSubmit (values, { resetForm }) {
    const {
      onSubmit,
      onSubmitSuccess,
      onSubmitFailure,
    } = this.props

    await this.setState({ error: '' })

    try {
      const result = await onSubmit(values)

      resetForm(values)
      await onSubmitSuccess(result)
    } catch (error) {
      resetForm(values)

      await this.setState({ error })
      await onSubmitFailure(error)
    }
  }

  render () {
    const {
      children,
      formError,
      initialValues,
      validationSchema,
    } = this.props
    const { error } = this.state

    const displayError = error || formError || ''

    return (
      <Form
        error={displayError}
        onSubmit={this.handleSubmit}
        validationSchema={validationSchema}
        initialValues={initialValues}
      >
        {children}
      </Form>
    )
  }
}

ApiKeyForm.propTypes = {
  rights: PropTypes.arrayOf(PropTypes.string),
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func,
  onSubmitFailure: PropTypes.func,
  validationSchema: PropTypes.object.isRequired,
  horizontal: PropTypes.bool,
  initialValues: PropTypes.object,
  formError: PropTypes.object,
}

ApiKeyForm.defaultProps = {
  rights: [],
  horizontal: true,
  initialValues: {},
  formError: null,
  onSubmitSuccess: () => null,
  onSubmitFailure: () => null,
}

export default ApiKeyForm
