// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import { connect } from 'react-redux'

import { getCollaboratorId } from '@ttn-lw/lib/selectors/id'

import { selectUserId } from '@console/store/selectors/logout'
import { selectUserById } from '@console/store/selectors/users'

const isCollaboratorUser = collaborator => collaborator.ids && 'user_ids' in collaborator.ids

export default CollaboratorForm =>
  connect((state, { collaborator, update = false, ...props }) => {
    const collaboratorId = getCollaboratorId(collaborator)
    const isUser = update && isCollaboratorUser(collaborator)
    const isAdmin =
      props.isAdmin || (isUser && Boolean(selectUserById(state, collaboratorId).admin))
    const isCurrentUser = isUser && selectUserId(state) === collaboratorId

    return {
      collaboratorId,
      isCollaboratorUser: isUser,
      isCollaboratorAdmin: isAdmin,
      isCollaboratorCurrentUser: isCurrentUser,
      currentUserId: selectUserId(state),
    }
  })(CollaboratorForm)
