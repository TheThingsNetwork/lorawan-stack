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

import withRequest from '@ttn-lw/lib/components/with-request'

import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'
import pipe from '@ttn-lw/lib/pipe'

import {
  getApplication,
  stopApplicationEventsStream,
  getApplicationsRightsList,
} from '@console/store/actions/applications'

import {
  selectSelectedApplication,
  selectApplicationFetching,
  selectApplicationError,
  selectApplicationRights,
  selectApplicationRightsFetching,
  selectApplicationRightsError,
} from '@console/store/selectors/applications'

const mapStateToProps = (state, props) => ({
  appId: props.match.params.appId,
  fetching: selectApplicationFetching(state) || selectApplicationRightsFetching(state),
  application: selectSelectedApplication(state),
  error: selectApplicationError(state) || selectApplicationRightsError(state),
  rights: selectApplicationRights(state),
  siteName: selectApplicationSiteName(),
})

const mapDispatchToProps = dispatch => ({
  stopStream: id => dispatch(stopApplicationEventsStream(id)),
  loadData: id => {
    dispatch(getApplication(id, 'name,description,attributes,dev_eui_counter'))
    dispatch(getApplicationsRightsList(id))
  },
})

const addHocs = pipe(
  withConnect(mapStateToProps, mapDispatchToProps),
  withRequest(
    ({ appId, loadData }) => loadData(appId),
    ({ fetching, application }) => fetching || !Boolean(application),
  ),
)

export default Application => addHocs(Application)
