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

import { isOtherOption } from '../utils'

// State.

export const defaultState = {
  version: {
    brand: undefined,
    model: undefined,
    hardwareVersion: undefined,
    firmwareVersion: undefined,
    band: undefined,
  },
  error: undefined,
}

// Action types.

const SET_BRAND = 'SET_BRAND'
const SET_MODEL = 'SET_MODEL'
const SET_HW_VERSION = 'SET_HW_VERSION'
const SET_FW_VERSION = 'SET_FW_VERSION'
const SET_BAND_ID = 'SET_BAND_ID'
const SET_ERROR = 'SET_ERROR'

// Action creators.

export const setBrand = brand => ({ type: SET_BRAND, brand })
export const setModel = model => ({ type: SET_MODEL, model })
export const setHwVersion = hardwareVersion => ({ type: SET_HW_VERSION, hardwareVersion })
export const setFwVersion = firmwareVersion => ({ type: SET_FW_VERSION, firmwareVersion })
export const setBand = band => ({ type: SET_BAND_ID, band })
export const setError = error => ({ type: SET_ERROR, error })

// Selectors.

export const selectVersion = state => state.version
export const selectBrand = state => selectVersion(state).brand
export const hasSelectedBrand = state => Boolean(selectBrand(state))
export const selectModel = state => selectVersion(state).model
export const hasSelectedModel = state => Boolean(selectModel(state))
export const selectHwVersion = state => selectVersion(state).hardwareVersion
export const hasSelectedHwVersion = state => Boolean(selectHwVersion(state))
export const selectFwVersion = state => selectVersion(state).firmwareVersion
export const hasSelectedFwVersion = state => Boolean(selectFwVersion(state))
export const selectBand = state => selectVersion(state).band
export const hasSelectedBand = state => Boolean(selectBand(state))
export const selectError = state => state.error

const valueSelectors = [selectBrand, selectModel, selectHwVersion, selectFwVersion, selectBand]
const selectors = [
  hasSelectedBrand,
  hasSelectedModel,
  hasSelectedHwVersion,
  hasSelectedFwVersion,
  hasSelectedBand,
]
export const hasAnySelectedOther = state =>
  valueSelectors.some(selector => isOtherOption(selector(state)))
export const hasCompletedSelection = state => selectors.every(selector => selector(state))

// Reducer.

export default (state, action) => {
  const { version } = state
  switch (action.type) {
    case SET_BRAND:
      if (version.brand === action.brand) {
        return state
      }

      return {
        ...state,
        version: {
          ...defaultState.version,
          brand: action.brand,
        },
      }
    case SET_MODEL:
      if (version.model === action.model) {
        return state
      }

      return {
        ...state,
        version: {
          ...defaultState.version,
          brand: state.version.brand,
          model: action.model,
        },
      }
    case SET_HW_VERSION:
      if (version.hardwareVersion === action.hardwareVersion) {
        return state
      }

      return {
        ...state,
        version: {
          ...defaultState.version,
          brand: state.version.brand,
          model: state.version.model,
          hardwareVersion: action.hardwareVersion,
        },
      }
    case SET_FW_VERSION:
      if (version.firmwareVersion === action.firmwareVersion) {
        return state
      }

      return {
        ...state,
        version: {
          ...defaultState.version,
          brand: state.version.brand,
          model: state.version.model,
          hardwareVersion: state.version.hardwareVersion,
          firmwareVersion: action.firmwareVersion,
        },
      }
    case SET_BAND_ID:
      return {
        ...state,
        version: {
          ...state.version,
          band: action.band,
        },
      }
    case SET_ERROR:
      return {
        ...state,
        error: action.error,
      }
  }

  return state
}
