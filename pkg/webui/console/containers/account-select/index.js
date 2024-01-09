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
import { useIntl } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'
import { components } from 'react-select'

import Field from '@ttn-lw/components/form/field'
import Select from '@ttn-lw/components/select'
import Icon from '@ttn-lw/components/icon'

import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { searchAccounts } from '@ttn-lw/lib/store/actions/search-accounts'
import { selectSearchResultAccountIds } from '@ttn-lw/lib/store/selectors/search-accounts'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import styles from './account-select.styl'

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

const Suggest = ({ entity, ...rest }) => {
  const dispatch = useDispatch()
  const searchResults = useSelector(selectSearchResultAccountIds)
  const searchResultsRef = useRef()
  searchResultsRef.current = searchResults
  const { formatMessage } = useIntl()
  const noOptionsMessage = useCallback(
    () => formatMessage(sharedMessages.noMatchingUserFound),
    [formatMessage],
  )
  const onlyUsers = entity === 'organization'

  const handleLoadingOptions = useCallback(
    async value => {
      if (value.length >= 1) {
        await dispatch(attachPromise(searchAccounts(value, onlyUsers)))
        const newOptions = searchResultsRef.current.map(account => ({
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
      }
    },
    [dispatch, searchResultsRef, formatMessage, onlyUsers],
  )

  return (
    <Field
      {...rest}
      component={Select.Suggested}
      openMenuOnFocus={false}
      openMenuOnClick={false}
      noOptionsMessage={noOptionsMessage}
      loadOptions={handleLoadingOptions}
      autoFocus
      showOptionIcon
      maxMenuHeight={300}
      customComponents={{ SingleValue }}
    />
  )
}

Suggest.propTypes = {
  entity: PropTypes.string.isRequired,
}

const AccountSelect = ({ entity, ...rest }) => (
  <Suggest {...rest} className={styles.userSelect} entity={entity.toLowerCase()} />
)

AccountSelect.propTypes = {
  entity: PropTypes.string.isRequired,
}

export default AccountSelect
