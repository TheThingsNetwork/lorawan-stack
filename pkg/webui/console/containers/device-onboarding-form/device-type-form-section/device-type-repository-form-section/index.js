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

import React, { useEffect } from 'react'
import { merge } from 'lodash'

import { useFormContext } from '@ttn-lw/components/form'

import { subtractObject } from '../../utils'

const initialValues = {}

const DeviceTypeRepositoryFormSection = props => {
  const { setValues } = useFormContext()

  useEffect(() => {
    // Set the section's initial values on mount.
    setTimeout(() => setValues(values => merge(values, initialValues)))

    // Remove initial values on on mount (on next tick).
    return () => setValues(values => subtractObject(values, initialValues))
  }, [setValues])

  return <span>Device Type Repository Form Section</span>
}

export default DeviceTypeRepositoryFormSection
