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

import ApiKeyCreated from './templates/api-key-created'
import ApiKeyChanged from './templates/api-key-changed'
import ClientRequested from './templates/client-requested'
import CollaboratorChanged from './templates/collaborator-changed'
import EntityStateChanged from './templates/entity-state-changed'
import PasswordChanged from './templates/password-changed'
import UserRequested from './templates/user-requested'

const notificationMap = {
  api_key_created: ApiKeyCreated,
  api_key_changed: ApiKeyChanged,
  client_requested: ClientRequested,
  collaborator_changed: CollaboratorChanged,
  entity_state_changed: EntityStateChanged,
  password_changed: PasswordChanged,
  user_requested: UserRequested,
}

export default notificationMap
