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

import ApiKeyModal from '../api-key-modal'
import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'
import SubmitBar from '../../../components/submit-bar'
import FormField from '../../../components/form/field'
import FormSubmit from '../../../components/form/submit'
import SubmitButton from '../../../components/submit-button'
import Input from '../../../components/input'
import Message from '../../../lib/components/message'
import RightsGroup from '../../../console/components/rights-group'
import ApiKeyForm from './form'
import validationSchema from './validation-schema'

@bind
class CreateForm extends React.Component {

  state = {
    modal: null,
  }

  async handleModalApprove () {
    const { onCreateSuccess } = this.props
    const { key } = this.state

    await this.setState({ modal: null })
    await onCreateSuccess(key)
  }

  async handleCreate (values) {
    const { onCreate } = this.props

    return await onCreate(values)
  }

  async handleCreateSuccess (key) {
    await this.setState({
      modal: {
        secret: key.key,
        rights: key.rights,
        onComplete: this.handleModalApprove,
        approval: false,
      },
      key,
    })
  }

  render () {
    const {
      rights,
      onCreateFailure,
      universalRights,
    } = this.props
    const { modal } = this.state

    const modalProps = modal ? modal : {}
    const modalVisible = Boolean(modal)
    const initialValues = {
      name: '',
      rights: [],
    }

    return (
      <React.Fragment>
        <ApiKeyModal
          {...modalProps}
          visible={modalVisible}
          approval={false}
        />
        <ApiKeyForm
          rights={rights}
          onSubmit={this.handleCreate}
          onSubmitSuccess={this.handleCreateSuccess}
          onSubmitFailure={onCreateFailure}
          validationSchema={validationSchema}
          initialValues={initialValues}
        >
          <Message
            component="h4"
            content={sharedMessages.generalInformation}
          />
          <FormField
            title={sharedMessages.name}
            name="name"
            autoFocus
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
              message={sharedMessages.createApiKey}
            />
          </SubmitBar>
        </ApiKeyForm>
      </React.Fragment>
    )
  }
}

CreateForm.propTypes = {
  /** The list of rights */
  rights: PropTypes.arrayOf(PropTypes.string),
  /**
   * The rights that imply all other rights, e.g. 'RIGHT_APPLICATION_ALL', 'RIGHT_ALL'
   */
  universalRights: PropTypes.arrayOf(PropTypes.string),
  /**
   * Called after successful creation of the API key.
   * Receives the key object as an argument.
   */
  onCreateSuccess: PropTypes.func,
  /**
   * Called after unsuccessful creation of the API key.
   * Receives the error object as an argument.
   */
  onCreateFailure: PropTypes.func,
  /**
   * Called on form submission.
   * Receives the key object as an argument.
   */
  onCreate: PropTypes.func.isRequired,
}

CreateForm.defaultProps = {
  rights: [],
  onCreateSuccess: () => null,
  onCreateFailure: () => null,
  universalRights: [],
}

export default CreateForm
