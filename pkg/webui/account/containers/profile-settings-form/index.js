// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import React, { useState, useCallback, useMemo, useEffect } from 'react'
import { useSelector, useDispatch } from 'react-redux'
import md5 from 'md5'
import axios from 'axios'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Radio from '@ttn-lw/components/radio-button'
import FileInput from '@ttn-lw/components/file-input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import ProfilePicture from '@ttn-lw/components/profile-picture'
import Notification from '@ttn-lw/components/notification'
import Link from '@ttn-lw/components/link'
import toast from '@ttn-lw/components/toast'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'

import {
  getClosestProfilePictureBySize,
  isGravatarProfilePicture,
  convertUriToProfilePicture,
} from '@ttn-lw/lib/selectors/profile-picture'
import { selectApplicationRootPath } from '@ttn-lw/lib/selectors/env'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import diff from '@ttn-lw/lib/diff'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import debounce from '@ttn-lw/lib/debounce'

import { checkFromState, mayPurgeEntities } from '@account/lib/feature-checks'

import { updateUser, deleteUser } from '@account/store/actions/user'

import { selectUser, selectUserId } from '@account/store/selectors/user'
import {
  selectUseGravatarConfiguration,
  selectDisableUploadConfiguration,
} from '@account/store/selectors/identity-server'

import validationSchema from './validation-schema'
import m from './messages'

import style from './profile-settings-form.styl'

const promisifiedDeleteUser = attachPromise(deleteUser)

const imageDataTransform = data => {
  // Handle empty value (when file was removed)
  if (data === '') {
    return null
  }

  // Handle selected image file
  const matches = data.match(/^data:(.*);base64,(.*)$/)

  // Discard invalid values
  if (matches && matches.length < 3) {
    return null
  }

  // Return value with correct schema for the API request
  return { embedded: { data: matches[2], mime_type: matches[1] } }
}

const imageDataDecoder = value => {
  if (!Boolean(value) || isGravatarProfilePicture(value)) {
    return null
  }

  return Boolean(value.embedded)
    ? imageUriFromEmbedded(value)
    : getClosestProfilePictureBySize(value, 128)
}

const imageUriFromEmbedded = value =>
  `data:${value.embedded.mime_type};base64,${value.embedded.data}`

