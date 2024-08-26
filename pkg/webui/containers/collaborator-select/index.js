// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useRef, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages, useIntl } from 'react-intl'
import { components } from 'react-select'

import Icon, { IconOrganization, IconUser } from '@ttn-lw/components/icon'
import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'
import Button from '@ttn-lw/components/button'
import { useFormContext } from '@ttn-lw/components/form'

import RequireRequest from '@ttn-lw/lib/components/require-request'
import Message from '@ttn-lw/lib/components/message'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { searchAccounts } from '@ttn-lw/lib/store/actions/search-accounts'
import { selectSearchResultAccountIds } from '@ttn-lw/lib/store/selectors/search-accounts'
import PropTypes from '@ttn-lw/lib/prop-types'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import { selectCollaborators } from '@ttn-lw/lib/store/selectors/collaborators'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { composeOption } from './util'

import styles from './collaborator-select.styl'

const customMenu = props => {
  const { showSuggestions } = props.selectProps

  return (
    <components.Menu {...props}>
      {showSuggestions && (
        <Message
          content={sharedMessages.suggestions}
          className="ml-cs-s mt-cs-xxs mb-cs-xxs"
          component="h4"
        />
      )}
      {props.children}
    </components.Menu>
  )
}

const SingleValue = props => (
  <components.SingleValue {...props} className="d-flex al-center">
    <Icon icon={props.data.icon} className="mr-cs-xs" />
    {props.data.label}
  </components.SingleValue>
)

SingleValue.propTypes = {
  data: PropTypes.shape({
    icon: PropTypes.icon.isRequired,
    description: PropTypes.string,
    label: PropTypes.string.isRequired,
    value: PropTypes.string.isRequired,
  }).isRequired,
}

const m = defineMessages({
  setYourself: 'Set yourself as {name}',
})

const Suggest = ({
  userId,
  name,
  initialOptions,
  userInputCustomComponent,
  entity,
  entityId,
  isResctrictedUser,
  ...rest
}) => {
  const dispatch = useDispatch()
  const { formatMessage } = useIntl()
  const { setFieldValue, values } = useFormContext()
  const collaboratorsList = useSelector(selectCollaborators)
  const searchResults = useSelector(selectSearchResultAccountIds)
  const [showSuggestions, setShowSuggestions] = useState(collaboratorsList !== 0)
  const searchResultsRef = useRef()
  searchResultsRef.current = searchResults
  const noOptionsMessage = useCallback(
    () => formatMessage(sharedMessages.noMatchingUserFound),
    [formatMessage],
  )

  const onlyUsers = entity === 'organization'

  const handleLoadingOptions = useCallback(
    async value => {
      if (Boolean(value)) {
        try {
          await dispatch(attachPromise(searchAccounts(value, onlyUsers)))
          setShowSuggestions(searchResultsRef?.current?.length !== 0)
          const newOptions = searchResultsRef?.current?.map(account => ({
            value:
              'user_ids' in account
                ? `user#${account.user_ids?.user_id}`
                : `organization#${account.organization_ids?.organization_id}`,
            label:
              'user_ids' in account
                ? account.user_ids?.user_id
                : account.organization_ids?.organization_id,
            icon: 'user_ids' in account ? IconUser : IconOrganization,
          }))

          const translatedOptions = newOptions?.map(option => {
            const { label, labelValues = {} } = option
            if (typeof label === 'object' && label.id && label.defaultMessage) {
              return { ...option, label: formatMessage(label, labelValues) }
            }

            return option
          })

          return translatedOptions
        } catch (error) {
          setShowSuggestions(false)
          return []
        }
      }
    },
    [dispatch, onlyUsers, searchResultsRef, formatMessage],
  )

  const handleSetYourself = useCallback(
    e => {
      e.preventDefault()
      setFieldValue(name, { user_ids: { user_id: userId } })
    },
    [setFieldValue, name, userId],
  )

  return (
    <>
      <Field
        {...rest}
        name={name}
        defaultOptions={initialOptions}
        component={Select.Suggested}
        noOptionsMessage={noOptionsMessage}
        loadOptions={handleLoadingOptions}
        showOptionIcon
        openMenuOnFocus
        className={styles.collaboratorSelect}
        customComponents={{ SingleValue, Menu: customMenu }}
        maxMenuHeight={300}
        value={values[name] ? composeOption(values[name]) : null}
        showSuggestions={showSuggestions}
      />
      {isResctrictedUser && (
        <Button
          icon={IconUser}
          onClick={handleSetYourself}
          message={{ ...m.setYourself, values: { name: name.replace('_', ' ') } }}
          secondary
        />
      )}
    </>
  )
}

Suggest.propTypes = {
  entity: PropTypes.string.isRequired,
  entityId: PropTypes.string.isRequired,
  initialOptions: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string.isRequired,
      value: PropTypes.string.isRequired,
      icon: PropTypes.icon.isRequired,
    }),
  ),
  isResctrictedUser: PropTypes.bool,
  name: PropTypes.string.isRequired,
  userId: PropTypes.string,
  userInputCustomComponent: PropTypes.shape({
    SingleValue: PropTypes.func,
  }),
}

Suggest.defaultProps = {
  initialOptions: [],
  userInputCustomComponent: {},
  isResctrictedUser: false,
  userId: undefined,
}

const CollaboratorSelect = ({ userId, name, entity, entityId, isResctrictedUser, ...rest }) => {
  const collaboratorsList = useSelector(selectCollaborators)
  const firstEightCollaborators = collaboratorsList
    .slice(0, 7)
    .map(collaborator => composeOption(collaborator))

  let collaboratorOptions = firstEightCollaborators
  if (entity === 'organization') {
    collaboratorOptions = firstEightCollaborators.filter(
      collaborator => collaborator.icon === IconUser,
    )
  }

  return (
    <RequireRequest requestAction={getCollaboratorsList(entity.toLowerCase(), entityId)}>
      <Suggest
        {...rest}
        userId={userId}
        name={name}
        initialOptions={collaboratorOptions}
        entity={entity.toLowerCase()}
        entityId={entityId}
        disabled={isResctrictedUser}
        isResctrictedUser={isResctrictedUser}
      />
    </RequireRequest>
  )
}

CollaboratorSelect.propTypes = {
  entity: PropTypes.string,
  entityId: PropTypes.string,
  isResctrictedUser: PropTypes.bool,
  name: PropTypes.string.isRequired,
  userId: PropTypes.string,
}

CollaboratorSelect.defaultProps = {
  entity: undefined,
  entityId: undefined,
  isResctrictedUser: false,
  userId: undefined,
}

export default CollaboratorSelect
