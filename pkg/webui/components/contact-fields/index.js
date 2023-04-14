// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useState } from 'react'
import { defineMessages } from 'react-intl'

import Form, { useFormContext } from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import Radio from '@ttn-lw/components/radio-button'
import Button from '@ttn-lw/components/button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  technicalContact: 'Add technical contact',
  administrativeContact: 'Add administrative contact',
  administrativeId: 'Administrative contact ID',
  technicalId: 'Technical contact ID',
  administrativeType: 'Administrative contact type',
  technicalType: 'Technical contact type',
  technicalContactRemove: 'Remove technical contact',
  administrativeContactRemove: 'Remove administrative contact',
})

const ContactFields = ({ name, hasInitialValue }) => {
  const { values, setFieldValue } = useFormContext()
  const [isAddingContact, setIsAddingContact] = useState(hasInitialValue)
  const addContact = useCallback(e => {
    e.preventDefault()
    setIsAddingContact(true)
  }, [])

  const removeContact = useCallback(
    e => {
      e.preventDefault()
      setIsAddingContact(false)
      setFieldValue(`_${name}_contact_type`, '', true)
      setFieldValue(`_${name}_contact_id`, '', true)
    },
    [setFieldValue, name],
  )

  const typeTitle = m[`${name}Type`]
  const idTitle = m[`${name}Id`]
  const addContactMessage = m[`${name}Contact`]
  const removeContactMessage = m[`${name}ContactRemove`]

  return (
    <>
      {isAddingContact ? (
        <>
          <Form.Field
            name={`_${name}_contact_type`}
            title={typeTitle}
            component={Radio.Group}
            horizontal
          >
            <Radio label={sharedMessages.user} value="user" />
            <Radio label={sharedMessages.organization} value="organization" />
          </Form.Field>
          <Form.Field name={`_${name}_contact_id`} component={Input} title={idTitle} />
          <Button
            type="button"
            icon="delete"
            message={removeContactMessage}
            danger
            onClick={removeContact}
          />
        </>
      ) : (
        <Button message={addContactMessage} icon="add" type="button" onClick={addContact} />
      )}
    </>
  )
}

ContactFields.propTypes = {
  hasInitialValue: PropTypes.bool.isRequired,
  name: PropTypes.string.isRequired,
}

export default ContactFields
