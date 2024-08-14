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

import React, { useCallback } from 'react'
import { useSelector } from 'react-redux'
import { createSelector } from 'reselect'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import ApiKeysTable from '@console/containers/api-keys-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getApiKeysList } from '@console/store/actions/api-keys'

import { selectApiKeys, selectApiKeysTotalCount } from '@console/store/selectors/api-keys'
import { selectUserId } from '@console/store/selectors/user'

const UserApiKeysList = () => {
  const userId = useSelector(selectUserId)

  const baseDataSelectors = createSelector(
    [selectApiKeys, selectApiKeysTotalCount],
    (keys, totalCount) => ({
      keys,
      totalCount,
    }),
  )

  const baseDataSelector = useCallback(
    state => baseDataSelectors(state, userId),
    [baseDataSelectors, userId],
  )

  const getApiKeys = React.useCallback(
    filters => getApiKeysList('users', userId, filters),
    [userId],
  )

  return (
    <>
      <IntlHelmet title={sharedMessages.personalApiKeys} />
      <ApiKeysTable baseDataSelector={baseDataSelector} getItemsAction={getApiKeys} />
    </>
  )
}

export default UserApiKeysList
