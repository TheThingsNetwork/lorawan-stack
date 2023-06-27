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

import { id as idRegexp } from '@ttn-lw/lib/regexp'

export const encodeAttributes = formValue =>
  (Array.isArray(formValue) &&
    formValue.reduce(
      (result, { key, value }) => ({
        ...result,
        [key]: value,
      }),
      {},
    )) ||
  undefined

export const decodeAttributes = attributesType =>
  (attributesType &&
    Object.keys(attributesType).reduce(
      (result, key) =>
        result.concat({
          key,
          value: attributesType[key],
        }),
      [],
    )) ||
  []

export const attributesCountCheck = object =>
  object === undefined ||
  object === null ||
  (object instanceof Object && Object.keys(object).length <= 10)
export const attributeValidCheck = object =>
  object === undefined ||
  object === null ||
  (object instanceof Object && Object.values(object).every(attribute => Boolean(attribute)))

export const attributeTooShortCheck = object =>
  object === undefined ||
  object === null ||
  (object instanceof Object && Object.keys(object).every(key => RegExp(idRegexp).test(key)))

export const attributeKeyTooLongCheck = object =>
  object === undefined ||
  object === null ||
  (object instanceof Object && Object.keys(object).every(key => key.length <= 36))

export const attributeValueTooLongCheck = object =>
  object === undefined ||
  object === null ||
  (object instanceof Object && Object.values(object).every(value => value.length <= 200))
