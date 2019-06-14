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
import { defineMessages } from 'react-intl'

import diff from '../../lib/diff'
import PropTypes from '../../lib/prop-types'
import sharedMessages from '../../lib/shared-messages'
import SubmitBar from '../submit-bar'
import ModalButton from '../button/modal-button'
import toast from '../toast'
import Message from '../../lib/components/message'
import FormField from '../form/field'
import FormSubmit from '../form/submit'
import SubmitButton from '../submit-button'
import Input from '../input'
import RightsGroup from '../../console/components/rights-group'
import ApiKeyForm from './form'
import validationSchema from './validation-schema'

const m = defineMessages({
  deleteKey: 'Delete Key',
  modalWarning:
    'Are you sure you want to delete the {keyName} API Key? Deleting an API Key cannot be undone!',
  updateSuccess: 'Successfully updated API Key',
  deleteSuccess: 'Successfully deleted API Key',
})

@bind
class EditForm extends React.Component {

  state = {
    error: null,
  }

  async handleEdit (values) {
    const { name, rights } = values
    const { apiKey, onEdit } = this.props

    const changed = diff({ name: apiKey.name }, { name })
    changed.rights = Object.keys(rights).filter(r => rights[r])

    return await onEdit(changed)
  }

  async handleEditSuccess (key) {
    const { onEditSuccess } = this.props

    toast({
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
    await onEditSuccess(key)
  }

  async handleDelete () {
    const {
      onDelete,
      apiKey,
    } = this.props

    await this.setState({ error: null })

    try {
      await onDelete(apiKey.id)
      await this.handleDeleteSuccess(apiKey.id)
    } catch (error) {
      await this.handleDeleteFailure(error)
    }
  }

  async handleDeleteSuccess (id) {
    const { onDeleteSuccess } = this.props

    toast({
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })
    await onDeleteSuccess(id)
  }

  async handleDeleteFailure (error) {
    const { onDeleteFailure } = this.props

    await this.setState({ error })
    await onDeleteFailure(error)
  }

  render () {
    const {
      rights,
      apiKey,
      onEditFailure,
      universalRights,
    } = this.props
    const { error } = this.state

    const hasUniversalRight = universalRights.length
      ? universalRights.some(universalRight => apiKey.rights.includes(universalRight))
      : false

    const rightsValues = rights.reduce(function (acc, right) {
      acc[right] = hasUniversalRight || apiKey.rights.includes(right)

      return acc
    }, {})
    const initialValues = {
      id: apiKey.id,
      name: apiKey.name,
      rights: rightsValues,
    }

    return (
      <ApiKeyForm
        formError={error}
        initialValues={initialValues}
        validationSchema={validationSchema}
        onSubmit={this.handleEdit}
        onSubmitSuccess={this.handleEditSuccess}
        onSubmitFailure={onEditFailure}
      >
        <Message
          component="h4"
          content={sharedMessages.generalInformation}
        />
        <FormField
          title={sharedMessages.keyId}
          name="id"
          required
          valid
          disabled
          component={Input}
        />
        <FormField
          title={sharedMessages.name}
          name="name"
          component={Input}
        />
        <FormField
          name="rights"
          title={sharedMessages.rights}
          required
          component={RightsGroup}
          rights={rights}
          universalRight={universalRights[0]}
        />
        <SubmitBar>
          <FormSubmit
            component={SubmitButton}
            message={sharedMessages.saveChanges}
          />
          <ModalButton
            type="button"
            icon="delete"
            danger
            naked
            message={m.deleteKey}
            modalData={{
              message: {
                values: { keyName: apiKey.name ? `"${apiKey.name}"` : '' },
                ...m.modalWarning,
              },
            }}
            onApprove={this.handleDelete}
          />
        </SubmitBar>
      </ApiKeyForm>
    )
  }
}

EditForm.propTypes = {
  /** The API key to be edited */
  apiKey: PropTypes.shape({
    id: PropTypes.string.isRequired,
    rights: PropTypes.arrayOf(PropTypes.string).isRequired,
    name: PropTypes.string,
  }),
  /** The list of rights */
  rights: PropTypes.arrayOf(PropTypes.string),
  /**
   * The rights that imply all other rights, e.g. 'RIGHT_APPLICATION_ALL', 'RIGHT_ALL'
   */
  universalRights: PropTypes.arrayOf(PropTypes.string),
  /**
   * Called on form submission.
   * Receives the updated key object as an argument.
   */
  onEdit: PropTypes.func.isRequired,
  /**
   * Called after successful update of the API key.
   * Receives the key object as an argument.
   */
  onEditSuccess: PropTypes.func,
  /**
   * Called after unsuccessful update of the API key.
   * Receives the error object as an argument.
   */
  onEditFailure: PropTypes.func,
  /**
   * Called on key deletion.
   * Receives the identifier of the API key as an argument.
   */
  onDelete: PropTypes.func.isRequired,
  /**
   * Called after successful deletion of the API key.
   * Receives the identifier of the API key as an argument.
   */
  onDeleteSuccess: PropTypes.func,
  /**
   * Called after unsuccessful deletion of the API key.
   * Receives the error object as an argument.
   */
  onDeleteFailure: PropTypes.func,
}

EditForm.defaultProps = {
  rights: [],
  onEditSuccess: () => null,
  onEditFailure: () => null,
  onDeleteSuccess: () => null,
  onDeleteFailure: () => null,
  universalRights: [],
}

export default EditForm
