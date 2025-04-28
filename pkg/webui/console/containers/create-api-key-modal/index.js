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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { defineMessages } from 'react-intl'

import { GATEWAY } from '@console/constants/entities'

import Icon, {
  IconKey,
  IconDownload,
  IconCheck,
  IconCopyCheck,
  IconCopy,
} from '@ttn-lw/components/icon'
import PortalledModal from '@ttn-lw/components/modal/portalled'
import Notification from '@ttn-lw/components/notification'
import SafeInspector from '@ttn-lw/components/safe-inspector'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import { ApiKeyModalCreateForm } from '@console/containers/api-key-modal-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  firstStepTitle: 'Create a new API key for {entityName}',
  secondStepTitle: 'Your API key has been created successfully',
  secondStepNotification:
    'After closing this window, the value of the key secret will not be accessible anymore. Make sure to copy and store it in a safe place now.',
  download: 'Download',
})

const CreateApiKeyModal = props => {
  const [apiKey, setApiKey] = useState(null)
  const [copied, setCopied] = useState(false)

  const canCopy = navigator.clipboard && navigator.clipboard.writeText

  const _timer = useRef(null)

  useEffect(() => {
    if (!props.modalVisible) {
      setApiKey(null)
      clearTimeout(_timer.current)
    }
    return () => {
      clearTimeout(_timer.current)
    }
  }, [props.modalVisible])

  const handleCloseModal = useCallback(() => {
    props.setModalVisible(false)
  }, [props])

  const handleCopyClick = useCallback(() => {
    if (copied) {
      return
    }
    if (canCopy) {
      navigator.clipboard.writeText(apiKey)
      setCopied(true)

      _timer.current = setTimeout(() => {
        setCopied(false)
      }, 3000)
    }
  }, [copied, canCopy, apiKey])

  const firstStep = (
    <div className="flex-column gap-cs-xl w-full">
      <div className="d-flex al-center gap-cs-m">
        <Icon icon={IconKey} large />
        <Message
          content={m.firstStepTitle}
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
        rights={props.rights}
        setApiKey={setApiKey}
      />
    </div>
  )

  const secondStep = (
    <div className="flex-column gap-cs-xl w-full">
      <Message content={m.secondStepTitle} className="fs-l fw-bold" />
      <Notification content={m.secondStepNotification} warning small />
      <hr className="w-full" />
      <div>
        <Message component="p" className="mt-0 mb-cs-xxs fw-bold" content={sharedMessages.apiKey} />
        <div className="d-flex gap-cs-xs w-full">
          <SafeInspector
            className="overflow-x-hidden w-full"
            data={apiKey}
            isBytes={false}
            disableResize
            noCopy={canCopy}
          />
          {canCopy && (
            <Button
              message={!copied ? sharedMessages.copy : sharedMessages.copiedToClipboard}
              className={copied && 'c-icon-success-normal'}
              secondary
              onClick={handleCopyClick}
              icon={copied ? IconCopyCheck : IconCopy}
            />
          )}
        </div>
      </div>
      <div className="d-flex gap-cs-xs j-end">
        <Button message={m.download} secondary onClick={handleCloseModal} icon={IconDownload} />
        <Button
          message={sharedMessages.copiedTheKey}
          primary
          onClick={handleCloseModal}
          icon={IconCheck}
        />
      </div>
    </div>
  )

  return (
    <PortalledModal
      visible={props.modalVisible}
      className="p-cs-xl c-bg-neutral-extralight br-l"
      noControlBar
      noTitleLine
      children={!Boolean(apiKey) ? firstStep : secondStep}
      onComplete={handleCloseModal}
    />
  )
}

CreateApiKeyModal.propTypes = {
  entityId: PropTypes.string.isRequired,
  entityName: PropTypes.string,
  modalVisible: PropTypes.bool.isRequired,
  rights: PropTypes.rightsRadioItems.isRequired,
  setModalVisible: PropTypes.func.isRequired,
}

CreateApiKeyModal.defaultProps = {
  entityName: undefined,
}

export default CreateApiKeyModal
