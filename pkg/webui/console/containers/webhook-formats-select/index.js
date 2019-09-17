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

import { defineMessages } from 'react-intl'

import CreateFetchSelect from '../fetch-select'
import {
  selectWebhookFormats,
  selectWebhookFormatsError,
  selectWebhookFormatsFetching,
} from '../../store/selectors/webhook-formats'
import { getWebhookFormats } from '../../store/actions/webhook-formats'

const m = defineMessages({
  title: 'Webhook Format',
  warning: 'Could not retrieve the list of available webhook formats',
})

export default CreateFetchSelect({
  fetchOptions: getWebhookFormats,
  optionsSelector: selectWebhookFormats,
  errorSelector: selectWebhookFormatsError,
  fetchingSelector: selectWebhookFormatsFetching,
  defaultWarning: m.warning,
  defaultTitle: m.title,
})
