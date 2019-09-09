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

package webui

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"

	echo "github.com/labstack/echo/v4"
)

// Data contains data to render templates.
type Data struct {
	TemplateData
	AppConfig interface{}
	PageData  interface{}
}

// TemplateData contains data to use in the App template.
type TemplateData struct {
	SiteName      string   `name:"site-name" description:"The site name"`
	Title         string   `name:"title" description:"The page title"`
	SubTitle      string   `name:"sub-title" description:"The page sub-title"`
	Description   string   `name:"descriptions" description:"The page description"`
	Language      string   `name:"language" description:"The page language"`
	ThemeColor    string   `name:"theme-color" description:"The page theme color"`
	CanonicalURL  string   `name:"canonical-url" description:"The page canonical URL"`
	AssetsBaseURL string   `name:"assets-base-url" description:"The base URL to the page assets"`
	IconPrefix    string   `name:"icon-prefix" description:"The prefix to put before the page icons (favicon.ico, touch-icon.png, og-image.png)"`
	CSSFiles      []string `name:"css-file" description:"The names of the CSS files"`
	JSFiles       []string `name:"js-file" description:"The names of the JS files"`
}

// MountPath derives the mount path from the canonical URL of the config.
func (t TemplateData) MountPath() string {
	if url, err := url.Parse(t.CanonicalURL); err == nil {
		if url.Path == "" {
			return "/"
		}
		return url.Path
	}
	return ""
}

const appHTML = `{{- $assetsBaseURL := .AssetsBaseURL -}}
<!doctype html>
<html lang="{{with .Language}}{{.}}{{else}}en{{end}}">
  <head>
    <title>{{.SiteName}}{{with .Title}} {{.}}{{end}}</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, minimum-scale=1">
    <meta name="theme-color" content="{{with .ThemeColor}}{{.}}{{else}}#0D83D0{{end}}">
    <meta http-equiv="X-UA-Compatible" content="IE=edge" >
    {{with .Description}}<meta name="description" content="{{.}}">{{end}}

    <meta property="og:url" content="{{.CanonicalURL}}">
    <meta property="og:site_name" content="{{.SiteName}}{{with .Title}} {{.}}{{end}}">
    {{with .SubTitle}}<meta property="og:title" content="{{.}}">{{end}}
    {{with .Description}}<meta property="og:description" content="{{.}}">{{end}}

    <meta property="og:image" content="{{$assetsBaseURL}}/{{.IconPrefix}}og-image.png">
    <meta property="og:image:secure_url" content="{{$assetsBaseURL}}/{{.IconPrefix}}og-image.png">
    <meta property="og:image:width" content="1200">
    <meta property="og:image:height" content="630">

    <link rel="shortcut icon" type="image/x-icon" href="{{$assetsBaseURL}}/{{.IconPrefix}}favicon.ico">
    <link rel="apple-touch-icon" sizes="180x180" href="{{$assetsBaseURL}}/{{.IconPrefix}}touch-icon.png">

    {{range .CSSFiles}}<link href="{{$assetsBaseURL}}/{{.}}" rel="stylesheet">{{end}}
  </head>
  <body>
    <div id="app"></div>
    <script>
      window.APP_ROOT={{.MountPath}};
      window.ASSETS_ROOT={{$assetsBaseURL}};
      window.APP_CONFIG={{.AppConfig}};
      window.SITE_NAME={{.SiteName}};
      window.SITE_TITLE={{.Title}};
      window.SITE_SUB_TITLE={{.SubTitle}};
      {{with .PageData}}window.PAGE_DATA={{.}};{{end}}
    </script>
    {{range .JSFiles}}<script type="text/javascript" src="{{$assetsBaseURL}}/{{.}}"></script>{{end}}
  </body>
</html>
`

// Template for rendering the web UI.
// The context is expected to contain TemplateData as "template_data".
// The "app_config" will be rendered into the environment.
var Template *AppTemplate

func init() {
	appHTML := appHTML
	appHTMLLines := strings.Split(appHTML, "\n")
	for i, line := range appHTMLLines {
		appHTMLLines[i] = strings.TrimSpace(line)
	}
	Template = NewAppTemplate(template.Must(template.New("app").Parse(strings.Join(appHTMLLines, ""))))
}

// AppTemplate wraps the application template for the web UI.
type AppTemplate struct {
	template *template.Template
}

// NewAppTemplate instantiates a new application template for the web UI.
func NewAppTemplate(t *template.Template) *AppTemplate {
	return &AppTemplate{template: t}
}

var hashedFiles = map[string]string{}

func registerHashedFile(original, hashed string) {
	hashedFiles[original] = hashed
}

// Render is the echo.Renderer that renders the web UI.
func (t *AppTemplate) Render(w io.Writer, _ string, pageData interface{}, c echo.Context) error {
	templateData := c.Get("template_data").(TemplateData)
	cssFiles := make([]string, len(templateData.CSSFiles))
	for i, cssFile := range templateData.CSSFiles {
		if hashedFile, ok := hashedFiles[cssFile]; ok {
			cssFiles[i] = hashedFile
		} else {
			cssFiles[i] = cssFile
		}
	}
	templateData.CSSFiles = cssFiles
	jsFiles := make([]string, len(templateData.JSFiles))
	for i, jsFile := range templateData.JSFiles {
		if hashedFile, ok := hashedFiles[jsFile]; ok {
			jsFiles[i] = hashedFile
		} else {
			jsFiles[i] = jsFile
		}
	}
	templateData.JSFiles = jsFiles
	return t.template.Execute(w, Data{
		TemplateData: templateData,
		AppConfig:    c.Get("app_config"),
		PageData:     pageData,
	})
}

// Handler is the echo.HandlerFunc that renders the web UI.
// The context is expected to contain TemplateData as "template_data".
// The "app_config" and "page_data" will be rendered into the environment.
func (t *AppTemplate) Handler(c echo.Context) error {
	buf := new(bytes.Buffer)
	if err := Template.Render(buf, "", c.Get("page_data"), c); err != nil {
		return err
	}
	return c.HTMLBlob(http.StatusOK, buf.Bytes())
}

// RenderError implements web.ErrorRenderer.
func (t *AppTemplate) RenderError(c echo.Context, statusCode int, err error) error {
	buf := new(bytes.Buffer)
	if err := Template.Render(buf, "", map[string]interface{}{"error": err}, c); err != nil {
		return err
	}
	return c.HTMLBlob(statusCode, buf.Bytes())
}
