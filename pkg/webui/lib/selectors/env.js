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

export const configSelector = () => window.__ttn_config__

export const selectApplicationRootPath = () => configSelector().APP_ROOT

export const selectAssetsRootPath = () => configSelector().ASSETS_ROOT

export const selectBrandingRootPath = () => configSelector().BRANDING_ROOT

export const selectApplicationConfig = () => configSelector().APP_CONFIG

export const selectApplicationSiteName = () => configSelector().SITE_NAME

export const selectApplicationSiteTitle = () => configSelector().SITE_TITLE

export const selectApplicationSiteSubTitle = () => configSelector().SITE_SUB_TITLE

export const selectSentryDsnConfig = () => configSelector().SENTRY_DSN

export const selectDevEUIConfig = () => ({
  devEUIIssuingEnabled: selectApplicationConfig().dev_eui_issuing_enabled,
  applicationLimit: selectApplicationConfig().dev_eui_application_limit,
})

export const selectCSRFToken = () => configSelector().CSRF_TOKEN

export const selectStackConfig = () => selectApplicationConfig().stack_config

export const selectGsConfig = () => selectStackConfig().gs
export const selectGsEnabled = () => selectGsConfig().enabled

export const selectIsConfig = () => selectStackConfig().is
export const selectIsEnabled = () => selectIsConfig().enabled

export const selectNsConfig = () => selectStackConfig().ns
export const selectNsEnabled = () => selectNsConfig().enabled

export const selectJsConfig = () => selectStackConfig().js
export const selectJsEnabled = () => selectJsConfig().enabled

export const selectAsConfig = () => selectStackConfig().as
export const selectAsEnabled = () => selectAsConfig().enabled

export const selectGcsConfig = () => selectStackConfig().gcs
export const selectGcsEnabled = () => selectGcsConfig().enabled

export const selectLanguageConfig = () => selectApplicationConfig().language

export const selectSupportLinkConfig = () => selectApplicationConfig().support_link

export const selectDocumentationUrlConfig = () => selectApplicationConfig().documentation_base_url

export const selectPageStatusBaseUrlConfig = () => selectApplicationConfig().status_page_base_url

export const selectEnableUserRegistration = () => selectApplicationConfig().enable_user_registration

export const selectPageData = () => configSelector().PAGE_DATA
