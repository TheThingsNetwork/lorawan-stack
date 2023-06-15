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

import React from 'react'
import { useSelector } from 'react-redux'
import { components } from 'react-select'

import Icon from '@ttn-lw/components/icon'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import PropTypes from '@ttn-lw/lib/prop-types'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import { selectCollaborators } from '@ttn-lw/lib/store/selectors/collaborators'

import AutoSuggest from '../autosuggest'

import composeOption from './util'

import styles from './user-input.styl'

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

const UserInput = ({ onlyCollaborators, onlyUsers, entity, entityId, ...rest }) => {
  const collaboratorsList = useSelector(selectCollaborators)

  const firstEightCollaborators = collaboratorsList
    .slice(0, 7)
    .map(collaborator => composeOption(collaborator))

  let collaboratorOptions = firstEightCollaborators
  if (onlyUsers) {
    collaboratorOptions = firstEightCollaborators.filter(
      collaborator => collaborator.icon === 'user',
    )
  }

  return (
    <RequireRequest requestAction={onlyCollaborators ? getCollaboratorsList(entity, entityId) : []}>
      <AutoSuggest
        {...rest}
        initialOptions={onlyCollaborators ? collaboratorOptions : []}
        onlyCollaborators={onlyCollaborators}
        showOptionIcon
        className={styles.userInput}
        userInputCustomComponent={{ SingleValue }}
        entity={entity}
        entityId={entityId}
      />
    </RequireRequest>
  )
}

UserInput.propTypes = {
  // When the input is a multi-select `controlShouldRenderValue` should be false and is required.
  controlShouldRenderValue: PropTypes.bool,
  // When the users to be selected can only be collaborators, entity and entityId are required.
  entity: PropTypes.string,
  entityId: PropTypes.string,
  // When the input is a multi-select `isClearable` should be false and is required.
  isClearable: PropTypes.bool,
  onlyCollaborators: PropTypes.bool,
  onlyUsers: PropTypes.bool,
  // When the users to be selected can only be collaborators, `openMenuOnFocus` has to be true and is required.
  openMenuOnFocus: PropTypes.bool,
}

UserInput.defaultProps = {
  entity: undefined,
  entityId: undefined,
  onlyCollaborators: false,
  onlyUsers: false,
  isClearable: true,
  controlShouldRenderValue: true,
  openMenuOnFocus: false,
}

export default UserInput
