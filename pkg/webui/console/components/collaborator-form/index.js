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

import React, { Component } from 'react'
import * as Yup from 'yup'
import bind from 'autobind-decorator'

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import { id as collaboratorIdRegexp } from '../../lib/regexp'

import Form from '../../../components/form'
import Input from '../../../components/input'
import Checkbox from '../../../components/checkbox'
import Select from '../../../components/select'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import Message from '../../../lib/components/message'
import toast from '../../../components/toast'
import ModalButton from '../../../components/button/modal-button'

import style from './collaborator-form.styl'

const validationSchema = Yup.object().shape({
  collaborator_id: Yup.string()
    .matches(collaboratorIdRegexp, sharedMessages.validateAlphanum)
    .required(sharedMessages.validateRequired),
  collaborator_type: Yup.string()
    .required(sharedMessages.validateRequired),
  rights: Yup.object().test(
    'rights',
    sharedMessages.validateRights,
    values => Object.values(values).reduce((acc, curr) => acc || curr, false)
  ),
})

@bind
export default class CollaboratorForm extends Component {

  static defaultProps = {
    onSubmitSuccess: () => null,
    onSubmitFailure: () => null,
    onDelete: () => null,
    onDeleteSuccess: () => null,
    oneleteFailure: () => null,
    rights: [],
  }

  static propTypes = {
    onSubmit: PropTypes.func.isRequired,
    onSubmitSuccess: PropTypes.func,
    onSubmitFailure: PropTypes.func,
    onDelete: PropTypes.func,
    onDeleteSuccess: PropTypes.func,
    onDeleteFailure: PropTypes.func,
    rights: PropTypes.array,
    initialFormValues: PropTypes.object,
    error: PropTypes.error,
  }

  state = {
    error: '',
  }

  async handleSubmit (values, { resetForm, setSubmitting }) {
    const { collaborator_id, collaborator_type, rights } = values
    const { onSubmit, onSubmitSuccess, onSubmitFailure } = this.props

    const collaborator_ids = {
      [`${collaborator_type}_ids`]: {
        [`${collaborator_type}_id`]: collaborator_id,
      },
    }
    const collaborator = {
      ids: collaborator_ids,
      rights: Object.keys(rights).filter(r => rights[r]),
    }

    await this.setState({ error: '' })

    try {
      await onSubmit(collaborator)
      resetForm(values)
      onSubmitSuccess()
    } catch (error) {
      setSubmitting(false)
      this.setState({ error })
      onSubmitFailure(error)
    }
  }

  async handleDelete () {
    const { collaborator, onDelete, onDeleteSuccess } = this.props
    const collaborator_type = collaborator.isUser ? 'user' : 'organization'

    const collaborator_ids = {
      [`${collaborator_type}_ids`]: {
        [`${collaborator_type}_id`]: collaborator.id,
      },
    }
    const updatedCollaborator = {
      ids: collaborator_ids,
    }

    try {
      await onDelete(updatedCollaborator)
      toast({
        message: sharedMessages.collaboratorDeleteSuccess,
        type: toast.types.SUCCESS,
      })
      onDeleteSuccess()
    } catch (error) {
      await this.setState({ error })
    }
  }

  computeInitialValues () {
    const { collaborator, rights, universalRightLiterals } = this.props

    if (!collaborator) {
      return {
        collaborator_id: '',
        collaborator_type: 'user',
        rights: {},
      }
    }

    const hasUniversalRights = universalRightLiterals.reduce(
      (acc, curr) => acc || collaborator.rights.includes(curr), false)
    const rightsValues = rights.reduce(
      function (acc, right) {
        acc[right] = hasUniversalRights || collaborator.rights.includes(right)

        return acc
      },
      {}
    )

    return {
      collaborator_id: collaborator.id,
      collaborator_type: collaborator.isUser ? 'user' : 'organization',
      rights: { ...rightsValues },
    }
  }

  render () {
    const {
      collaborator,
      rights,
      error: passedError,
      update,
    } = this.props
    const { error: submitError } = this.state

    const error = passedError || submitError

    const rightsItems = rights.map(
      right => (
        <Checkbox
          className={style.rightLabel}
          key={right}
          name={right}
          label={{ id: `enum:${right}` }}
        />
      )
    )

    return (
      <Form
        horizontal
        error={error}
        onSubmit={this.handleSubmit}
        initialValues={this.computeInitialValues()}
        validationSchema={validationSchema}
      >
        <Message
          component="h4"
          content={sharedMessages.generalInformation}
        />
        <Form.Field
          name="collaborator_id"
          component={Input}
          title={sharedMessages.collaboratorId}
          required
          autoFocus={!update}
        />
        <Form.Field
          name="collaborator_type"
          component={Select}
          title={sharedMessages.type}
          required
          disabled={update}
          options={[
            { value: 'user', label: sharedMessages.user },
            { value: 'organization', label: sharedMessages.organization },
          ]}
        />
        <Form.Field
          name="rights"
          title={sharedMessages.rights}
          required
          component={Checkbox.Group}
        >
          {rightsItems}
        </Form.Field>
        <SubmitBar>
          <Form.Submit
            component={SubmitButton}
            message={update
              ? sharedMessages.saveChanges
              : sharedMessages.collaboratorAdd
            }
          />
          { update && (
            <ModalButton
              type="button"
              icon="delete"
              danger
              naked
              message={sharedMessages.removeCollaborator}
              modalData={{
                message: {
                  values: { collaboratorId: collaborator.id },
                  ...sharedMessages.collaboratorModalWarning,
                },
              }}
              onApprove={this.handleDelete}
            />
          )}
        </SubmitBar>
      </Form>
    )
  }
}

