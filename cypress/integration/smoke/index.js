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

import deviceTests from './devices'
import registrationTests from './registration'
import applicationTests from './applications'
import featureToggleTests from './feature-toggles'
import gatewayTests from './gateways'
import organizationTests from './organizations'
import forgotPasswordTests from './forgot-password'
import contactInfoValidationTests from './contact-info-validation'
import authorizationTests from './authorization'
import profileSettingsTests from './profile-settings'

export default [
  ...registrationTests,
  ...applicationTests,
  ...deviceTests,
  ...featureToggleTests,
  ...gatewayTests,
  ...organizationTests,
  ...forgotPasswordTests,
  ...contactInfoValidationTests,
  ...authorizationTests,
  ...profileSettingsTests,
]
