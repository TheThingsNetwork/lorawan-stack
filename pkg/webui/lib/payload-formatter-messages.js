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

import { defineMessages } from 'react-intl'

import TYPES from '@console/constants/formatter-types'

import sharedMessages from './shared-messages'

const m = defineMessages({
  repository: 'Repository',
  javascript: 'Javascript',
  cayennelpp: 'CayenneLPP',
})

export default Object.freeze({
  [TYPES.JAVASCRIPT]: m.javascript,
  [TYPES.REPOSITORY]: m.repository,
  [TYPES.NONE]: sharedMessages.none,
  [TYPES.GRPC]: m.grpc,
  [TYPES.CAYENNELPP]: m.cayennelpp,
})
