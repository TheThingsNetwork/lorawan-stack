// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'
import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'

import { GET_IS_CONFIGURATION_BASE } from '@console/store/actions/identity-server'

const EMPTY_OBJ = {}

const selectIsStore = state => state.is

export const selectIsConfiguration = state => selectIsStore(state).configuration
export const selectIsConfigurationFetching = createFetchingSelector(GET_IS_CONFIGURATION_BASE)
export const selectIsConfigurationError = createErrorSelector(GET_IS_CONFIGURATION_BASE)

export const selectUserRegistration = state =>
  selectIsConfiguration(state).user_registration || EMPTY_OBJ
export const selectPasswordRequirements = state =>
  selectUserRegistration(state).password_requirements || EMPTY_OBJ

export const selectProfilePictureConfiguration = state =>
  selectIsConfiguration(state).profile_picture || EMPTY_OBJ
export const selectUseGravatarConfiguration = state =>
  selectProfilePictureConfiguration(state).use_gravatar
export const selectDisableUploadConfiguration = state =>
  selectProfilePictureConfiguration(state).disable_upload

export const selectIsUserRightsConfig = state =>
  selectIsConfiguration(state).user_rights || EMPTY_OBJ
export const selectDisableEmail = () => false // Added for compatibility reasons with enterprise.
export const selectDisableName = () => false // Added for compatibility reasons with enterprise.
