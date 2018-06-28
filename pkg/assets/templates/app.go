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

package templates

import "html/template"

// AppData contains data to use in the App template.
type AppData struct {
	Title    string `name:"title" description:"The page title."`
	FileName string `name:"name" description:"The file name."`
}

const appHTML string = `<!doctype html>
<html>
	<head>
		<title>{{.Data.Title}}</title>
		<meta name="viewport" content="width=device-width,initial-scale=1,minimum-scale=1">
	</head>
	<body>
		<div id="app"></div>
		<script type="text/javascript" src="{{.Root}}/{{.Data.FileName}}"></script>
	</body>
</html>
`

// App is a template for rendering a JavaScript application using AppData as data.
var App = template.Must(template.New("app").Parse(appHTML))
