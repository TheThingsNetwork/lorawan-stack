// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import React from 'react'
import { storiesOf } from '@storybook/react'
import { action } from '@storybook/addon-actions'
import { withInfo } from '@storybook/addon-info'

import Field from '../field'
import CheckboxGroup from '../checkbox/group'
import Button from '../button'
import style from './story.styl'
import Form from '.'


const handleSubmit = function (data, { setSubmitting }) {
  action('Submit')(data)
  setTimeout(() => setSubmitting(false), 1000)
}

const containerStyles = {
  maxWidth: '300px',
}

storiesOf('Form', module)
  .addDecorator((story, context) => withInfo({
    inline: true,
    header: false,
    source: true,
    propTables: [ Form ],
  })(story)(context))
  .add('Login', () => (
    <div style={containerStyles}>
      <Form
        onSubmit={handleSubmit}
        initialValues={{
          user_id: '',
          password: '',
        }}
        submitEnabledWhenInvalid
      >
        <Field
          title="Username or Email"
          name="user_id"
          type="text"
          form
        />
        <Field
          title="Password"
          name="password"
          type="password"
          form
        />
        <Button type="submit" message="Login" />
        <Button naked message="Create an account" />
      </Form>
    </div>
  ))
  .add('Rights selection', () => (
    <div style={containerStyles}>
      <Form
        onSubmit={handleSubmit}
        submitEnabledWhenInvalid
      >
        <CheckboxGroup
          className={style.checkboxGroup}
          selectAllTitle="All Application Rights"
          selectAllName="RIGHT_APPLICATION_ALL"
          transfer
          name="rights"
          title="Rights"
          form
        >
          <Field title="Application Rights Read" name="APPLICATION_RIGHT_READ" form />
          <Field title="Application Rights Edit" name="APPLICATION_RIGHT_WRITE" />
          <Field title="Application Device Read" name="APPLICATION_RIGHT_DEVICE_READ" />
          <Field title="Application Info Read" name="RIGHT_APPLICATION_INFO" />
          <Field title="Application Basic Settings" name="RIGHT_APPLICATION_SETTINGS_BASIC" />
          <Field title="Application Api Keys" name="RIGHT_APPLICATION_SETTINGS_API_KEYS" />
          <Field title="Application Collaborators" name="RIGHT_APPLICATION_SETTINGS_COLLABORATORS" />
          <Field title="Application Delete" name="RIGHT_APPLICATION_DELETE" />
          <Field title="Application Delete Dayum Son Herrlo World" name="RIGHT_APPLICATION_DELETE2" />
        </CheckboxGroup>
        <Button type="submit" message="Save" />
      </Form>
    </div>
  ))
