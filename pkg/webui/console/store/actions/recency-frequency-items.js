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

import { createAction } from '@reduxjs/toolkit'

export const GET_RECENCY_FREQUENCY_ITEMS = 'GET_TOP_ENTITIES'
export const getTopRecencyFrequencyItems = createAction(GET_RECENCY_FREQUENCY_ITEMS)

export const TRACK_RECENCY_FREQUENCY_ITEM = 'TRACK_RECENCY_FREQUENCY_ITEM'
export const trackRecencyFrequencyItem = createAction(TRACK_RECENCY_FREQUENCY_ITEM, (type, id) => ({
  payload: { type, id },
}))

export const DELETE_RECENCY_FREQUENCY_ITEM = 'DELETE_RECENCY_FREQUENCY_ITEM'
export const deleteRecencyFrequencyItem = createAction(
  DELETE_RECENCY_FREQUENCY_ITEM,
  (type, id) => ({
    payload: { type, id },
  }),
)
