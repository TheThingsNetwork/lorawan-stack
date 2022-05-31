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

import { selectUserIsAdmin, selectUserRights } from '@account/store/selectors/user'

export const checkFromState = (featureCheck, state) =>
  featureCheck.check(featureCheck.rightsSelector(state))

// Admin feature checks.
export const mayPerformAdminActions = {
  rightsSelector: selectUserIsAdmin,
  check: isAdmin => isAdmin,
}

export const mayPerformAllClientActions = {
  rightsSelector: selectUserRights,
  check: rights => rights.includes('RIGHT_CLIENT_ALL'),
}

export const mayPurgeEntities = mayPerformAdminActions
