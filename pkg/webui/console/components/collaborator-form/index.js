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
import { defineMessages } from 'react-intl'

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import { id as collaboratorIdRegexp } from '../../lib/regexp'
import { RIGHT_ALL } from '../../lib/rights'

import Form from '../../../components/form'
import Input from '../../../components/input'
import Select from '../../../components/select'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import Message from '../../../lib/components/message'
import toast from '../../../components/toast'
import ModalButton from '../../../components/button/modal-button'
import RightsGroup from '../../components/rights-group'
import Notification from '../../../components/notification'

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

const m = defineMessages({
  cannotModifyRightAll:
    'This user possesses universal admin rights that cannot be modified.',
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
    universalRights: [],
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
    universalRights: PropTypes.array,
  }

  state = {
    error: '',
  }

  async handleSubmit (values, { resetForm, setSubmitting }) {
    const { collaborator_id, collaborator_type, rights: rightMap } = values
    const { onSubmit, onSubmitSuccess, onSubmitFailure } = this.props

    const collaborator_ids = {
      [`${collaborator_type}_ids`]: {
        [`${collaborator_type}_id`]: collaborator_id,
      },
    }

    // Prefer universal wildcard rights over individual rights, if set
    const rights = Object.keys(rightMap).filter(r => rightMap[r])
    const universalRights = rights.filter(r => r.endsWith('_ALL'))
    const computedRights = universalRights.length
      ? universalRights
      : rights

    const collaborator = {
      ids: collaborator_ids,
      rights: computedRights,
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
      this.setState({ error })
    }
  }

  computeInitialValues () {
    const { collaborator, rights } = this.props

    if (!collaborator) {
      return {
        collaborator_id: '',
        collaborator_type: 'user',
        rights: {},
      }
    }

    // Translate universal rights back to individual values
    const hasRightAll = collaborator.rights.includes(RIGHT_ALL)
    const hasUniversalRight = collaborator.rights.some(r => r.endsWith('_ALL'))

    const rightsValues = rights.reduce(
      function (acc, right) {
        acc[right] = hasUniversalRight || collaborator.rights.includes(right)

        return acc
      },
      { [RIGHT_ALL]: hasRightAll }
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
      universalRights,
    } = this.props

    const { error: submitError } = this.state

    const error = passedError || submitError

    const hasRightAll = Boolean(collaborator && collaborator.rights.includes(RIGHT_ALL))

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
        { hasRightAll && <Notification small info={m.cannotModifyRightAll} />}
        <Form.Field
          name="rights"
          title={sharedMessages.rights}
          required
          strict
          component={RightsGroup}
          rights={rights}
          universalRight={universalRights[0]}
          disabled={hasRightAll}
        />
        <SubmitBar>
          <Form.Submit
            component={SubmitButton}
            message={update
              ? sharedMessages.saveChanges
              : sharedMessages.collaboratorAdd
            }
          />
          { update && !hasRightAll && (
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
