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
import Radio from '../../../components/radio-button'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import Message from '../../../lib/components/message'
import toast from '../../../components/toast'
import ModalButton from '../../../components/button/modal-button'
import RightsGroup from '../../components/rights-group'

const validationSchema = Yup.object().shape({
  collaborator_id: Yup.string()
    .matches(collaboratorIdRegexp, sharedMessages.validateAlphanum)
    .required(sharedMessages.validateRequired),
  collaborator_type: Yup.string().required(sharedMessages.validateRequired),
  rights: Yup.array().min(1, sharedMessages.validateRights),
})

@bind
export default class CollaboratorForm extends Component {
  static defaultProps = {
    onSubmitSuccess: () => null,
    onSubmitFailure: () => null,
    onDelete: () => null,
    onDeleteSuccess: () => null,
    onDeleteFailure: () => null,
    pseudoRights: [],
    error: '',
    collaborator: undefined,
    update: false,
  }

  static propTypes = {
    collaborator: PropTypes.collaborator,
    error: PropTypes.error,
    onDelete: PropTypes.func,
    onDeleteFailure: PropTypes.func,
    onDeleteSuccess: PropTypes.func,
    onSubmit: PropTypes.func.isRequired,
    onSubmitFailure: PropTypes.func,
    onSubmitSuccess: PropTypes.func,
    rights: PropTypes.rights.isRequired,
    pseudoRights: PropTypes.rights,
    update: PropTypes.bool,
  }

  state = {
    error: '',
  }

  async handleSubmit(values, { resetForm, setSubmitting }) {
    const { collaborator_id, collaborator_type, rights } = values
    const { onSubmit, onSubmitSuccess, onSubmitFailure } = this.props

    const collaborator_ids = {
      [`${collaborator_type}_ids`]: {
        [`${collaborator_type}_id`]: collaborator_id,
      },
    }

    const collaborator = {
      ids: collaborator_ids,
      rights,
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

  async handleDelete() {
    const { collaborator, onDelete, onDeleteSuccess, onDeleteFailure } = this.props
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
      this.setState({ error })
      onDeleteFailure(error)
    }
  }

  computeInitialValues() {
    const { collaborator, pseudoRights } = this.props

    if (!collaborator) {
      return {
        collaborator_id: '',
        collaborator_type: 'user',
        rights: [...pseudoRights],
      }
    }

    return {
      collaborator_id: collaborator.id,
      collaborator_type: collaborator.isUser ? 'user' : 'organization',
      rights: [...collaborator.rights],
    }
  }

  render() {
    const { collaborator, rights, pseudoRights, error: passedError, update } = this.props

    const { error: submitError } = this.state

    const error = passedError || submitError

    return (
      <Form
        horizontal
        error={error}
        onSubmit={this.handleSubmit}
        initialValues={this.computeInitialValues()}
        validationSchema={validationSchema}
      >
        <Message component="h4" content={sharedMessages.generalInformation} />
        <Form.Field
          name="collaborator_id"
          component={Input}
          title={sharedMessages.collaboratorId}
          required
          autoFocus={!update}
          disabled={update}
        />
        <Form.Field
          name="collaborator_type"
          title={sharedMessages.type}
          component={Radio.Group}
          horizontal={false}
          disabled={update}
          required
        >
          <Radio label={sharedMessages.user} value="user" />
          <Radio label={sharedMessages.organization} value="organization" />
        </Form.Field>
        <Form.Field
          name="rights"
          title={sharedMessages.rights}
          required
          component={RightsGroup}
          rights={rights}
          pseudoRight={pseudoRights[0]}
          entityTypeMessage={sharedMessages.collaborator}
        />
        <SubmitBar>
          <Form.Submit
            component={SubmitButton}
            message={update ? sharedMessages.saveChanges : sharedMessages.addCollaborator}
          />
          {update && (
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
