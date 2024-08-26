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

import { combineReducers } from 'redux'

import init from '@ttn-lw/lib/store/reducers/init'
import status from '@ttn-lw/lib/store/reducers/status'
import fetching from '@ttn-lw/lib/store/reducers/ui/fetching'
import error from '@ttn-lw/lib/store/reducers/ui/error'

import user from './user'
import is from './identity-server'

export default combineReducers({
  init,
  user,
  is,
  ui: combineReducers({
    fetching,
    error,
  }),
  status,
})
