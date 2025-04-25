// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback } from 'react'
import { defineMessages } from 'react-intl'
import PropTypes from 'prop-types'

import { GATEWAY } from '@console/constants/entities'

import Icon, { IconKey, IconPlus } from '@ttn-lw/components/icon'
import ModalButton from '@ttn-lw/components/button/modal-button'
import PortalledModal from '@ttn-lw/components/modal/portalled'

import Message from '@ttn-lw/lib/components/message'

import { ApiKeyCreateForm } from '@console/containers/api-key-form'
import { ApiKeyModalCreateForm } from '@console/containers/api-key-modal-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const m = defineMessages({
  modalTitle: 'Create a new API key for {entityName}',
})

const CreateApiKeyModal = props => {
  const handleCloseModal = useCallback(() => {
    props.setModalVisible(false)
  }, [props])
  const content = (
    <div className="flex-column gap-cs-xl w-full">
      <div className="d-flex al-center gap-cs-m">
        <Icon icon={IconKey} large />
        <Message
          content={m.modalTitle}
          values={{
            entityName: props.entityName || props.entityId,
          }}
          className="fs-l fw-bold"
        />
      </div>
      <hr className="w-full" />
      <ApiKeyModalCreateForm
        entityId={props.entityId}
        entity={GATEWAY}
        handleCancel={handleCloseModal}
      />
    </div>
  )

  return (
    <PortalledModal
      visible={props.modalVisible}
      className="p-cs-xl c-bg-neutral-extralight br-l"
      noControlBar
      noTitleLine
      children={content}
      onComplete={handleCloseModal}
    />
  )
}

CreateApiKeyModal.propTypes = {
  entityId: PropTypes.string.isRequired,
  entityName: PropTypes.string.isRequired,
  modalVisible: PropTypes.bool.isRequired,
  setModalVisible: PropTypes.func.isRequired,
}

export default CreateApiKeyModal
