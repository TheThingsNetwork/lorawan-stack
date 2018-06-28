// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package testdata

// ExpectedAppLocal contains the expected value of the executed template using a file system.
const ExpectedAppLocal = `<!doctype html>
<html>
	<head>
		<title>Test App</title>
		<meta name="viewport" content="width=device-width,initial-scale=1,minimum-scale=1">
	</head>
	<body>
		<div id="app"></div>
		<script type="text/javascript" src="test/test.123.js"></script>
	</body>
</html>
`

// ExpectedAppCDN contains the expected value of the executed template using a file system.
const ExpectedAppCDN = `<!doctype html>
<html>
	<head>
		<title>Test App</title>
		<meta name="viewport" content="width=device-width,initial-scale=1,minimum-scale=1">
	</head>
	<body>
		<div id="app"></div>
		<script type="text/javascript" src="https://cdn.thethings.network/test.123.js"></script>
	</body>
</html>
`

// ExpectedErrorTemplated contains the expected value of the error handler using a template.
const ExpectedErrorTemplated = `<!doctype html>
<html>
	<head>
		<title>Internal Server Error</title>
		<meta name="viewport" content="width=device-width,initial-scale=1,minimum-scale=1">
	</head>
	<body>
		error:pkg/assets_test:test (Test error)
	</body>
</html>
`
