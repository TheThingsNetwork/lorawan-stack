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

import React, { useCallback } from 'react'
import { defineMessages } from 'react-intl'
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import Notification from '@ttn-lw/components/notification'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import humanFileSize from '@ttn-lw/lib/human-file-size'

import ButtonGroup from '../button/group'

import style from './file-input.styl'

const m = defineMessages({
  selectAFile: 'Select a file…',
  changeFile: 'Change file…',
  noFileSelected: 'No file selected',
  fileProvided: 'A file has been provided',
  tooBig: 'The selected file is too large',
})

const defaultDataTransform = content => content.replace(/^.*;base64,/, '')

const StatusMessage = props => {
  const { value, providedMessage, mayRemove, filename, error, handleRemoveClick } = props
  const hasInitialValue = value && !filename
  const hasError = Boolean(error)

  if (hasError) {
    return (
      <React.Fragment>
        <Icon className={style.errorIcon} icon="error" />
        <Message className={style.error} content={error} />
      </React.Fragment>
    )
  } else if (hasInitialValue || Boolean(filename)) {
    return (
      <React.Fragment>
        {hasInitialValue ? <Message content={providedMessage} /> : filename}
        {mayRemove && (
          <Button
            className="ml-cs-s"
            message={m.remove}
            onClick={handleRemoveClick}
            type="button"
            icon="delete"
            danger
            naked
          />
        )}
      </React.Fragment>
    )
  }

  return <Message className="tc-subtle-gray" content={m.noFileSelected} />
}

StatusMessage.propTypes = {
  error: PropTypes.message,
  filename: PropTypes.string,
  handleRemoveClick: PropTypes.func.isRequired,
  mayRemove: PropTypes.bool.isRequired,
  providedMessage: PropTypes.message.isRequired,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.shape({})]),
}

StatusMessage.defaultProps = {
  error: undefined,
  filename: undefined,
  value: undefined,
}

const FileInput = props => {
  const {
    onChange,
    dataTransform,
    image,
    maxSize,
    warningSize,
    message,
    changeMessage,
    name,
    id,
    accept,
    value,
    disabled,
    imageClassName,
    largeFileWarningMessage,
    center,
    providedMessage,
    mayRemove,
  } = props

  const [filename, setFilename] = React.useState('')
  const [isLarger, setIsLarger] = React.useState(false)
  const [error, setError] = React.useState(undefined)
  const fileInputRef = React.useRef()
  const imageRef = React.useRef()
  const reader = new FileReader()

  const handleFileRead = useCallback(
    event => {
      const { result: content } = event.target

      if (image && Boolean(imageRef.current)) {
        imageRef.current.style.display = 'block'
      }

      const data = dataTransform(content)
      onChange(data, true)
    },
    [dataTransform, image, onChange],
  )

  reader.onload = handleFileRead

  const handleChange = useCallback(
    event => {
      const { files } = event.target

      if (files && files[0]) {
        if (files[0].size >= maxSize) {
          setError(m.tooBig)
        } else {
          reader.readAsDataURL(files[0])
          setFilename(files[0].name)
          setError(undefined)
          setIsLarger(files[0].size >= warningSize)
        }
      }
    },
    [maxSize, warningSize, reader],
  )

  const handleChooseClick = useCallback(() => {
    fileInputRef.current.click()
  }, [])

  const handleRemoveClick = useCallback(() => {
    fileInputRef.current.value = null
    setFilename('')
    setError(undefined)
    setIsLarger(false)
    onChange(dataTransform(''), true)
  }, [dataTransform, onChange])

  const handleImageError = useCallback(error => {
    error.target.style.display = 'none'
  }, [])

  const warningThreshold = humanFileSize(warningSize)

  return (
    <div className={classnames(style.container, { [style.center]: center })}>
      {isLarger && (
        <Notification
          className={style.notification}
          content={largeFileWarningMessage}
          messageValues={{ warningThreshold }}
          small
          warning
        />
      )}
      <div>
        {image && Boolean(value) && (
          <img
            className={classnames(style.image, imageClassName)}
            alt="Current image"
            src={value}
            onError={handleImageError}
            ref={imageRef}
          />
        )}
        <ButtonGroup
          align={center ? 'center' : 'start'}
          className={classnames({ [style.buttonGroupCenter]: center })}
        >
          <Button
            type="button"
            aria-controls="fileupload"
            onClick={handleChooseClick}
            disabled={disabled}
            message={!value ? message : changeMessage}
            icon="attachment"
            className="mr-cs-s"
          />
          <StatusMessage
            value={value}
            providedMessage={providedMessage}
            mayRemove={mayRemove}
            filename={filename}
            error={error}
            handleRemoveClick={handleRemoveClick}
          />
        </ButtonGroup>
        <input
          name={name}
          id={id}
          className={style.input}
          type="file"
          onChange={handleChange}
          ref={fileInputRef}
          accept={accept}
          disabled={disabled}
          tabIndex="-1"
        />
      </div>
    </div>
  )
}

FileInput.propTypes = {
  accept: PropTypes.oneOfType([PropTypes.string, PropTypes.array]),
  center: PropTypes.bool,
  changeMessage: PropTypes.message,
  /** `dataTransform` is a marshaler used to transform the raw field value into
   * a value matching the field schema. */
  dataTransform: PropTypes.func,
  disabled: PropTypes.bool,
  id: PropTypes.string.isRequired,
  image: PropTypes.bool,
  imageClassName: PropTypes.string,
  largeFileWarningMessage: PropTypes.message,
  maxSize: PropTypes.number,
  mayRemove: PropTypes.bool,
  message: PropTypes.message,
  name: PropTypes.string.isRequired,
  onChange: PropTypes.func.isRequired,
  providedMessage: PropTypes.message,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.shape({})]),
  warningSize: PropTypes.number,
}

FileInput.defaultProps = {
  accept: undefined,
  center: false,
  dataTransform: defaultDataTransform,
  disabled: false,
  image: false,
  imageClassName: undefined,
  largeFileWarningMessage: undefined,
  maxSize: 16 * 1024 * 1024, // 16 MB
  mayRemove: true,
  message: m.selectAFile,
  changeMessage: m.changeFile,
  providedMessage: m.fileProvided,
  value: undefined,
  warningSize: undefined,
}

export default FileInput
