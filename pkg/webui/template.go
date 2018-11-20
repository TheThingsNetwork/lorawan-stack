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

package webui

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/errors"
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

    <meta property="og:url" content="{{.}}">
    <meta property="og:site_name" content="{{.SiteName}}{{with .Title}} {{.}}{{end}}">
    {{with .SubTitle}}<meta property="og:title" content="{{.}}">{{end}}
    {{with .Description}}<meta property="og:description" content="{{.}}">{{end}}

    <meta property="og:image" content="{{$assetsBaseURL}}/{{.IconPrefix}}og-image.png">
    <meta property="og:image:secure_url" content="{{$assetsBaseURL}}/{{.IconPrefix}}og-image.png">
    <meta property="og:image:width" content="1200">
    <meta property="og:image:height" content="630">

    <link rel="shortcut icon" type="image/x-icon" href="{{$assetsBaseURL}}/{{.IconPrefix}}favicon.ico">
    <link rel="apple-touch-icon" sizes="180x180" href="{{$assetsBaseURL}}/{{.IconPrefix}}touch-icon.png">

    <link href="https://fonts.gstatic.com/s/materialicons/v38/flUhRq6tzZclQEJ-Vdg-IuiaDsNcIhQ8tQ.woff2" rel="preload" as="font" type="font/woff2" crossorigin>
    {{range .CSSFiles}}<link href="{{$assetsBaseURL}}/{{.}}" rel="stylesheet">{{end}}
  </head>
  <body>
    <div id="app"></div>
    <script>
      window.APP_ROOT={{.MountPath}};
      window.ASSETS_ROOT={{$assetsBaseURL}};
      window.APP_CONFIG={{.AppConfig}};
      {{with .PageData}}window.PAGE_DATA={{.}};{{end}}
    </script>
    {{range .JSFiles}}<script type="text/javascript" src="{{$assetsBaseURL}}/{{.}}"></script>{{end}}
  </body>
</html>
`

// Template for rendering the Web UI.
// The context is expected to contain TemplateData as "template_data".
// The "app_config" will be rendered into the environment.
var Template echo.Renderer

func init() {
	appHTML := appHTML
	appHTMLLines := strings.Split(appHTML, "\n")
	for i, line := range appHTMLLines {
		appHTMLLines[i] = strings.TrimSpace(line)
	}
	Template = &appTemplate{
		template: template.Must(template.New("app").Parse(strings.Join(appHTMLLines, ""))),
	}
}

type appTemplate struct {
	template *template.Template
}

var hashedFiles = map[string]string{}

func registerHashedFile(original, hashed string) {
	hashedFiles[original] = hashed
}

func (t *appTemplate) Render(w io.Writer, _ string, pageData interface{}, c echo.Context) error {
	templateData := c.Get("template_data").(TemplateData)
	for i, cssFile := range templateData.CSSFiles {
		if hashedFile, ok := hashedFiles[cssFile]; ok {
			templateData.CSSFiles[i] = hashedFile
		}
	}
	for i, jsFile := range templateData.JSFiles {
		if hashedFile, ok := hashedFiles[jsFile]; ok {
			templateData.JSFiles[i] = hashedFile
		}
	}
	return t.template.Execute(w, Data{
		TemplateData: templateData,
		AppConfig:    c.Get("app_config"),
		PageData:     pageData,
	})
}

// Render the WebUI.
// The context is expected to contain TemplateData as "template_data".
// The "app_config" and "page_data" will be rendered into the environment.
func Render(c echo.Context) error {
	buf := new(bytes.Buffer)
	if err := Template.Render(buf, "", c.Get("page_data"), c); err != nil {
		return err
	}
	return c.HTMLBlob(http.StatusOK, buf.Bytes())
}

// RenderErrors renders errors into the WebUI if they occur.
func RenderErrors(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err == nil || c.Response().Committed {
			return err
		}
		c.Set("app_config", map[string]interface{}{
			"error": true,
		})

		status := http.StatusInternalServerError
		if echoErr, ok := err.(*echo.HTTPError); ok {
			status = echoErr.Code
			if ttnErr, ok := errors.From(echoErr.Internal); ok {
				if status == http.StatusInternalServerError {
					status = errors.HTTPStatusCode(ttnErr)
				}
				err = ttnErr
			}
		} else if ttnErr, ok := errors.From(err); ok {
			status = errors.HTTPStatusCode(ttnErr)
			err = ttnErr
		}

		if strings.Contains(c.Request().Header.Get("accept"), "application/json") {
			return c.JSON(status, err)
		}

		buf := new(bytes.Buffer)
		if err := Template.Render(buf, "", map[string]interface{}{
			"error": err,
		}, c); err != nil {
			return err
		}
		return c.HTMLBlob(status, buf.Bytes())
	}
}
