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
import { defineMessages } from 'react-intl'
import { connect as formConnect, getIn } from 'formik'
import bind from 'autobind-decorator'

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import Field from '../../../components/field'

import {
  gsFrequencyPlansSelector,
  nsFrequencyPlansSelector,
  errorSelector,
  fetchingSelector,
} from '../../store/selectors/configuration'

import {
  getGsFrequencyPlans,
  getNsFrequencyPlans,
} from '../../store/actions/configuration'

const m = defineMessages({
  freqPlansFetchingFailure: 'Could not retrieve the list of available frequency plans',
})

const formatOptions = plans => plans.map(plan => ({ value: plan.id, label: plan.name }))

@formConnect
@storeConnect(function (state, props) {
  return {
    plansGs: gsFrequencyPlansSelector(state, props),
    plansNs: nsFrequencyPlansSelector(state, props),
    error: errorSelector(state, props),
    fetching: fetchingSelector(state, props),
  }
},
dispatch => ({
  getGsFrequencyPlans: () => dispatch(getGsFrequencyPlans()),
  getNsFrequencyPlans: () => dispatch(getNsFrequencyPlans()),
}))
@bind
class FrequencyPlansSelect extends React.PureComponent {

  constructor (props) {
    super(props)

    const { name, formik } = props

    formik.registerField(name, this)
  }

  componentDidMount () {
    const {
      source,
      getNsFrequencyPlans,
      getGsFrequencyPlans,
    } = this.props

    if (source === 'ns') {
      getNsFrequencyPlans()
    } else {
      getGsFrequencyPlans()
    }
  }

  componentWillUnmount () {
    const { formik, name } = this.props

    formik.unregisterField(name)
  }

  getOptions () {
    const { source, plansGs, plansNs } = this.props

    return source === 'ns'
      ? formatOptions(plansNs)
      : formatOptions(plansGs)
  }

  render () {
    const {
      name,
      required,
      title,
      autoFocus,
      horizontal,
      error,
      fetching,
      menuPlacement,
    } = this.props

    const {
      setFieldValue,
      setFieldTouched,
    } = this.props.formik

    const fieldOptions = this.getOptions()
    const fieldError = getIn(this.props.formik.errors, name)
    const fieldTouched = getIn(this.props.formik.touched, name)

    return (
      <Field
        horizontal={horizontal}
        type="select"
        options={fieldOptions}
        name={name}
        required={required}
        title={title}
        autoFocus={autoFocus}
        isLoading={fetching}
        warning={Boolean(error) ? m.freqPlansFetchingFailure : undefined}
        error={fieldError}
        touched={fieldTouched}
        menuPlacement={menuPlacement}
        setFieldTouched={setFieldTouched}
        setFieldValue={setFieldValue}
      />
    )
  }
}

FrequencyPlansSelect.propTypes = {
  source: PropTypes.oneOf([ 'ns', 'gs' ]).isRequired,
  name: PropTypes.string.isRequired,
  required: PropTypes.bool.isRequired,
  title: PropTypes.message,
  autoFocus: PropTypes.bool,
  horizontal: PropTypes.bool,
  menuPlacement: PropTypes.oneOf([ 'top', 'bottom', 'auto' ]),
}

FrequencyPlansSelect.defaultProps = {
  title: sharedMessages.frequencyPlan,
  autoFocus: false,
  horizontal: false,
  menuPlacement: 'auto',
}

export default FrequencyPlansSelect
