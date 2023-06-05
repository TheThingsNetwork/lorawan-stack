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

import React, { useCallback, useEffect, useRef } from 'react'
import { defineMessages, useIntl } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'
import { components } from 'react-select'

import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { searchAccounts } from '@ttn-lw/lib/store/actions/search-accounts'
import { selectSearchResults } from '@ttn-lw/lib/store/selectors/search-accounts'

const m = defineMessages({
  noOptionsMessage: 'No matching user or organization was found',
  suggestions: 'Suggestions',
})

const customMenu = props => (
  <components.Menu {...props}>
    <Message content={m.suggestions} className="ml-cs-s mt-cs-s mb-cs-s" component="h4" />
    {props.children}
  </components.Menu>
)

const AutoSuggest = ({
  initialOptions,
  onlyCollaborators,
  onlyUsers,
  userInputCustomComponent,
  entity,
  entityId,
  ...rest
}) => {
  const dispatch = useDispatch()
  const searchResults = useSelector(selectSearchResults)
  const { formatMessage } = useIntl()

  const handleNoOptions = useCallback(() => formatMessage(m.noOptionsMessage), [formatMessage])

  const searchResultsRef = useRef(searchResults)

  const collaboratorOf = onlyCollaborators
    ? {
        path: `${entity}_ids.${entity}_id`,
        id: entityId,
      }
    : undefined

  useEffect(() => {
    searchResultsRef.current = searchResults
  }, [searchResults])

  const handleLoadingOptions = useCallback(
    async value => {
      if (value.length >= 1) {
        await dispatch(attachPromise(searchAccounts(value, onlyUsers, collaboratorOf)))
        const newOptions = searchResults.map(account => ({
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

        const translatedOptions = newOptions.map(option => {
          const { label, labelValues = {} } = option
          if (typeof label === 'object' && label.id && label.defaultMessage) {
            return { ...option, label: formatMessage(label, labelValues) }
          }

          return option
        })

        return translatedOptions
      }
    },
    [dispatch, onlyUsers, searchResults, collaboratorOf, formatMessage],
  )

  return (
    <Field
      {...rest}
      defaultOptions={initialOptions}
      component={Select}
      noOptionsMessage={handleNoOptions}
      loadOptions={handleLoadingOptions}
      autoFocus
      maxMenuHeight={300}
      hasAutosuggest
      customComponents={
        onlyCollaborators
          ? { ...userInputCustomComponent, Menu: customMenu }
          : { ...userInputCustomComponent }
      }
    />
  )
}

AutoSuggest.propTypes = {
  entity: PropTypes.string,
  entityId: PropTypes.string,
  fetching: PropTypes.bool,
  initialOptions: PropTypes.arrayOf(
    PropTypes.shape({ value: PropTypes.string, label: PropTypes.message }),
  ),
  onlyCollaborators: PropTypes.bool,
  onlyUsers: PropTypes.bool,
  userInputCustomComponent: PropTypes.shape({
    SingleValue: PropTypes.func,
  }),
}

AutoSuggest.defaultProps = {
  fetching: false,
  initialOptions: [],
  onlyCollaborators: false,
  onlyUsers: false,
  userInputCustomComponent: {},
  entity: undefined,
  entityId: undefined,
}

export default AutoSuggest
