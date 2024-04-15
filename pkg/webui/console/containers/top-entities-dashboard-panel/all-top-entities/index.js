// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import {
  selectBookmarksList,
  selectBookmarksTotalCount,
} from '@console/store/selectors/user-preferences'

import EntitiesList from '../list'

const AllTopEntitiesList = ({ loadNextPage }) => {
  const headers = [
    {
      name: 'icon',
      render: icon => <Icon icon={icon} />,
    },
    {
      name: 'name',
      displayName: 'Name',
      render: (name, id) => <Message content={name === '' ? id : name} />,
    },
  ]

  return (
    <EntitiesList
      loadNextPage={loadNextPage}
      itemsCountSelector={selectBookmarksTotalCount}
      itemsSelector={selectBookmarksList}
      headers={headers}
      emptyMessage={'No top entities yet'}
      emptyDescription={'Your most visited, and bookmarked entities will be listed here.'}
    />
  )
}

AllTopEntitiesList.propTypes = {
  loadNextPage: PropTypes.func.isRequired,
}

export default AllTopEntitiesList