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
import { connect as storeConnect } from 'react-redux'
import bind from 'autobind-decorator'

import PropTypes from '../../../lib/prop-types'
import Field from '../../../components/form/field'
import Select from '../../../components/select'

const formatOptions = options =>
  Object.keys(options).map(key => ({ value: key, label: options[key] }))

export default function({
  optionsSelector,
  errorSelector,
  fetchingSelector,
  fetchOptions,
  defaultWarning,
  defaultTitle,
  optionsFormatter = formatOptions,
}) {
  @storeConnect(
    function(state) {
      return {
        options: optionsFormatter(optionsSelector(state)),
        error: errorSelector(state),
        fetching: fetchingSelector(state),
      }
    },
    { fetchOptions },
  )
  class FetchSelect extends React.PureComponent {
    static propTypes = {
      autoFocus: PropTypes.bool,
      error: PropTypes.error,
      fetchOptions: PropTypes.func.isRequired,
      fetching: PropTypes.bool,
      menuPlacement: PropTypes.oneOf(['top', 'bottom', 'auto']),
      name: PropTypes.string.isRequired,
      onChange: PropTypes.func,
      options: PropTypes.arrayOf(
        PropTypes.shape({ value: PropTypes.string, label: PropTypes.message }),
      ),
      required: PropTypes.bool,
      title: PropTypes.message,
      warning: PropTypes.message,
    }

    static defaultProps = {
      autoFocus: false,
      menuPlacement: 'auto',
      error: undefined,
      fetching: false,
      onChange: () => null,
      options: [],
      required: false,
      title: defaultTitle,
      warning: defaultWarning,
    }

    componentDidMount() {
      const { fetchOptions } = this.props

      fetchOptions()
    }

    @bind
    handleChange(value) {
      const { onChange, options } = this.props

      onChange(options.find(e => e.value === value))
    }

    render() {
      const {
        name,
        required,
        autoFocus,
        error,
        fetching,
        menuPlacement,
        warning,
        title,
        options,
      } = this.props

      return (
        <Field
          component={Select}
          type="select"
          options={options}
          name={name}
          required={required}
          title={title}
          autoFocus={autoFocus}
          isLoading={fetching}
          warning={Boolean(error) ? warning : undefined}
          menuPlacement={menuPlacement}
          onChange={this.handleChange}
        />
      )
    }
  }

  return FetchSelect
}
