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

import SubmitBar from '@ttn-lw/components/submit-bar'
import ModalButton from '@ttn-lw/components/button/modal-button'
import toast from '@ttn-lw/components/toast'
import FormField from '@ttn-lw/components/form/field'
import FormSubmit from '@ttn-lw/components/form/submit'
import SubmitButton from '@ttn-lw/components/submit-button'
import Input from '@ttn-lw/components/input'
import RightsGroup from '@ttn-lw/components/rights-group'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import ApiKeyForm from './form'
import validationSchema from './validation-schema'
import { encodeExpiryDate, decodeExpiryDate } from './utils'

const m = defineMessages({
  deleteKey: 'Delete key',
  modalWarning:
    'Are you sure you want to delete the {keyName} API key? Deleting an API key cannot be undone.',
  updateSuccess: 'API key updated',
  deleteSuccess: 'API key deleted',
})

class EditForm extends React.Component {
  static propTypes = {
    /** The API key to be edited. */
    apiKey: PropTypes.apiKey,
    /**
     * Called on key deletion. Receives the identifier of the API key as an
     * argument.
     */
    onDelete: PropTypes.func.isRequired,
    /**
     * Called after unsuccessful deletion of the API key. Receives the error
     * object as an argument.
     */
    onDeleteFailure: PropTypes.func,
    /**
     * Called after successful deletion of the API key. Receives the identifier
     * of the API key as an argument.
     */
    onDeleteSuccess: PropTypes.func.isRequired,
    /**
     * Called on form submission. Receives the updated key object as an
     * argument.
     */
    onEdit: PropTypes.func.isRequired,
    /**
     * Called after unsuccessful update of the API key. Receives the error
     * object as an argument.
     */
    onEditFailure: PropTypes.func,
    /**
     * Called after successful update of the API key. Receives the key object as
     * an argument.
     */
    onEditSuccess: PropTypes.func,
    /**
     * The rights that imply all other rights, e.g. 'RIGHT_APPLICATION_ALL',
     * 'RIGHT_ALL'.
     */
    pseudoRights: PropTypes.arrayOf(PropTypes.string),
    /** The list of rights. */
    rights: PropTypes.arrayOf(PropTypes.string),
  }

  state = {
    error: null,
  }

  static defaultProps = {
    apiKey: undefined,
    rights: [],
    onEditFailure: () => null,
    onEditSuccess: () => null,
    onDeleteFailure: () => null,
    pseudoRights: [],
  }

  @bind
  async handleEdit(values) {
    const castedValues = validationSchema.cast(values)
    const { onEdit } = this.props

    return await onEdit(castedValues)
  }

  @bind
  async handleEditSuccess(key) {
    const { onEditSuccess } = this.props

    toast({
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
    await onEditSuccess(key)
  }

  @bind
  async handleDelete() {
    const { onDelete, apiKey } = this.props

    await this.setState({ error: null })

    try {
      await onDelete(apiKey.id)
      await this.handleDeleteSuccess(apiKey.id)
    } catch (error) {
      await this.handleDeleteFailure(error)
    }
  }

  @bind
  async handleDeleteSuccess(id) {
    const { onDeleteSuccess } = this.props

    toast({
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })
    await onDeleteSuccess(id)
  }

  @bind
  async handleDeleteFailure(error) {
    const { onDeleteFailure } = this.props

    await this.setState({ error })
    await onDeleteFailure(error)
  }

  render() {
    const { rights, apiKey, onEditFailure, pseudoRights } = this.props
    const { error } = this.state

    const initialValues = {
      id: apiKey.id,
      name: apiKey.name,
      rights: apiKey.rights,
      expires_at: apiKey.expires_at,
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
          placeholder={sharedMessages.apiKeyNamePlaceholder}
          name="name"
          component={Input}
        />
        <FormField
          title={'Expiry date'}
          name="expires_at"
          type="date"
          decode={decodeExpiryDate}
          encode={encodeExpiryDate}
          component={Input}
        />
        <FormField
          name="rights"
          title={sharedMessages.rights}
          required
          component={RightsGroup}
          rights={rights}
          pseudoRight={pseudoRights}
          entityTypeMessage={sharedMessages.apiKey}
        />
        <SubmitBar>
          <FormSubmit component={SubmitButton} message={sharedMessages.saveChanges} />
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

export default EditForm
