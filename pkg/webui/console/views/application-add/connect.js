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
import { push } from 'connected-react-router'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import pipe from '@ttn-lw/lib/pipe'

import { mayCreateApplications } from '@console/lib/feature-checks'

import { selectUserId, selectUserRights } from '@console/store/selectors/user'

const mapStateToProps = state => ({
  userId: selectUserId(state),
  rights: selectUserRights(state),
})

const mapDispatchToProps = dispatch => ({
  navigateToApplication: appId => dispatch(push(`/applications/${appId}`)),
})

const addHocs = pipe(
  withFeatureRequirement(mayCreateApplications, { redirect: '/applications' }),
  withConnect(mapStateToProps, mapDispatchToProps),
)

export default ApplicationAdd => addHocs(ApplicationAdd)
