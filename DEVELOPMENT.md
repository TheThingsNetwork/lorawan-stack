# The Things Stack for LoRaWAN Development

The Things Stack components are primarily built in Go, we use React for web front-ends. It is assumed that you have decent knowledge and experience with these technologies. If you want to get more familiar with Go, we strongly recommend to take [A Tour of Go](https://tour.golang.org/).

## Table of contents

- [Development Environment](#development-environment)
- [Cloning the Repository](#cloning-the-repository)
- [Getting Started](#getting-started)
- [Running a development build of The Things Stack](#running-a-development-build-of-the-things-stack)
  - [Pre-requisites](#pre-requisites)
  - [Steps](#steps)
- [Using the CLI with the Development Environment](#using-the-cli-with-the-development-environment)
- [Managing the Development Databases](#managing-the-development-databases)
  - [PostgreSQL](#PostgreSQL)
  - [Redis](#redis)
- [Building the Frontend](#building-the-frontend)
- [Starting The Things Stack](#starting-the-things-stack)
- [Project Structure](#project-structure)
  - [API](#api)
  - [Documentation](#documentation)
  - [Web UI](#web-ui)
- [Code Style](#code-style)
  - [Code Formatting](#code-formatting)
  - [Line Length](#line-length)
  - [Formatting and Linting](#formatting-and-linting)
  - [Documentation Site](#documentation-site)
- [Naming Guidelines](#naming-guidelines)
  - [API Method Naming](#api-method-naming)
  - [Variable Naming](#variable-naming)
  - [Event Naming](#event-naming)
  - [Error Naming](#error-naming)
  - [Log Field Keys, Event Names, Error Names, Error Attributes and Task Identifiers](#log-field-keys-event-names-error-names-error-attributes-and-task-identifiers)
  - [Comments](#comments)
- [JavaScript Code Style](#javascript-code-style)
  - [Code Formatting](#code-formatting)
  - [Code Comments](#code-comments)
  - [Import Statement Order](#import-statement-order)
  - [React Component Syntax (Functional, Class Components and Hooks)](#react-component-syntax-functional-class-components-and-hooks)
  - [React Component Types](#react-component-types)
  - [Frontend Related Pull Requests](#frontend-related-pull-requests)
- [Translations](#translations)
  - [Backend Translations](#backend-translations)
  - [Frontend Translations](#frontend-translations)
- [Events](#events)
- [Testing](#testing)
  - [Unit Tests](#unit-tests)
  - [End-to-end Tests](#end-to-end-tests)
- [Building and Running](#building-and-running)
- [Releasing](#releasing)
  - [Release From Master](#release-from-master)
  - [Release Backports](#release-backports)
- [Troubleshooting](#troubleshooting)
  - [Console](#console)

## Development Environment

The Things Network's development tooling uses [Mage](https://magefile.org/). Under the hood, `mage` calls other tools such as `git`, `go`, `yarn`, `docker` etc. Recent versions are supported; Node v18.x and Go v1.18.x.

- Follow [Go's installation guide](https://golang.org/doc/install) to install Go.
- Download Node.js [from their website](https://nodejs.org) and install it.
- Follow [Yarn's installation guide](https://yarnpkg.com/en/docs/install) to install Yarn.
- Follow the guides to [install Docker](https://docs.docker.com/install/#supported-platforms) and to [install Docker Compose](https://docs.docker.com/compose/install/#install-compose).

## Cloning the Repository

If you are unfamiliar with forking projects on GitHub or cloning them locally, please [see the GitHub documentation](https://help.github.com/articles/fork-a-repo/).

## Getting Started

The first step to get started is to initialize tooling and some dependencies:

```bash
$ make init
```

You may want to run this commands from time to time to stay up-to-date with changes to tooling and dependencies.

## Running a development build of The Things Stack

This section explains how to get a bare-bones version of The Things Stack running on your local machine. This will build whatever code is present in your local repository (along with local changes) and run in it using the default ports.

If you want to just run a docker image of The Things Stack, then check the [Installation](https://thethingsstack.io/getting-started/installation/) section of the documentation.

### Pre-requisites

1. This section requires that the required tools from [Development Environment](##Development-Environment) are installed.
2. This repository must be cloned inside the `GOPATH`. Check the [official documentation](https://golang.org/doc/gopath_code.html) on working with `GOPATH`.
3. Make sure that you've run `$ make init` before continuing.
4. If this is not the first time running the stack, make sure to clear any environment variables that you've been using earlier. You can do check what variables are set currently by using

```
$ printenv | grep "TTN_LW_*"
```

### Steps

1. Build the frontend assets

```bash
$ tools/bin/mage js:build
```

This will build the frontend assets and place it in the `public` folder.

2. Start the databases

```bash
$ tools/bin/mage dev:dbStart # This requires Docker to be running.
```

This will start one instance each of `postgres` and `Redis` as Docker containers. To verify this, you can run

```bash
$ docker ps
```

3. Initialize the database with defaults.

```bash
$ tools/bin/mage dev:initStack
```

This creates a database, migrates tables and creates a user `admin` with password `admin`.
- An API Key for the admin user with `RIGHTS_ALL` is also created and stored in `.env/admin_api_key.txt`.

4. Start a development instance of The Things Stack

```bash
$ go run ./cmd/ttn-lw-stack -c ./config/stack/ttn-lw-stack.yml start
```

5. Login to The Things Stack via the Console

In a web browser, navigate to `http://localhost:1885/` and login using credentials from step 3.

6. Customizing configuration

To customize the configuration, copy the configuration file `/config/stack/ttn-lw-stack.yml` to a different location (ex: the `.env` folder in your repo). The configuration is documented in the [Configuration Reference](https://thethingsstack.io/reference/configuration/).

You can now use the modified configuration with

```bash
$ go run ./cmd/ttn-lw-stack -c <custom-location>/ttn-lw-stack.yml start
```

## Using the CLI with the Development Environment

In order to login, you will need to use the correct OAuth Server Address. `make init` uses CFSSL to generate a `ca.pem` CA certificate to support https:

```bash
$ export TTN_LW_CA=./ca.pem
$ export TTN_LW_OAUTH_SERVER_ADDRESS=https://localhost:8885/oauth
$ go run ./cmd/ttn-lw-cli login
```

## Managing the Development Databases

You can use the following commands to start, stop and erase databases.

```bash
$ tools/bin/mage dev:dbStart # Starts all databases in a Docker container.
$ tools/bin/mage dev:dbStop  # Stops all databases.

# The contents of the databases will be saved in .env/data

$ tools/bin/mage dev:dbErase # Stops all databases and erase storage.
```

### PostgreSQL

PostgreSQL is a SQL database that we use in the Identity Server.

You can use `tools/bin/mage dev:dbSQL` to enter an SQL shell.

### Redis

Redis is an in-memory data store that we use as a database for "hot" data.

You can use `tools/bin/mage dev:dbRedisCli` to enter a Redis-CLI shell.

## Building the Frontend

You can use `tools/bin/mage js:build` to build the frontend.

## Starting The Things Stack

You can use `go run ./cmd/ttn-lw-stack start` to start The Things Stack.

#### Codec

Most data is stored as base64-encoded protocol buffers. For debugging purposes it is often useful to inspect or update the stored database models - you can use Redis codec tool located at `./pkg/redis/codec` to decode/encode them to/from JSON.

##### Examples

###### Get and Decode

```bash
$ redis-cli get "ttn:v3:ns:devices:uid:test-app:test-dev" | go run ./pkg/redis/codec -type 'ttnpb.EndDevice'
```

###### Get, Decode, Modify, Encode and Set

```
$ redis-cli get "ttn:v3:ns:devices:uid:test-app.test-dev" \
  | go run ./pkg/redis/codec -type 'ttnpb.EndDevice' \
  | jq '.supports_join = false' \
  | go run ./pkg/redis/codec -type 'ttnpb.EndDevice' -encode \
  | redis-cli -x set "ttn:v3:ns:devices:uid:test-app.test-dev"
```

## Project Structure

The folder structure of the project looks as follows:

```
.
├── .editorconfig       configuration for your editor, see editorconfig.org
├── CODEOWNERS          maintainers of folders who are required to approve pull requests
├── CONTRIBUTING.md     guidelines for contributing: branching, commits, code style, etc.
├── DEVELOPMENT.md      guide for setting up your development environment
├── docker-compose.yml  deployment file (including databases) for Docker Compose
├── Dockerfile          formula for building Docker images
├── LICENSE             the license that explains what you're allowed to do with this code
├── Makefile            dev/test/build tooling
├── tools               dev/test/build tooling
├── README.md           general information about this project
│   ...
├── api                 contains the protocol buffer definitions for our API
├── cmd                 contains the different binaries that form The Things Stack for LoRaWAN
│   ├── internal        contains internal files shared between the different binaries
│   │   ...
│   ├── ttn-lw-cli      the command-line-interface for The Things Stack for LoRaWAN
│   └── ttn-lw-stack    bundles the server binaries that form The Things Stack for LoRaWAN
├── config              configuration for our JavaScript SDK and frontend
├── data                data from external repositories, such as devices, frequency plans and webhook templates
├── doc                 detailed documentation on the workings of The Things Stack for LoRaWAN
├── pkg                 contains all libraries used in The Things Stack for LoRaWAN
│   ├── component       contains the base component; all other components extend this component
│   ├── config          package for configuration using config files, environment and CLI flags
│   ├── console         package that provides the web server for the console
│   ├── errors          package for rich errors that include metadata and cross API boundaries
│   ├── log             package for logging
│   ├── messages        contains non-proto messages (such as the messages that are sent over MQTT)
│   ├── metrics         package for metrics collection
│   ├── ttnpb           contains generated code from our protocol buffer definitions and some helper functions
│   ├── types           contains primitive types
│   ├── webui           contains js code for the console and oauth provider
│   └── ...
├── public              frontend code will be compiled to this folder - not added to git
├── release             binaries will be compiled to this folder - not added to git
└── sdk                 source code for our SDKs
    └── js              source code for our JavaScript SDK
```

### API

Our APIs are defined in `.proto` files in the `api` folder. These files describe the messages and interfaces of the different components of The Things Stack. If this is the first time you hear the term "protocol buffers" you should probably read the [protocol buffers documentation](https://developers.google.com/protocol-buffers/docs/proto3) before you continue.

From the `.proto` files, we generate code using the `protoc` compiler. As we plan to compile to a number of different languages, we decided to put the compiler and its dependencies in a [Docker image](https://github.com/TheThingsIndustries/docker-protobuf). The actual commands for compilation are handled by our tooling, so the only thing you have to execute when updating the API, is:

```bash
$ tools/bin/mage proto:clean proto:all jsSDK:definitions
```

### Documentation

The documentation site for The Things Stack is built from the [`lorawan-stack-docs`](https://github.com/TheThingsIndustries/lorawan-stack-docs) repository.

### Web UI

The Things Stack for LoRaWAN includes two frontend applications: the **Console** and **Account App**. Both applications use [React](https://reactjs.org/) as frontend framework. The `console` and `account` packages of the backend expose their respective web servers and handle all logic that cannot be done in the browser. Otherwise both applications are single page applications (SPA) that run entirely in the browser.

The folder structure of the frontend looks as follows:

```
./pkg/webui
├── assets            assets (eg. vectors, images) used by the frontend
├── components        react components shared throughout the frontend
├── console           root of the console application
│   ├── api           api definitions to communicate with The Things Stack
│   ├── containers    container components
│   ├── lib           utility classes and functions
│   ├── store         redux actions, reducers and logic middlewares
│   ├── views         whole view components of the console (~pages)
├── containers        global react container components
├── lib               global utility classes and functions
├── locales           frontend and backend locale jsons used for i18n
├── oauth             root of the oauth application
│   ├── api           api definitions to communicate with The Things Stack
│   ├── store         redux actions, reducers and logic middlewares
│   ├── views         whole view components of the oauth provider (~pages)
├── styles            global stylus (~css) styles and mixins
├── console.js        entry point of the console app
├── oauth.js          entry point of the oauth app
├── template.go       go template module used to render the frontend HTML
```

For development purposes, the frontend can be run using `webpack-dev-server`. After following the [Getting Started](#getting-started) section to initialize The Things Stack and doing an initial build of the frontend via `tools/bin/mage js:build`, it can be served using:

```bash
$ export NODE_ENV=development
$ tools/bin/mage js:serve
```

The development server runs on `http://localhost:8080` and will proxy all api calls to port `1885`. The serve command watches any changes inside `pkg/webui` and refreshes automatically.

#### Development Configuration

In order to set up The Things Stack to support running the frontend via `webpack-dev-server`, the following environment setup is needed:

```bash
# .dev.env
export NODE_ENV="development"
export TTN_LW_LOG_LEVEL="debug"
export TTN_LW_CONSOLE_UI_CANONICAL_URL="http://localhost:8080/console"
export TTN_LW_CONSOLE_OAUTH_AUTHORIZE_URL="http://localhost:8080/oauth/authorize"
export TTN_LW_CONSOLE_OAUTH_LOGOUT_URL="http://localhost:8080/oauth/logout"
export TTN_LW_CONSOLE_OAUTH_TOKEN_URL="http://localhost:8080/oauth/token"
export TTN_LW_IS_OAUTH_UI_CANONICAL_URL="http://localhost:8080/oauth"
export TTN_LW_IS_EMAIL_NETWORK_IDENTITY_SERVER_URL="http://localhost:8080/oauth"
export TTN_LW_IS_EMAIL_PROVIDER="dir"
export TTN_LW_IS_EMAIL_DIR=".dev/email"
export TTN_LW_CONSOLE_UI_ASSETS_BASE_URL="http://localhost:8080/assets"
export TTN_LW_IS_OAUTH_UI_CONSOLE_URL="http://localhost:8080/console"
export TTN_LW_CONSOLE_UI_ACCOUNT_URL="http://localhost:8080/oauth"
```

We recommend saving this configuration as an `.dev.env` file and sourcing it like `source .dev.env`. This allows you to easily apply development configuration when needed.

> Note: It is important to **source these environment variables in all terminal sessions** that run The Things Stack or the `tools/bin/mage` commands. Failing to do so will result in erros such as blank page renders. See also [troubleshooting](#troubleshooting).

#### Optional Configuration

##### Disable [Hot Module Replacement](https://webpack.js.org/concepts/hot-module-replacement/)

If you experience trouble seeing the WebUIs updated after a code change, you can also disable hot module replacement and enforce a hard reload on code changes instead. This method is a bit slower but more robust. To do so apply the following variable:

```bash
WEBPACK_DEV_SERVER_DISABLE_HMR="true"
```

> Note: Webpack-related configuration can be loaded from environment variables only. It cannot be sourced from a config file.

##### Enable TLS in `webpack-dev-server`

```bash
WEBPACK_DEV_SERVER_USE_TLS="true"
```
This option uses the key and certificate set via `TTN_LW_TLS_KEY` and `TTN_LW_TLS_CERTIFICATE` environment variables. Useful when developing functionalities that rely on TLS.

> Note: To use this option, The Things Stack for LoRaWAN must be properly setup for TLS. You can obtain more information about this in the **Getting Started** section of the The Things Stack for LoRaWAN documentation.

## Code Style

### Code Formatting

We want our code to be consistent across our projects, so we'll have to agree on a number of formatting rules. These rules should usually usually be applied by your editor. Make sure to install the [editorconfig](https://editorconfig.org) plugin for your editor.

Our editorconfig contains the following rules:

- We use the **utf-8** character set.
- We use **LF** line endings.
- We have a **final newline** in each file.
- We we **trim whitespace** at the end of each line (except in Markdown).
- All `.go` files are indented using **tabs**
- The `Makefile` and all `.make` files are indented using **tabs**
- All other files are indented using **two spaces**

### Line Length

We don't have strict rules for line length, but in our experience the following guidelines result in nice and readable code:

- If a line is longer than 80 columns, try to find a "natural" break
- If a line is longer than 120 columns, insert a line break
- In very special cases, longer lines are tolerated

### Formatting and Linting

Go code can be automatically formatted using tools such as [`gofmt`](https://godoc.org/github.com/golang/go/src/cmd/gofmt) and [`goimports`](https://godoc.org/golang.org/x/tools/cmd/goimports). The [Go language server](https://github.com/golang/tools/tree/master/gopls) can also help with formatting code. There are many editor plugins that automatically format your code when you save your files. We highly recommend using those.

We use [`golangci-lint`](https://golangci-lint.run/) to lint Go code and [`eslint`](https://eslint.org) to lint JavaScript code. These tools should automatically be installed when initializing your development environment.

### Documentation Site

Please respect the following guidelines for content in our documentation site. A copy and paste template for creating new documentation can be found [here](doc/content/example-template).

- Use the `{{< new-in-version "3.8.5" >}}` shortcode to tag documentation for features added in a particular version. For documentation that targets `v3.n`, that's the next patch bump, e.g `3.8.x`. For documentation targeting `v3.n+1` that's the next minor bump, e.g `3.9.0`.
- The title of a doc page is already rendered by the build system as a h1, don't add an extra one.
- Use title case for headings.
- A documentation page starts with an introduction, and then the first heading. The first paragraph of the introduction is typically a summary of the page. Use a `<!--more-->` to indicate where the summary ends.
- Divide long documents into separate files, each with its own folder and `_index.md`.
- Use the `weight`tag in the [Front Matter](https://gohugo.io/content-management/front-matter/) to manually sort sections if necessary. If not, they will be sorted alphabetically.
- Since the title is a `h1`, everything in the content is at least `h2` (`##`).
- Paragraphs typically consist of at least two sentences.
- Use an empty line between all blocks (headings, paragraphs, lists, ...).
- Prefer text over bullet lists or enumerations. For bullets, use `-`, for enumerations `1.` etc.
- Explicitly call this product "The Things Stack". Not "the stack" etc. You can use the shortcode `{{% tts %}}` which will expand to "The Things Stack".
- Avoid shortening, i.e. write "it is" instead of "it's".
- Write guides as a goal-oriented journey.
- Unless already clear from context, use a clearer term than "user", especially if there are multiple kinds (network operator, gateway owner, application developer, ...).
- The user does not have a gender (so use they/them/their).
- Taking screenshots is done as follows:
  - In Chrome: activate the **Developer Tools** and toggle the **Device Toolbar**. In the **Device Toolbar**, select **Laptop with HiDPI screen** (add it if not already there), and click **Capture Screenshot** in the menu on the right.
  - In Firefox: enter **Responsive Design Mode**. In the **Device Toolbar**, select "Laptop with HiDPI screen" (add it if not already there) and **Take a screenshot of the viewport**.
- Use `**Strong**` when referring to buttons in the Console.
- Use `>Note:`to add a note.
- Use fenced code blocks with a language:
  - `bash` for lists of environment variables: `SOME_ENV="value"`.
  - `bash` for CLI examples. Prefix commands with `$ `. Wrap strings with double quotes `""` (except when working with JSON, which already uses double quotes).
  - Wrap large CLI output with `<details><summary>Show CLI output</summary> ... output here ... </details>`.
  - `yaml` (not `yml`) for YAML. Wrap strings with single quotes `''` (because of frequent Go templates that use double quotes).

## Naming Guidelines

### API Method Naming

All API method names should follow the naming convention of `VerbNoun` in upper camel case, where the verb uses the imperative mood and the noun is the resource type.

The following snippet defines the basic CRUD definitions for a resource named `Type`.
Note also that the order of the methods is defined by CRUD.

```
CreateType
GetType
ListTypes (returns slice)
UpdateType
DeleteType

AddTypeAttribute
SetTypeAttribute
GetTypeAttribute
ListTypeAttributes (returns slice)
RemoveTypeAttribute
```

### Variable Naming

Variable names should be short and concise.

We follow the [official go guidelines](https://github.com/golang/go/wiki/CodeReviewComments#variable-names) and try to be consistent with Go standard library as much as possible, everything not defined in the tables below should follow Go standard library naming scheme. In general, variable names are English and descriptive, omitting abbreviations as much as possible (except for the tables below), as well as putting adjectives and adverbs before the noun and verb respectively.

#### Single-word Entities

| entity               | name    | example type                                                  |
| :------------------: | :-----: | :-----------------------------------------------------------: |
| context              | ctx     | context.Context                                               |
| mutex                | mu      | sync.Mutex                                                    |
| configuration        | conf    | go.thethings.network/lorawan-stack/v3/pkg/config.Config       |
| logger               | logger  | go.thethings.network/lorawan-stack/v3/pkg/log.Logger          |
| message              | msg     | go.thethings.network/lorawan-stack/v3/api/gateway.UplinkMessage  |
| status               | st      | go.thethings.network/lorawan-stack/v3/api/gateway.Status         |
| server               | srv     | go.thethings.network/lorawan-stack/v3/pkg/networkserver.Server|
| ID                   | id      | string                                                        |
| unique ID            | uid     | string                                                        |
| counter              | cnt     | int                                                           |
| gateway              | gtw     |                                                               |
| application          | app     |                                                               |
| end device           | dev     |                                                               |
| user                 | usr / user |                                                            |
| transmit             | tx / Tx |                                                               |
| receive              | rx / Rx |                                                               |

The EUI naming scheme can be found in the well-known variable names section bellow.

#### 2-word Entities

In case both of the words have an implementation-specific meaning, the variable name is the combination of first letter of each word.

| entity                                                  | name    |
| :-----------------------------------------------------: | :-----: |
| wait group                                              | wg      |
| Application Server                                      | as      |
| Gateway Server                                          | gs      |
| Identity Server                                         | is      |
| Join Server                                             | js      |
| Network Server                                          | ns      |

In case one of the words specifies the meaning of the variable in a specific language construct context, the variable name is the combination of abbrevations of the words.

#### Well-known Variable Names

These are the names of variables that occur often in the code. Be consistent in naming them, even when their
meaning is obvious from the context.

| entity                          | name    |
| :-----------------------------: | :-----: |
| gateway ID                      | gtwID   |
| gateway EUI                     | gtwEUI  |
| application ID                  | appID   |
| application EUI                 | appEUI  |
| join EUI                        | joinEUI |
| device ID                       | devID   |
| device EUI                      | devEUI  |
| user ID                         | usrID / userID |

### Event Naming

Events are defined with

```go
events.Define("event_name", "event description")
```

The event name is usually of the form `component.entity.action`. Examples are `ns.up.receive_duplicate` and `is.user.update`. We have some exceptions, such as `ns.up.join.forward`, which is specifically used for join messages.

The event description describes the event in simple English. The description is capitalized by the frontend, so the message should be lowercase, and typically doesn't end with a period. Event descriptions will be translated to different languages.

### Error Naming

Errors are defined with

```go
errors.Define<Type>("error_name", "description with `{attribute}`", "other", "public", "attributes")
```

Error definitions must be defined as close to the return statements as possible; in the same package, and preferably above the concerning function(s). Avoid exporting error definitions unless they are meaningful to other packages, i.e. for testing the exact error definition. Keep in mind that error definitions become part of the API.

Prefer using a specific error type, i.e. `errors.DefineInvalidArgument()`. If you are using a cause (using `WithCause()`), you may use `Define()` to fallback to the cause's type.

The error name in snake case is a short and unique identifier of the error within the package. There is no need to append `_failed` or `_error` or prepend `failed_to_` as an error already indicates something went wrong. Be consistent in wording (i.e. prefer the more descriptive `missing_field` over `no_field`), order (i.e. prefer the more clear `missing_field` over `field_missing`) and avoid entity abbreviations.

The error description in lower case, with only names in title case, is a concise plain English text that is human readable and understandable. Do not end the description with a period. You may use attributes, in snake case, in the description defined between backticks (`` ` ``) by putting the key in curly braces (`{ }`). See below for naming conventions. Only provide primitive types as attribute values using `WithAttributes()`. Error descriptions will be translated to different languages.

### Log Field Keys, Event Names, Error Names, Error Attributes and Task Identifiers

Any `name` defined in the following statements:

- Logging field key: `logger.WithField("name", "value")`
- Event name: `events.Define("name", "description")`
- Error name: `errors.Define("name", "description")`
- Error attribute: ``errors.Define("example", "description `{name}`")``
- Task identifier: `c.RegisterTask("name", ...)`

Shall be snake case, optionally having an event name prepended with a dotted namespace, see above. The spacer `_` shall be used in LoRaWAN terms: `DevAddr` is `dev_addr`, `AppSKey` is `app_s_key`, etc.

### Comments

All comments should be English sentences, starting with a capital letter and ending with a period.

Every Go package should have a package comment. Every top-level Go type, const, var and func should have a comment. These comments are recognized by Go's tooling, and added to the [Godoc for our code](https://godoc.org/go.thethings.network/lorawan-stack). See [Effective Go](https://golang.org/doc/effective_go.html#commentary) for more details.

Although Go code should typically explain itself, it's sometimes important to add additional comments inside funcs to communicate what a block of code does, how a block of code does that, or why it's implemented that way.

Comments can also be used to indicate steps to take in the future (*TODOs*). Such comments look as follows:

```go
// TODO: Open the pod bay doors (https://github.com/TheThingsNetwork/lorawan-stack/issues/<number>).
```

In our API definitions (`.proto` files) we'd like to see short comments on every service, method, message and field. Code that is generated from these files does not have to comply with guidelines (such as Go's guideline for starting the comment with the name of the thing that is commented on).

## JavaScript Code Style

For our frontend development, we use a syntax based on ES6 with a couple of extensions from later standards. The code is transpiled via [`webpack`](https://webpack.js.org/) using [`babel`](https://babeljs.io/) to be interpreted by the browser or Node.JS.

### Code Formatting

We use [`prettier`](https://prettier.io/) and [`eslint`](https://eslint.org/) to conform our code to our guidelines as far as possible. Committed code that violates these rules will cause a CI failure. For your convenience, it is hence recommended to set up your development environment to apply autoformatting on every save. We usually don't enforce any formatting or styles that go beyond of what we can ensure using our linting setup. You can check the respective configuration in `/config/eslintrc.yaml`, `/config/.prettierrc.yaml` as well as the global `.editorconfig`.

To run the linter, you can use `mage js:lint` and to format all JavaScript files, you can run `mage js:fmt`.

### Code Comments

Additionally to the overall code comment rules outlined above, we use [JSDoc](https://jsdoc.app)-conform documentation of classes and functions. We also use full English sentences and ending sentence periods here.

Please make sure that these multi-line comments follow the correct format, especially leaving the first line of this multiline JSDoc comments empty:

```js
// Bad.

/** Converts a byte string from hex to base64.
 * @param {string} bytes - The bytes, represented as hex string.
 * @returns {string} The bytes, represented as base64 string.
 */

// Good.

/**
 * Converts a byte string from hex to base64.
 * @param {string} bytes - The bytes, represented as hex string.
 * @returns {string} The bytes, represented as base64 string.
 */
```

It also makes sense to wrap code bits, variable names and URLs in \`\`  quotes, so they can easily be recognized and do not clash with our capitalization rules enforced by eslint, when they are at the beginning of a sentence:

```js
// Bad. This will get flagged by the linter.

// devAddr is a hex string.
const devAddr = '270000FF'

// Good.

// `devAddr` is a hex string.
const devAddr = '270000FF'
```

### Import Statement Order

Our `import` statements use the following order, each separated by empty newlines:

1. Node "builtin" modules (e.g. `path`)
2. External modules (e.g. `react`)
3. Internal modules (e.g. `@ttn-lw/*`, `@console`, etc.)
  1. Constants
  2. API module
  3. Components
    1. Global presentational components (`@ttn-lw/components/*`)
    2. Global container components (`@ttn-lw/containers/*`)
    3. Global utility components (`@ttn-lw/lib/components/*`)
    4. Local presentational components (`@{console|oauth}/components/*`)
    5. Local container components (`@{console|oauth}/containers/*`)
    6. Local utility components (`@{console|oauth}/lib/components/*`)
    7. View components (`@{console|oauth}/views/*`)
  4. Utilities
    1. Global utilities (`@ttn-lw/lib/*`)
    2. Local utilities (`@{console|oauth}/lib/*`)
  5. Store modules
    1. Actions
    2. Reducers
    3. Selectors
    4. Middleware and logics
  6. Assets and styles
4. Parent modules (e.g. `../../../module`)
5. Sibling modules (e.g. `./validation-schema`, `./button.styl`)
6. Index of the current directory (`.`)

Note that this order is enforced by our linter and will cause a CI fail when not respected. Again, settting up your development environment to integrate linting will assist you greatly here.

### React Component Syntax (Functional, Class Components and Hooks)

Lately, we have been embracing [react hooks](https://reactjs.org/docs/hooks-overview.html) and write all new components using this approach. However, there are a lot of class components from the time before react hooks which we will try to refactor successively.

#### A note on decorators and HOCs

Decorators provided an easy syntax to wrap Classes around functions and we have used this syntax extensively during early stages of development. We now consider decorators and HOCs as hindrance with regards to our aim to adopt hooks. As a result, we refrain from introducing new higher order components and implement hooks instead. This will help us avoiding decorators as well as literal (concatenated) wrappers for function components.

### React Component Types

We differentiate four different component scopes:
- Presentational Components (global and application level)
- Container Components (global and application level)
- View Components
- Utility Components (global and application level)

The differentiation is not always 100% clear and we tend not to be too dogmatic about it. Additionally, the introduction of react hooks tends too break up these traditional categorizations even more and might necessitate a review of these in the near future.

Generally we understand these component types as follows:

#### Presentational Components

These are UI elements that primarily serve a presentational purpose. They implement the basic visual interface elements of the application, focusing on interaction and plain UI logic. They never connect to the store or perform any data fetching or have any other side effects and render rich DOM trees which are also styled according to our design guidelines.

Examples for presentational components are simple UI elements such as buttons, input elements, navigations, breadcrumbs. They can also combine and extend functionality of other presentational components by composition to achieve more complex elements, such as forms. We also regard our application specific forms as such components, as long as they don't connect to the store or perform the data fetching themselves.

To decide whether a component is a presentational component, ask yourself:
- Is this component more concerned with how things look, rather than how things work?
- Does this component use no state or only UI state?
- Does this component not fetch or send data?
- Does this component render a lot of (nested and styled) DOM nodes?

If you answered more than 2 questions with yes, then you likely have a presentational component.

Presentational components should **always** define storybook stories, to provide usage information for other developers.

#### Container Components

Container components focus more on state logic, data fetching, store management and similar concerns. They usually perform business logic and eventually render results using presentational components.

An example for a container components are our table components, that manage the fetching and preparation of the respective entity and render the result using our `<Table />` component.

To decide whether a component is a container component, ask yourself:
- Is this component more concerned with how things work, rather than how things look?
- Does this component connect to the store?
- Does this component fetch or send data?
- Is the component generated by higher order components?
- Does this component render simple nodes, like a single presentational component?

If you can answer more than 2 questions with yes, then you likely have a container component.

#### View components

View components always represent a single view of the application, represented by a single route. They structurize the overall appearance of the page, obtain global state information, fetch necessary data and pass it down (implicitly via the store or explicitly as props) mostly to container components, but also to presentational components. Usually, these components also define submit and error handlers of the forms that they render. Otherwise, these components should not employ excessive (stateful) logic which should rather be handled by container components. It should focus on globally structurizing the page using the grid system and respective containers.

##### View component checklist

- Conciseness and no stateful logic (use containers instead)
- Uses `<PageTitle />` component to define heading and page title
- Uses breadcrumbs (if within breadcrumb view)
- Fetching necessary data (via `withRequest` HOC), if not done by a container
- Unavailable "catch-all"-routes are caught by `<NotFoundRoute />` component, including subviews
- Errors should be caught by the `<ErrorView />` error boundary component
- Ensured responsiveness and usage of the grid system

#### Utility components

These components do not render any DOM elements and are hence not *visible* by themselves. Utility components can be higher order components or similar components, that modify their children or introduce a side effect to the render tree.

To decide whether a component is a utility component ask yourself:
- Is this component a higher order component?
- Is this component invisible on its own?
- Is this component an abstraction layer on top of another component?

If you can answer at least one of those questions with yes, then you likely have a container component.

#### Global or Application Scope?

Components can be categorized as either local (e.g. `pkg/webui/{console|oauth}/{components|containers}`) or global (e.g. `pkg/webui/{components|containers}`). The distinction should come naturally: Global components are ones that can be used universally in every application. Local components are tied to a specific use case inside the respective application.

Sometimes, you might find that during implementing an application specific component that it can actually be generalized without much refactoring and hence be a useful addition to our global component library.

### Frontend Related Pull Requests

Pull requests for frontend related changes generally follow our overall pull request scheme. However, in order to assist reviewers, a browser screenshot of the changes is included in the PR comment, if applicable.

It might help you to employ the following checklist before opening the pull request:

* [ ] All visible text is using [i18n messages](#frontend-translations)?
* [ ] Assets minified or compressed if possible (e.g. SVG assets)?
* [ ] Screenshot in PR description?
* [ ] Responsiveness checked?
* [ ] New components [categorized correctly](#global-or-application-scope) (global/local, container/component)?
* [ ] Feature flags added (if applicable)?
* [ ] All views use  `<PageTitle />` or `<IntlHelmet />` properly?
* [ ] Storybook story added / updated?
* [ ] Prop types / default props added?

## Translations

We do our best to make all text that could be visible to users available for translation. This means that all text of the console's user interface, as well as all text that it may forward from the backend, needs to be defined in such a way that it can be translated into other languages than English.

### Backend Translations

In the API, the enum descriptions, error messages and event descriptions available for translation. Enum descriptions are defined in `pkg/ttnpb/i18n.go`. Error messages and event descriptions are defined with `errors.Define(...)` and `events.Define(...)` respectively.

These messages are then collected in the `config/messages.json` file, which will be processed in the frontend build process, but may also be used by other (native) user interfaces. When you define new enums, errors or events or when you change them, the messages need to be updated into the `config/messages.json` file.

```bash
$ tools/bin/mage go:messages
```

If you forget to do so, this will cause a CI failure.

Adding translations of messages to other languages than English is a matter of adding key/value pairs to `translations` in `config/messages.json`.

### Frontend Translations

The frontend uses [`react-intl`](https://github.com/yahoo/react-intl), which helps us greatly to define text messages used in the frontend.

The workflow for defining messages is as follows:

1. Add a `react-intl` message using `intl.defineMessages({…})`
  * This can be done either inline, or by adding it to `pkg/webui/lib/shared-messages.js`
2. Use the message in components (e.g. `sharedMessage.myMessage`)

After adding messages this way, it needs to be added the locales file `pkg/webui/locales/*.js` by using:

```bash
$ tools/bin/mage js:translations
```

> Note: When using `tools/bin/mage js:serve`, this command will be run automatically after any change.

The message definitions in `pkg/webui/locales` can be used to provide translations in other languages (e.g. `fr.js`). Keep in mind that locale files are checked in and committed, any discrepancy in the locales file with the defined messages will lead to a CI failure.

## Events

In addition to the previously described translation file that we generate, we also generate a data file that contains all event definitions. This file is then loaded by the documentation system so that we can generate documentation for our events.

After adding or changing events, regenerate this file with:

```bash
$ tools/bin/mage go:eventData
```

## Testing

### Unit Tests

To run unit tests, use the following mage targets:

```bash
$ tools/bin/mage go:test js:test jsSDK:test
```

### End-to-end Tests

We use [Cypress](https://cypress.io) for running frontend-based end-to-end tests. The tests specifications are located at `/cypress/integration`.

#### Running frontend end-to-end tests locally

Make sure to [build the frontend assets](#building-the-frontend) and run `tools/bin/mage dev:initStack dev:sqlDump` to create a seed database which the tests can reset the database to in between runs. To run the stack when working on end-to-end tests, use the `tools/bin/mage dev:startDevStack` command. This will run the runs The Things Stack with proper configuration for the end-to-end tests. Note: this command does not output anything, but logs are written to `.cache/devStack.log`.

[Cypress](https://www.cypress.io/) provides two modes for running tests: headless and interactive.
- **Headless mode** will not display any browser GUI and output test progress into your terminal instead. This is helpful when one just needs see the results of the tests.
- **Interactive mode** will run an Electron based application together with the full-fledged browser. This is helpful when developing frontend applications as it provides hot reload, time travelling, browser extensions and DOM access.

> Note: Currently, we test our frontend only in Chromium based browsers.

You can run Cypress in the headless mode by running the following command:

```bash
$ tools/bin/mage js:cypressHeadless
```

You can run Cypress in the interactive mode by running the following command:

```bash
$ tools/bin/mage js:cypressInteractive
```

#### JavaScript based tests

We find the [JS Unit Testing Guide](https://github.com/mawrkus/js-unit-testing-guide) a good starting point for informing our testing guidelines and we recommend reading through this guide. Note, that we employ some different approaches regarding [Grammar and Capitalization](#Grammar-and-capitalization).

We have extracted and adapted the most important parts below.

##### Pattern

The goal of naming our tests is to have a concise and streamlined description helping us to understand what a test is testing specifically. In order to do that, we follow a **"unit of work - scenario/context - expected behaviour"** pattern:
```js
// Schema.
describe('[unit of work]', () => {
  it('[expected behaviour] when [scenario/context]', () => {
    …
  });
});

// Example.
describe('Login', () => {
  it('succeeds when using correct credentials', () => {
    …
  });
});
```

This pattern will also help you organizing your tests better.

##### Grammar and Capitalization

Avoid using the modal verb `should` when describing tests. This will add redundancy and unnecessary verbosity to the test description. Instead, use a simple present tense sentence without any modality and with `when` as conjunction. Don't use end of sentence periods.

The `[unit of work]` bit, as part of the outermost `describe()` function is always capitalized, whereas the `[expected behavior]` part of the `it()` function is always lowercase. This way, the suit will generate proper english sentences when concatenating the test descriptions.

```js
// Bad: using `should`.
describe('Login', () => {
  it('should succeed when using correct credentials', () => {
    …
  });
});

// Bad: wrong capitalization.
describe('login', () => {
  it('Succeeds when using correct credentials', () => {
    …
  });
});

// Good: No should and proper capitalization.
describe('Login', () => {
  it('succeeds when using correct credentials', () => {
    …
  });
});

```

##### React Components

When testing react components, the name of the component is written as `<ReactComponent />`.

```js
// Bad: not using JSX syntax.
describe('MyComponent', () => {
  it('matches snapshot', () => {
    …
  });
});

// Bad: describing the component instead of naming it.
describe('My component', () => {
  it('matches snapshot', () => {
    …
  });
});

// Good
describe('<MyComponent />', () => {
  it('matches snapshot', () => {
    …
  });
});

```

##### Structurizing tests

We always use the `describe() / it()` hooks to write all tests, even if there's only one test in the suite. This keeps our tests streamlined and allows for easy extension of the test suite.

```js
// Bad: using `test()` hook.
test('flattens the object', () => {
  …
});

// Good: using `describe() / it()` hooks
describe('Get by path', () => {
  it('succeeds when using correct credentials', () => {
    …
  });
});

```

It's fine to use multiple hierarchies of `describe()` to group related tests more accurately:

```js
describe('User registration', () => {
  it('succeeds when using valid inputs', () => {
  });

  describe('when using invalid input values', () => {
    it('shows an error notification', () => {
    });

    it('does not perform a redirect', () => {
    });
  });

  describe('when using an already registered email', () => {
    it('shows an error notification', () => {
    });
  });
});
```

##### Test Driven Development (TDD)

Test Driven Development is a development philosophy that puts tests at the core of development. At The Things Industries, we don't enforce this method but we strongly encourage to adopt a process that emphasizes testing. Since adding fontend-based end-to-end tests to our codebase, we plan to do the following:

1. Writing end-to-end tests for all newly added features
2. Writing end-to-end tests for each (significant) bug that was resolved
3. Gradually adding coverage to existing features

Currently, we only employ frontend-based end-to-end tests, meaning that these tests can only be written if they are also operable through the frontend.

#### Writing End-to-End Tests

It is highly suggested to read [Cypress documentation](https://docs.cypress.io/guides) before starting to write tests.

##### Guiding Principle

We follow the following principle for writing useful end-to-end tests:

> The more your tests resemble the way your software is used, the more confidence they can give you.

This means that when writing tests, we always consider the real-life equivalent of the test scenario to design the test setup. This means:

##### Selecting elements

In line with the principle mentioned above, we have also included [`Testing Library`](https://testing-library.com/) to use advanced testing utilities. `Testing Library` has a good guide for [how to select elements](https://testing-library.com/docs/guide-which-query). We try to follow this guide for our end-to-end tests.

In some cases it can be necessary to select DOM elements using a special selection data attribute. We use `data-test-id` for this purpose. Use this attribute to select DOM elements when more realistic means of selection are not sufficient. Use meaningful but concise ID values, such as `error-notification`.

- Select DOM elements using text captions and labels when possible.
  - Select form fields by its label via `cy.findByLabelText`, e.g. `cy.findByLabelText('User ID')`. Same for field errors,warnings and descriptions, use `cy.findErrorByLabelText`, `cy.findWarningByLabelText` and `cy.findDescriptionByLabelText`.
  - Select buttons, links, tabs and other elements that are described by [ARIA roles](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/ARIA_Techniques#Roles) via `cy.findByRole`, e.g. `cy.findByRole('button', {name: 'Submit'})`.
  - Select text elements via `cy.findByText`.
- Assert that selected elements are visible.

  ```html
  <!-- Instead of `visibility: hidden` it could be `display:none` or `z-index: -1` as well. -->
  <div data-test-id="test" style="visibility: hidden">
    Test content
  </div>
  ```

  ```js
  // Bad. This assertions will pass while not being visible to the user.
  cy.findByTestId('test').should('exists')

  // Good. This assertion will rightfully fail.
  cy.findByTestId('test').should('be.visible')
  ```

##### Test runner globals

Cypress uses [Mocha](https://mochajs.org/) as the test runner internally, while for unit tests we use [Jest](https://jestjs.io/). To keep our tests consistent we prefer using globals from `Jest` when possible.


Jest globals | Mocha globals | Used for
--- | --- | ---
`describe` | `describe` | Group together related tests
`it` | `it` | Define a single test
`beforeEach`/`afterEach` | `beforeEach`/`AfterEach` | Hook before/after each test (`it`)
`beforeAll`/`afterAll` | `before`/`after` | Hook before/after test block (`describe`)

##### End-to-end tests file structure

```bash
./pkg/cypress
|-- fixtures                    Cypress mocks
|-- integration                 frontend end-to-end specifications (1)
|   |-- console                 Console related end-to-end tests
|   |   |-- users               tests related to user entity
|   |   |-- ...
|   |   |-- shared              tests that are not directly related to a specific entity (2)
|   |-- oauth                   OAuth related end-to-end tests
|   |   `-- ...
|   `--smoke                    smoke tests (3)
|-- plugins                     Cypress plugins
|-- support                     Cypress commands and test utilities
|-- screenshots                 screenshots generated when running tests (4)
`-- videos                      videos generated when running tests (5)
```

1. `pkg/cypress/integration` contains all test specifications.
  - Each test file must be placed into corresponding folder (`console`/`oauth`/`smoke`).
  - Each test must follow the following naming: `{context}.spec.js`.
  - One test file must have end-to-end tests dedicated only to a specific entity or view.
2. `pkg/cypress/{console|oauth}/shared` contains all test specification not directly related to a single entity or a view. For example, `side-navigation.spec.js` or `header.spec.js` must be placed into the `cypress/console/shared` folder because both components are present on multiple views and are partially related to the stack entities. Make sure to scope cypress selections within the tested component using `cy.within`.
3. `pkg/cypress/integration/smoke` contains tests that simulate a complete user story trying to do almost everything a typical user would do. For example, a typical smoke test can verify that the user is able to register, login, create application and register The Things Uno. For more details and diffeence between regular end-to-end and smoke tests see the [End-to-end tests structure](#organizing-end-to-end-tests) section.
4. and 5. Cypress stores screenshots and videos to the appropriate folder after running end-to-end tests. These should not be added to the repository.

##### Organizing end-to-end tests

When writing end-to-end tests we comply with the following guidelines:

- Tests are grouped by views for a specific entity. For example, when testing creation of application API keys:
  1. Add test file `cypress/integration/console/applications/api-keys/create.spec.js`.
  2. Test the behavior of the API key create view independently from any other specification.
- Do not repeat actions via the UI that are not related to the current test context. Consider adding reusable [Cypress commands](https://docs.cypress.io/api/cypress-api/custom-commands.html) that do necessary test setup programmatically. This means that when testing any UI that is not the login specification and requires the user to be logged in, there is no need to log in through the login page, while we simply fetch the access token. Note: this does not mean that one cannot create a cypress command that performs actions via UI.
- Extract components that appear on various views and test them separately instead of making assertions in each test where this component is used. For example, such components could be the page header and side navigation.
- Dedicate at least one test to assert that the view displays its UI elements in place on initial load. Assert on UI changes in tests that trigger these changes.
- Prefer duplicating entities with non-conflicting ids in tests instead of executing database teardown before each test. For example, when testing various scenarios for registering gateway, consider creating gateways with different ID's and EUI's instead of using a single gateway and drop the database before each test. Note: try to use this approach when possible, otherwise do not hesitate to restore database state before each test.
- Consider various stack configurations when writing end-to-end tests. Some views have different UI depending on availability of different stack components. For example, the end device wizard looks different for deployments with complete cluster (NS+JS+AS) and for JS-only configuration. Likewise, sections and entire views can be enabled or disabled based on our feature toggles. If your test scenario differs based on different feature toggle conditions, make sure to probe these preconditions in your tests.

##### Smoke tests

We distinguish between regular end-to-end tests and smoke tests. While regular end-to-end tests are scoped to a specific view or component and tests those in depth, smoke tests are testing complete user stories that are critical to the overall integrity of the application and usually comprise multiple components and views, e.g. login flow, user registration or creation of applications. When writing smoke tests we comply with the following guidelines:

- Smoke tests are testing complete user stories in a **wide and shallow** manner, meaning:
  - performing some complex and critical flow that touches multiple components, APIs and/or views
  - not testing different configurations or preconditions of the same flow in depth
  - For example, when testing registration of The Things Uno:
    1. Add test file `cypress/integration/smoke/devices/create.js`
    2. Describe the whole user story to register the device including creating an application (or using an existing one), link the application and create the end device.
- One smoke test should be encapsulated into a single `describeSmokeTest` declaration.

## Building and Running

There is a single binary for the server, `ttn-lw-stack`, as well as a binary for the command-line interface `ttn-lw-cli`. The single binary contains all components start one or multiple components. This allows you to run The Things Stack with one command in simple deployment scenarios, as well as distributing micro-services for more advanced scenarios.

We provide binary releases for all supported platforms, including packages for various package managers at https://github.com/TheThingsNetwork/lorawan-stack/releases. We suggest you use the compiled packages we provide in production scenarios.

Before the binaries are built, the frontend needs to be built. You can control whether to build the frontend for production or development by setting the `NODE_ENV` environment variable to either `development` or `production`.

The difference of a development build includes:

- Including source maps
- Using DLL bundle for modules to reduce build time
- A couple of special build options to improve usage with `webpack-dev-server`

The frontend can then be built using:

```bash
$ tools/bin/mage js:build
```

For development/testing purposes we suggest to run the binaries directly via `go run`:

```bash
$ go run ./cmd/ttn-lw-stack start
```

It is also possible to use `go build`, or release snapshots, as described below.

## Releasing

The Things Stack uses [GoReleaser](https://goreleaser.com/) for releases. If you want to build a release (snapshot), you first need to install GoReleaser:

```bash
$ go install github.com/goreleaser/goreleaser@v1.2.5
```

The command for building a release snapshot is:

```bash
$ goreleaser --snapshot -f .goreleaser.snapshot.yml --rm-dist
```

The command for building a full release is:

```bash
$ goreleaser -f .goreleaser.release.yml --rm-dist
```

> Note: Goreleaser is configured to sign binaries, as per GitHub Action in `.github/workflows/release-*.yml`. If you're doing a release locally, you will need key's passphrase, or need to skip the signing step.

> Note: You will at least need to have [`rpm`](http://rpm5.org/) and [`snapcraft`](https://snapcraft.io/) in your `PATH` if you want to build a full release.

This will compile binaries for all supported platforms, `deb`, `rpm` and Snapcraft packages, release archives in `dist`, as well as Docker images.

> Note: The operating system and architecture represent the name of the directory in `dist` in which the binaries are placed.
> For example, the binaries for Darwin x64 (macOS) will be located at `dist/darwin_amd64`.

A new version is released from the `v3.n` branch. The necessary steps for each are detailed below.

> Note: To get the target version, you can run `version=$(tools/bin/mage version:bumpXXX version:current)`, where xxx is the type of new release (minor/patch/RC). Check the section [Version Bump](#version-bump) for more information.

### Release From Master

Create a `Release` issue in this repository and follow the steps.

### Release Backports

Create a `Backport Release` issue in this repository and follow the steps.

## Troubleshooting

### Console

#### Problem: Assets are not found

The Console will render a blank page and you can see backend logs like e.g.:
```
INFO Request handled                          duration=40.596µs error=error:pkg/errors/web:unknown (Not Found) message=Not Found method=GET namespace=web remote_addr=[::1]:50450 request_id=01DZ2CJDWKAFS10QD1NKZ1D56H status=404 url=/assets/console.36fcac90fa2408a19e4b.js
```

You might also see error messages in the Console such as:
```
Uncaught ReferenceError: libs_472226f4872c9448fc26 is not defined
    at eval (eval at dll-reference libs_472226f4872c9448fc26 (console.js:26130), <anonymous>:1:18)
    at Object.dll-reference libs_472226f4872c9448fc26 (console.js:26130)
    at …
```

#### Possible causes

##### Using incorrect or no configs

If you plan to run the Console / Account App in development mode, it is important to ensure that the right configuration is loaded both for running The Things Stack itself, as well as the development tooling (e.g. `tools/bin/mage serve`).

##### Possible solution

Make sure that you source the environment variables as [described above](#development-configuration) **in all terminal sessions** that run The Things Stack or the `tools/bin/mage` commands.

##### Missing restart

The stack has not been restarted after the Console bundle has changed. In production mode, The Things Stack will access the bundle via a filename that contains a content-hash, which is set during the build process of the Console. The hash cannot be updated during runtime and will take effect only after a restart.

##### Possible solution

  1. Restart the The Things Stack

##### Accidentally deleted bundle files

The bundle files have been deleted. This might happen e.g. when a mage target encountered an error and quit before running through.

##### Possible solution

  1. Rebuild the Console `tools/bin/mage js:clean js:build`
  2. Restart The Things Stack

##### Mixing up production and development builds

If you switch between production and development builds of the Console, you might forget to re-run the build process and to restart The Things Stack. Likewise, you might have arbitrary config options set that are specific to a respective build type.

##### Possible solution

  1. Double check whether you have set the correct environment: `echo $NODE_ENV`, it should be either `production` or `development`
  2. Double check whether [your The Things Stack config](#development-configuration) is set correctly (especially `TTN_LW_CONSOLE_UI_JS_FILE`, `TTN_LW_CONSOLE_UI_CANONICAL_URL` and similar settings). Run `ttn-lw-stack config --env` to see all environment variables
  3. Make sure to rebuild the Console `tools/bin/mage js:clean js:build`
  4. Restart The Things Stack

#### Problem: Console rendering blank page and showing arbitrary error message in console logs, e.g.:

```
console.4e67a17c1ce5a74f3f50.js:104 Uncaught TypeError: m.subscribe is not a function
    at Object../pkg/webui/console/api/index.js (console.4e67a17c1ce5a74f3f50.js:104)
    at o (console.4e67a17c1ce5a74f3f50.js:1)
    at Object../pkg/webui/console/store/middleware/logics/index.js (console.4e67a17c1ce5a74f3f50.js:104)
    at o (console.4e67a17c1ce5a74f3f50.js:1)
    at Object.<anonymous> (console.4e67a17c1ce5a74f3f50.js:104)
    at Object../pkg/webui/console/store/index.js (console.4e67a17c1ce5a74f3f50.js:104)
    at o (console.4e67a17c1ce5a74f3f50.js:1)
    at Module../pkg/webui/console.js (console.4e67a17c1ce5a74f3f50.js:104)
    at o (console.4e67a17c1ce5a74f3f50.js:1)
    at Object.0 (console.4e67a17c1ce5a74f3f50.js:104)
```

#### Possible causes

##### Bundle using old JS SDK

The bundle integrates an old version of the JS SDK. This is likely a caching/linking issue of the JS SDK dependency.

##### Possible solutions

- Re-establish a proper module link between the Console and the JS SDK
  - Run `tools/bin/mage js:cleanDeps js:deps`
  - Check whether the `ttn-lw` symlink exists inside `node_modules` and whether it points to the right destination: `lorawan-stack/sdk/js/dist`
    - If you have cloned multiple `lorawan-stack` forks in different locations, `yarn link` might associate the JS SDK module with the SDK on another ttn repository
  - Rebuild the Console and (only after the build has finished) restart The Things Stack

#### Problem: Console rendering blank page and showing `Module not found` message in console logs, e.g.:

```
ERROR in ./node_modules/redux-logic/node_modules/rxjs/operators/index.js Module not found: Error: Can't resolve '../internal/operators/audit' in '/lorawan-stack/node_modules/redux-logic/node_modules/rxjs/operators'
```

##### Possible cause: Broken yarn or npm cache

##### Possible solution: Clean package manager caches

- Clean yarn cache: `yarn cache clean`
- Clean npm cache: `npm cache clean`
- Clean and reinstall dependencies: `tools/bin/mage js:cleanDeps js:deps`

#### Problem: The build crashes without showing any helpful error message

##### Cause: Not running mage in verbose mode

`tools/bin/mage` runs in silent mode by default. In verbose mode, you might get more helpful error messages

##### Solution

Run mage in verbose mode: `tools/bin/mage -v {target}`

#### Problem: Browser displays error:
`Cannot GET /`

##### Cause: No endpoint is exposed at root

##### Solution:

Console is typically exposed at `http://localhost:8080/console`,
API at `http://localhost:8080/console`,
OAuth at `http://localhost:8080/oauth`,
etc

#### Problem: Browser displays error:
`Error occurred while trying to proxy to: localhost:8080/console`

##### Cause: Stack is not available or not running

##### Solution:

For development, remember to run the stack with `go run`:

```bash
$ go run ./cmd/ttn-lw-stack start
```

#### General advice

A lot of problems during build stem from fragmented, incomplete runs of mage targets (due to arbitrary errors happening during a run). Oftentimes, it then helps to build the entire Web UI from scratch: `tools/bin/mage jsSDK:cleanDeps jsSDK:clean js:cleanDeps js:clean js:build`, and (re-)start The Things Stack after running this.
