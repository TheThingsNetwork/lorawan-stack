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

export default defineMessages({
  // Keep these sorted alphabetically.
  technicalDetails: 'Technical details',
  attachToSupportInquiries: 'Please attach the technical details below to support inquiries',
  contactAdministrator: 'If the error persists, please contact an administrator.',
  contactSupport: 'If the error persists, please contact support.',
  inconvenience: "We're sorry for the inconvenience.",
  error: 'Error',
  genericError: 'An unknown error occurred. Please try again later.',
  genericNotFound: 'The page you requested cannot be found',
  subviewErrorExplanation: 'There was a problem when displaying this section',
  subviewErrorTitle: "We're sorry!",
  unknownErrorTitle: 'An unknown error occurred',
  errorOccurred: 'An error occurred and the request could not be completed.',
  errorId: 'Error ID: <code>{errorId}</code>',
  correlationId: 'Correlation ID: <code>{correlationId}</code>',
  loginFailed: 'Login failed',
  loginFailedDescription:
    'There was an error causing the login to fail. This might be due to server-side misconfiguration or a browser-cookie problem. Please try logging in again.',
  loginFailedAbortDescription:
    'The login process was aborted during the authentication with the login provider. You can use the button below to retry logging in.',
  connectionFailure:
    'The servers are currently unavailable. This might be due to a scheduled update or maintenance. Please check the status page (if available) for more information and try again later.',
})
