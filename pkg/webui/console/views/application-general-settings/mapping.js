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

export const mapAttributesToFormValue = attributesType =>
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

export const mapAttributesFormValueToAttributes = formValue =>
  (formValue &&
    formValue.reduce(
      (result, { key, value }) => ({
        ...result,
        [key]: value,
      }),
      {},
    )) ||
  null

export const mapFormValuesToApplication = function (values) {
  return {
    name: values.name,
    description: values.description,
    attributes: mapAttributesFormValueToAttributes(values.attributes),
  }
}

export const mapApplicationToFormValues = application => ({
  ids: {
    application_id: application.ids.application_id,
  },
  name: application.name,
  description: application.description,
  attributes: mapAttributesToFormValue(application.attributes),
})
