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

import { connect as withConnect } from 'react-redux'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import pipe from '@ttn-lw/lib/pipe'

import {
  checkFromState,
  mayViewApplicationLink,
  maySetApplicationPayloadFormatters,
} from '@console/lib/feature-checks'

import { getApplicationLink } from '@console/store/actions/link'

import {
  selectApplicationLink,
  selectApplicationLinkFetching,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'

const mapStateToProps = state => {
  const link = selectApplicationLink(state)
  const fetching = selectApplicationLinkFetching(state)
  const mayViewLink = checkFromState(mayViewApplicationLink, state)

  return {
    appId: selectSelectedApplicationId(state),
    fetching: (fetching || !link) && mayViewLink,
    mayViewLink,
  }
}

const mapDispatchToProps = dispatch => ({
  getLink: (id, selector) => dispatch(getApplicationLink(id, selector)),
})

const addHocs = pipe(
  withConnect(mapStateToProps, mapDispatchToProps),
  withFeatureRequirement(maySetApplicationPayloadFormatters, {
    redirect: ({ appId }) => `/applications/${appId}`,
  }),
)

export default ApplicationPayloadFormatters => addHocs(ApplicationPayloadFormatters)
