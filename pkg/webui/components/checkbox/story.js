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
import { withInfo } from '@storybook/addon-info'
import bind from 'autobind-decorator'

import Field from '../field'
import style from './story.styl'

import CheckboxGroup from './group'

const entries = [ ...new Array(10) ]
  .map((e, i) => ({
    title: `value-${i}`,
    name: `value-${i}`,
  }))

@bind
class Example extends React.Component {
  state = {
    values: {},
  }

  setValues (values) {
    this.setState({ values })
  }

  onChange (name, value) {
    this.setState(prev => ({
      values: {
        ...prev.values,
        [name]: value,
      },
    }))
  }

  render () {
    const { groupProps } = this.props

    return (
      <CheckboxGroup
        onChange={this.onChange}
        setValues={this.setValues}
        {...groupProps}
      >
        <Field title="value-1" name="value-1" type="checkbox" />
        <Field title="value-2" name="value-2" type="checkbox" />
        <Field title="value-3" name="value-3" type="checkbox" />
        <Field title="value-4" name="value-4" type="checkbox" />
        <Field title="value-5" name="value-5" type="checkbox" />
        <Field title="value-6" name="value-6" type="checkbox" />
        <Field title="value-7" name="value-7" type="checkbox" />
        <Field title="value-8" name="value-8" type="checkbox" />
        <Field title="value-9" name="value-9" type="checkbox" />
      </CheckboxGroup>
    )
  }
}

const onSearch = (entry, query) => entry.title.includes(query)

storiesOf('Checkbox', module)
  .addDecorator((story, context) => withInfo({
    inline: true,
    header: false,
    source: false,
    propTables: [ CheckboxGroup ],
  })(story)(context))
  .add('Group', () => (
    <Example
      groupProps={{
        className: style.checkboxGroup,
        name: 'checkbox-group',
        title: 'Title',
        entries: [],
        form: false,
      }}
    />
  ))
  .add('Search Group', () => (
    <Example
      groupProps={{
        className: style.checkboxGroup,
        name: 'checkbox-group',
        title: 'Title',
        entries,
        search: true,
        onSearch,
        form: false,
      }}
    />
  ))
  .add('Transfer Group', () => (
    <Example
      groupProps={{
        className: style.checkboxGroupTransfer,
        title: 'Title',
        name: 'checkbox-group',
        transfer: true,
        entries,
        form: false,
      }}
    />
  ))
  .add('Search/Transfer Group', () => (
    <Example
      groupProps={{
        className: style.checkboxGroupTransfer,
        title: 'Title',
        name: 'checkbox-group',
        transfer: true,
        search: true,
        onSearch,
        entries,
        form: false,
      }}
    />
  ))
