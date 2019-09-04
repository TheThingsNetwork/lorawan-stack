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

import { connect } from 'react-redux'

import {
  selectSelectedOrganization,
  selectOrganizationFetching,
  selectOrganizationError,
} from '../../store/selectors/organizations'
import { getOrganization, stopOrganizationEventsStream } from '../../store/actions/organizations'

import withRequest from '../../../lib/components/with-request'

const mapStateToProps = (state, props) => ({
  orgId: props.match.params.orgId,
  fetching: selectOrganizationFetching(state),
  organization: selectSelectedOrganization(state),
  error: selectOrganizationError(state),
})

const mapDispatchToProps = dispatch => ({
  stopStream: id => dispatch(stopOrganizationEventsStream(id)),
  getOrganization: id => dispatch(getOrganization(id, 'name,description')),
})

export default Organization =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
  )(
    withRequest(
      ({ orgId, getOrganization }) => getOrganization(orgId),
      ({ fetching, organization }) => fetching || !Boolean(organization),
    )(Organization),
  )
