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
import ReactSelect, { components } from 'react-select'
import AsyncSelect from 'react-select/async'
import { defineMessage, injectIntl } from 'react-intl'
import bind from 'autobind-decorator'
import classnames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import Icon from '../icon'
import Button from '../button'

import style from './select.styl'

const m = defineMessage({
  remove: 'Remove',
})

const Input = props => {
  const { selectProps } = props

  return <components.Input {...props} aria-describedby={selectProps['aria-describedby']} />
}

Input.propTypes = {
  selectProps: PropTypes.shape({
    'aria-describedby': PropTypes.string,
  }).isRequired,
}

// Map value to a plain string, instead of value object.
// See: https://github.com/JedWatson/react-select/issues/2841
const getValue = (opts, val) => opts.find(o => o.value === val)

class Select extends React.PureComponent {
  static propTypes = {
    className: PropTypes.string,
    customComponents: PropTypes.shape({
      Option: PropTypes.func,
      SingleValue: PropTypes.func,
    }),
    disabled: PropTypes.bool,
    error: PropTypes.bool,
    hasAutosuggest: PropTypes.bool,
    id: PropTypes.string,
    inputWidth: PropTypes.inputWidth,
    intl: PropTypes.shape({
      formatMessage: PropTypes.func,
    }).isRequired,
    loadOptions: PropTypes.func,
    menuPlacement: PropTypes.string,
    name: PropTypes.string.isRequired,
    onBlur: PropTypes.func,
    onChange: PropTypes.func,
    onFocus: PropTypes.func,
    options: PropTypes.arrayOf(
      PropTypes.shape({
        value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
        label: PropTypes.message,
      }),
    ),
    placeholder: PropTypes.message,
    showOptionIcon: PropTypes.bool,
    value: PropTypes.oneOf([PropTypes.string, PropTypes.shape({})]),
    warning: PropTypes.bool,
  }

  static defaultProps = {
    className: undefined,
    onChange: () => null,
    onBlur: () => null,
    onFocus: () => null,
    options: [],
    disabled: false,
    error: false,
    warning: false,
    value: undefined,
    id: undefined,
    inputWidth: 'm',
    placeholder: undefined,
    menuPlacement: 'auto',
    hasAutosuggest: false,
    loadOptions: () => null,
    showOptionIcon: false,
    customComponents: {},
  }

  constructor(props) {
    super(props)

    let value
    if ('value' in props && this.context) {
      value = props.value
    }

    this.state = {
      checked: value,
    }
  }

  static getDerivedStateFromProps(props, state) {
    const { value } = props

    if ('value' in props && value !== state.value) {
      return { value }
    }

    return null
  }

  @bind
  async onChange(value) {
    const { onChange, hasAutosuggest } = this.props

    if (!('value' in this.props)) {
      // The Autosuggest Select (AsyncSelect) relies on the whole object to decode the selected option.
      this.setState({ value: hasAutosuggest ? value : value?.value })
    }

    onChange(hasAutosuggest ? value : value?.value, true)
  }

  @bind
  onBlur(event) {
    const { value } = this.state
    const { onBlur, name } = this.props

    // https://github.com/JedWatson/react-select/issues/3523
    // Make sure the input name is always present in the event object.
    event.target.name = name

    if (typeof value !== 'undefined') {
      // https://github.com/JedWatson/react-select/issues/3175
      event.target.value = value
    }

    onBlur(event)
  }

  @bind
  loadOptions(inputValue, callback) {
    const { loadOptions } = this.props
    const requestResults = loadOptions(inputValue)

    callback(requestResults)
  }

  @bind
  async handleRemoveSelected(e, option) {
    const newValue = this.state.value.filter(o => o.value !== option.value)

    this.onChange(newValue)
  }

  render() {
    const {
      className,
      options,
      inputWidth,
      intl,
      value,
      onChange,
      onBlur,
      onFocus,
      disabled,
      error,
      warning,
      name,
      id,
      placeholder,
      hasAutosuggest,
      showOptionIcon,
      customComponents,
      ...rest
    } = this.props

    const formatMessage = (label, values) => (intl ? intl.formatMessage(label, values) : label)
    const cls = classnames(className, style.container, style[`input-width-${inputWidth}`], {
      [style.error]: error,
      [style.warning]: warning,
    })
    const selectedOptionsClasses = classnames(
      style.container,
      style.selectedOptionsContainer,
      style[`input-width-${inputWidth}`],
      'mt-cs-xs',
    )

    const translatedOptions = options.map(option => {
      const { label, labelValues = {} } = option
      if (typeof label === 'object' && label.id && label.defaultMessage) {
        return { ...option, label: formatMessage(label, labelValues) }
      }

      return option
    })

    const customOption = props => (
      <components.Option {...props}>
        {showOptionIcon && <Icon icon={props.data.icon} className="mr-cs-xs" />}
        <b>{props.label}</b>
      </components.Option>
    )

    return hasAutosuggest ? (
      <>
        <AsyncSelect
          loadOptions={this.loadOptions}
          className={cls}
          inputId={id}
          classNamePrefix="select"
          onChange={this.onChange}
          onBlur={this.onBlur}
          onFocus={onFocus}
          isDisabled={disabled}
          value={value}
          name={name}
          components={{ Input, Option: customOption, ...customComponents }}
          aria-describedby={rest['aria-describedby']}
          placeholder={Boolean(placeholder) ? formatMessage(placeholder) : undefined}
          {...rest}
        />
        {rest.isMulti &&
          this.state.value?.map(option => (
            <div key={option.value} className={selectedOptionsClasses}>
              <Icon icon={option.icon} className="mr-cs-s" />
              <Message content={option.description ?? option.label} />
              <Button
                type="button"
                naked
                message={m.remove}
                value={option}
                onClick={this.handleRemoveSelected}
                className={style.removeOptionButton}
              />
            </div>
          ))}
      </>
    ) : (
      <ReactSelect
        className={cls}
        inputId={id}
        classNamePrefix="select"
        value={getValue(translatedOptions, value) || null}
        options={translatedOptions}
        onChange={this.onChange}
        onBlur={this.onBlur}
        onFocus={onFocus}
        isDisabled={disabled}
        name={name}
        components={{ Input }}
        aria-describedby={rest['aria-describedby']}
        placeholder={Boolean(placeholder) ? formatMessage(placeholder) : undefined}
        {...rest}
      />
    )
  }
}

export default injectIntl(Select)
export { Select }