const ProfileEditForm = () => {
  const dispatch = useDispatch()

  const user = useSelector(selectUser)
  const userId = useSelector(selectUserId)
  const useGravatarConfig = useSelector(selectUseGravatarConfiguration)
  const disableUploadConfig = useSelector(selectDisableUploadConfiguration)
  const initialProfilePictureSource =
    (useGravatarConfig && isGravatarProfilePicture(user.profile_picture)) ||
    (user.profile_picture === null && useGravatarConfig)
      ? 'gravatar'
      : 'upload'
  const mayPurge = useSelector(state => checkFromState(mayPurgeEntities, state))

  const [profilePictureSource, setProfilePictureSource] = useState(initialProfilePictureSource)
  const [gravatarPreview, setGravatarPreview] = useState(null)
  const [error, setError] = useState()
  const [emailAddress, setEmailAddress] = useState(user.primary_email_address)

  const { debouncedFunction: debouncedSetEmail, cancel: debounceCancel } = useMemo(
    () => debounce(setEmailAddress, 1000),
    [setEmailAddress],
  )

  const validationContext = useMemo(
    () => ({
      disableUploadConfig,
      useGravatarConfig,
      initialProfilePictureSource,
    }),
    [disableUploadConfig, useGravatarConfig, initialProfilePictureSource],
  )

  // Cancel debounced queue when unmounting.
  useEffect(() => () => debounceCancel(), [debounceCancel])

  // Fetch a new gravatar preview when the primary email address changes.
  useEffect(() => {
    const fetchGravatarPreview = async () => {
      const src = `https://gravatar.com/avatar/${md5(emailAddress.toLowerCase())}?s=256&d=404`

      try {
        await axios.get(src)
        setGravatarPreview(convertUriToProfilePicture(src))
      } catch {
        setGravatarPreview(null)
      }
    }
    fetchGravatarPreview()
  }, [emailAddress, setGravatarPreview])

  const handleSubmit = useCallback(
    async (values, { setSubmitting, resetForm }) => {
      setError(undefined)
      let patch = diff(user, validationSchema.cast(values, { context: validationContext }), {
        exclude: ['_profile_picture_source'],
      })
      if (Object.keys(patch).length === 0) {
        patch = { ...values }
        delete patch.profile_picture
        delete patch._profile_picture_source
      }
      try {
        const updatedUser = await dispatch(
          attachPromise(updateUser({ id: user.ids.user_id, patch })),
        )
        resetForm({
          values: {
            ...user,
            ...updatedUser,
            _profile_picture_source: values._profile_picture_source,
          },
        })
        toast({
          title: sharedMessages.success,
          message: m.successMessage,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setError(error)
        setSubmitting(false)
      }
    },
    [dispatch, user, validationContext],
  )

  const handleDelete = useCallback(
    async shouldPurge => {
      try {
        await dispatch(promisifiedDeleteUser(userId, { purge: shouldPurge }))

        // The hard redirect will conclude the deletion by deleting the
        // (now invalid) session cookie and redirecting back to the login screen.
        // The `account-deleted` query will cause a success notification to be
        // shown on the login screen.
        const appRoot = selectApplicationRootPath()
        window.location = `${appRoot}/login?account-deleted`
      } catch {
        toast({
          title: m.deleteAccount,
          message: m.deleteAccountError,
          type: toast.types.ERROR,
        })
      }
    },
    [dispatch, userId],
  )

  const initialValues = validationSchema.cast(
    { profile_picture: undefined, ...user },
    { context: validationContext },
  )

  const gravatarInfoMessage = disableUploadConfig ? m.gravatarInfoGravatarOnly : m.gravatarInfo

  return (
    <Form
      initialValues={initialValues}
      validationSchema={validationSchema}
      validationContext={validationContext}
      onSubmit={handleSubmit}
      error={error}
      validateOnChange
    >
      {useGravatarConfig && !disableUploadConfig && (
        <Form.Field
          name="_profile_picture_source"
          title={m.profilePicture}
          component={Radio.Group}
          onChange={setProfilePictureSource}
        >
          <Radio label={m.useGravatar} value="gravatar" />
          <Radio label={sharedMessages.uploadAnImage} value="upload" />
        </Form.Field>
      )}
      {!disableUploadConfig && profilePictureSource === 'upload' && (
        <Form.Field
          name="profile_picture"
          component={FileInput}
          title={m.imageUpload}
          message={m.chooseImage}
          changeMessage={m.changeImage}
          providedMessage={m.imageProvided}
          accept={['.jpg', '.jpeg', '.png']}
          decode={imageDataDecoder}
          dataTransform={imageDataTransform}
          imageClassName={style.uploadImagePreview}
          required={useGravatarConfig}
          mayRemove
          image
        />
      )}
      {useGravatarConfig && profilePictureSource === 'gravatar' && (
        <Form.InfoField title={m.gravatarImage}>
          <div className={style.profilePictureInfo}>
            <ProfilePicture
              profilePicture={
                isGravatarProfilePicture(user.profile_picture) &&
                user.primary_email_address === emailAddress
                  ? user.profile_picture
                  : gravatarPreview
              }
              className={style.profilePicture}
            />
            <Notification
              small
              info
              content={{
                ...gravatarInfoMessage,
                values: {
                  link: val => (
                    <Link.Anchor primary external href="https://gravatar.com" key="gravatar-link">
                      {val}
                    </Link.Anchor>
                  ),
                },
              }}
              className={style.profilePictureNotification}
            />
          </div>
        </Form.InfoField>
      )}
      {!useGravatarConfig && disableUploadConfig && (
        <Form.InfoField title={m.profilePicture}>
          <div className={style.profilePictureInfo}>
            <ProfilePicture
              profilePicture={user.profile_picture}
              className={style.profilePicture}
            />
            <Notification
              small
              info
              content={m.profilePictureDisabled}
              className={style.profilePictureNotification}
            />
          </div>
        </Form.InfoField>
      )}
      <Form.Field name="ids.user_id" component={Input} title={sharedMessages.userId} disabled />
      <Form.Field name="name" component={Input} title={sharedMessages.name} />
      <Form.Field
        name="primary_email_address"
        component={Input}
        title={sharedMessages.emailAddress}
        description={m.primaryEmailAddressDescription}
        onChange={debouncedSetEmail}
        required
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
        <DeleteModalButton
          title={m.deleteTitle}
          message={m.deleteAccount}
          entityId={userId}
          entityName={user.name}
          onApprove={handleDelete}
          shouldConfirm
          mayPurge={mayPurge}
          defaultMessage={m.deleteWarning}
          purgeMessage={m.purgeWarning}
          confirmMessage={m.deleteConfirmMessage}
        />
      </SubmitBar>
    </Form>
  )
}

export default ProfileEditForm
