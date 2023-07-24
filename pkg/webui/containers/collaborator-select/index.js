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

import React, { useCallback, useRef } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages, useIntl } from 'react-intl'
import { components } from 'react-select'

import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'
import Icon from '@ttn-lw/components/icon'
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

import { composeOption } from './util'

import styles from './collaborator-select.styl'

const customMenu = props => (
  <components.Menu {...props}>
    <Message content={m.suggestions} className="ml-cs-s mt-cs-xxs mb-cs-xxs" component="h4" />
    {props.children}
  </components.Menu>
)

const SingleValue = props => (
  <components.SingleValue {...props}>
    <Icon icon={props.data.icon} className="mr-cs-xs" />
    {props.data.label}
  </components.SingleValue>
)

SingleValue.propTypes = {
  data: PropTypes.shape({
    icon: PropTypes.string.isRequired,
    description: PropTypes.string,
    label: PropTypes.string.isRequired,
    value: PropTypes.string.isRequired,
  }).isRequired,
}

const m = defineMessages({
  noOptionsMessage: 'No matching user or organization was found',
  suggestions: 'Suggestions',
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
  const searchResults = useSelector(selectSearchResultAccountIds)
  const searchResultsRef = useRef()
  searchResultsRef.current = searchResults
  const handleNoOptions = useCallback(() => formatMessage(m.noOptionsMessage), [formatMessage])
  const collaboratorOf = {
    path: `${entity}_ids.${entity}_id`,
    id: entityId,
  }
  const onlyUsers = entity === 'organization'

  const handleLoadingOptions = useCallback(
    async value => {
      if (Boolean(value)) {
        try {
          await dispatch(attachPromise(searchAccounts(value, onlyUsers, collaboratorOf)))
          const newOptions = searchResultsRef?.current?.map(account => ({
            value:
              'user_ids' in account
                ? account.user_ids?.user_id
                : account.organization_ids?.organization_id,
            label:
              'user_ids' in account
                ? account.user_ids?.user_id
                : account.organization_ids?.organization_id,
            icon: 'user_ids' in account ? 'user' : 'organization',
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
          return []
        }
      }
    },
    [dispatch, onlyUsers, searchResultsRef, collaboratorOf, formatMessage],
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
        noOptionsMessage={handleNoOptions}
        loadOptions={handleLoadingOptions}
        showOptionIcon
        openMenuOnFocus
        className={styles.collaboratorSelect}
        customComponents={{ SingleValue, Menu: customMenu }}
        maxMenuHeight={300}
        value={values[name] ? composeOption(values[name]) : null}
      />
      {isResctrictedUser && (
        <Button
          icon="user"
          onClick={handleSetYourself}
          message={{ ...m.setYourself, values: { name: name.replace('_', ' ') } }}
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
      icon: PropTypes.string.isRequired,
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
      collaborator => collaborator.icon === 'user',
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
        disabled={isResctrictedUser || collaboratorsList.length === 1}
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
