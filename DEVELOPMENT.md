# The Things Stack for LoRaWAN Development

The Things Stack components are primarily built in Go, we use React for web front-ends. It is assumed that you have decent knowledge and experience with these technologies. If you want to get more familiar with Go, we strongly recommend to take [A Tour of Go](https://tour.golang.org/).

## Development Environment

The Things Network's development tooling uses [Mage](https://magefile.org/). Under the hood, `mage` calls other tools such as `git`, `go`, `yarn`, `docker` etc. Recent versions are supported; Node v12.x and Go v1.13.x.

- Follow [Go's installation guide](https://golang.org/doc/install) to install Go.
- Download Node.js [from their website](https://nodejs.org) and install it.
- Follow [Yarn's installation guide](https://yarnpkg.com/en/docs/install) to install Yarn.
- Follow the guides to [install Docker](https://docs.docker.com/install/#supported-platforms) and to [install Docker Compose](https://docs.docker.com/compose/install/#install-compose).

## Cloning the repository

If you are unfamiliar with forking projects on GitHub or cloning them locally, please [see the GitHub documentation](https://help.github.com/articles/fork-a-repo/).

## Getting started

As most of the tasks will be managed by `make` and `mage` we will first initialize the tooling:

```bash
$ make init
```

You may want to run this commands from time to time.

Now you can initialize the development databases with some defaults.

>Note: this requires Docker to be running.

```bash
$ make dev.stack.init
```

This starts a CockroachDB and Redis database in Docker containers, creates a database, migrates tables and creates a user `admin` with password `admin`.

## Managing the development databases

You can use the following commands to start, stop and erase databases.

```bash
$ make dev.databases.start # Starts all databases in a Docker container
$ make dev.databases.stop  # Stops all databases

# The contents of the databases will be saved in .dev/data.

$ make dev.databases.erase # Stop all databases and erase storage.
```

### CockroachDB

CockroachDB is a distributed SQL database that we use in the Identity Server.

You can use `make dev.databases.sql` to enter an SQL shell.

### Redis

Redis is an in-memory data store that we use as a database for "hot" data.

You can use `make dev.databases.redis-cli` to enter a Redis-CLI shell.

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
├── .mage               dev/test/build tooling
├── README.md           general information about this project
│   ...
├── api                 contains the protocol buffer definitions for our API
├── cmd                 contains the different binaries that form The Things Stack for LoRaWAN
│   ├── internal        contains internal files shared between the different binaries
│   │   ...
│   ├── ttn-lw-cli      the command-line-interface for The Things Stack for LoRaWAN
│   └── ttn-lw-stack    bundles the server binaries that form The Things Stack for LoRaWAN
├── config              configuration for our JavaScript SDK and frontend
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
$ ./mage proto:clean proto:all jsSDK:definitions
```

### Documentation

The documentation site for The Things Stack is built from the `doc` folder. 
All content is stored as Markdown files in `doc/content`.

In order to build the documentation site with the right theme, you need to run
`./mage docs:deps` from time to time. 

You can start a development server with live reloading by running
`./mage docs:server`. This command will print the address of the server.

The documentation site can be built by running `./mage docs:build`. This will 
output the site to `docs/public`.

For more details on how our documentation site is written, see the [Hugo docs](https://gohugo.io/documentation/).

### Web UI

The Things Stack for LoRaWAN includes two frontend applications: the **Console** and **OAuth Provider**. Both applications use [React](https://reactjs.org/) as frontend framework. The `console` and `oauth` packages of the backend expose their respective web servers and handle all logic that cannot be done in the browser. Otherwise both applications are single page applications (SPA) that run entirely in the browser.

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
├── manifest.go       generated manifest of the frontend, containing file hashes
├── template.go       go template module used to render the frontend HTML
```

For development purposes, the frontend can be run using `webpack-dev-server`. After following the [Getting Started](#getting-started) section to initialize The Things Stack and doing an initial build of the frontend via `mage js:build`, it can be served using:

```bash
$ export NODE_ENV=development
$ mage js:serve
```

The development server runs on `http://localhost:8080` and will proxy all api calls to port `1885`. The serve command watches any changes inside `pkg/webui` and refreshes automatically.

In order to set up The Things Stack to support running the frontend via `webpack-dev-server`, the following environment setup is needed:

```bash
NODE_ENV="development"
TTN_LW_LOG_LEVEL="debug"
TTN_LW_IS_OAUTH_UI_JS_FILE="libs.bundle.js oauth.js"
TTN_LW_CONSOLE_UI_JS_FILE="libs.bundle.js console.js"
TTN_LW_CONSOLE_UI_CANONICAL_URL="http://localhost:8080/console"
TTN_LW_CONSOLE_OAUTH_AUTHORIZE_URL="http://localhost:8080/oauth/authorize"
TTN_LW_CONSOLE_OAUTH_TOKEN_URL="http://localhost:8080/oauth/token"
TTN_LW_IS_OAUTH_UI_CANONICAL_URL="http://localhost:8080/oauth"
TTN_LW_IS_EMAIL_NETWORK_IDENTITY_SERVER_URL="http://localhost:8080/oauth.js"
TTN_LW_CONSOLE_UI_ASSETS_BASE_URL="http://localhost:8080/assets"
```

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

### Line length

We don't have strict rules for line length, but in our experience the following guidelines result in nice and readable code:

- If a line is longer than 80 columns, try to find a "natural" break
- If a line is longer than 120 columns, insert a line break
- In very special cases, longer lines are tolerated

### Formatting and Linting

Go code can be automatically formatted using tools such as [`gofmt`](https://godoc.org/github.com/golang/go/src/cmd/gofmt) and [`goimports`](https://godoc.org/golang.org/x/tools/cmd/goimports). The [Go language server](https://github.com/golang/tools/tree/master/gopls) can also help with formatting code. There are many editor plugins that automatically format your code when you save your files. We highly recommend using those.

We use [`revive`](http://github.com/mgechev/revive) to lint Go code and [`eslint`](https://eslint.org) to lint JavaScript code. These tools should automatically be installed when initializing your development environment.

### Documentation Site

Please respect the following guidelines for content in our documentation site:

- The title of a doc page is already rendered by the build system as a h1, don't add an extra one.
- A documentation page starts with an introduction, and then the first heading. The first paragraph of the introduction is typically a summary of the page. Use a `<!--more-->` to indicate where the summary ends.
- Since the title is a `h1`, everything in the content is at least `h2` (`##`).
- Paragraphs typically consist of at least two sentences.
- Use an empty line between all blocks (headings, paragraphs, lists, ...).
- Prefer text over bullet lists or enumerations. For bullets, use `-`, for enumerations `1.` etc.
- Explicitly call this product "The Things Stack". Not "the stack" etc.
- Avoid shortening, i.e. write "it is" instead of "it's".
- Write guides as a goal-oriented journey.
- Unless already clear from context, use a clearer term than "user", especially if there are multiple kinds (network operator, gateway owner, application developer, ...).
- The user does not have a gender (so use they/them/their).
- Taking screenshots is done as follows:
  - In Chrome: activate the **Developer Tools** and toggle the **Device Toolbar**. In the **Device Toolbar**, select **Laptop with HiDPI screen** (add it if not already there), and click **Capture Screenshot** in the menu on the right.
  - In Firefox: enter **Responsive Design Mode**. In the **Device Toolbar**, select "Laptop with HiDPI screen" (add it if not already there) and **Take a screenshot of the viewport**.
- Use `**Strong**` when referring to buttons in the Console
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
| configuration        | conf    | go.thethings.network/lorawan-stack/pkg/config.Config          |
| logger               | logger  | go.thethings.network/lorawan-stack/pkg/log.Logger             |
| message              | msg     | go.thethings.network/lorawan-stack/api/gateway.UplinkMessage  |
| status               | st      | go.thethings.network/lorawan-stack/api/gateway.Status         |
| server               | srv     | go.thethings.network/lorawan-stack/pkg/network-server.Server  |
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

## Translations

We do our best to make all text that could be visible to users available for translation. This means that all text of the console's user interface, as well as all text that it may forward from the backend, needs to be defined in such a way that it can be translated into other languages than English.

### Backend Translations

In the API, the enum descriptions, error messages and event descriptions available for translation. Enum descriptions are defined in `pkg/ttnpb/i18n.go`. Error messages and event descriptions are defined with `errors.Define(...)` and `events.Define(...)` respectively.

These messages are then collected in the `config/messages.json` file, which will be processed in the frontend build process, but may also be used by other (native) user interfaces. When you define new enums, errors or events or when you change them, the messages need to be updated into the `config/messages.json` file.

```bash
$ ./mage go:messages
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
$ mage js:translations
```

> Note: When using `mage js:serve`, this command will be run automatically after any change.

The message definitions in `pkg/webui/locales` can be used to provide translations in other languages (e.g. `fr.js`). Keep in mind that locale files are checked in and committed, any discrepancy in the locales file with the defined messages will lead to a CI failure.

## Testing

```bash
$ ./mage go:test js:test jsSDK:test
```

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
$ mage js:build
```

For development/testing purposes we suggest to run the binaries directly via `go run`:

```bash
$ go run ./cmd/ttn-lw-stack start
```

It is also possible to use `go build`, or release snapshots, as described below.

## Releasing

You can build a release snapshot with `go run github.com/goreleaser/goreleaser --snapshot`.

> Note: You will at least need to have [`rpm`](http://rpm5.org/) and [`snapcraft`](https://snapcraft.io/) in your `PATH`.

This will compile binaries for all supported platforms, `deb`, `rpm` and Snapcraft packages, release archives in `dist`, as well as Docker images.

> Note: The operating system and architecture represent the name of the directory in `dist` in which the binaries are placed.
> For example, the binaries for Darwin x64 (macOS) will be located at `dist/darwin_amd64`.

Releasing a new version consists of the following steps:

1. Creating a `release/<version>` branch(further, called "release branch") (e.g. `release/3.2.1`).
2. Updating the `CHANGELOG.md` file:
  - Change the **Unreleased** section to the new version and add date obtained via `date +%Y-%m-%d` (e.g. `## [3.2.1] - 2019-10-11`)
  - Check if we didn't forget anything important
  - Remove empty subsections
  - Update the list of links in the bottom of the file
  - Add new **Unreleased** section:
    ```md
    ## [Unreleased]

    ### Added

    ### Changed

    ### Deprecated

    ### Removed

    ### Fixed

    ### Security
    ```
3. Updating the `SECURITY.md` file with the supported versions
4. Bumping the version
5. Writing the version files
6. Creating the version bump commit
7. Creating a pull request from release branch containing all changes made so far to `master`
8. Merging all commits from release branch to `master` locally via `git merge --ff-only release/<version>`
9. Creating the version tag
10. Pushing the version tag
11. Pushing `master`
12. Building the release and pushing to package managers (this is done by CI)

Our development tooling helps with this process. The `mage` command has the following commands for version bumps:

```bash
$ ./mage version:bumpMajor   # bumps a major version (from 3.4.5 -> 4.0.0).
$ ./mage version:bumpMinor   # bumps a minor version (from 3.4.5 -> 3.5.0).
$ ./mage version:bumpPatch   # bumps a patch version (from 3.4.5 -> 3.4.6).
$ ./mage version:bumpRC      # bumps a release candidate version (from 3.4.5-rc1 -> 3.4.5-rc2).
$ ./mage version:bumpRelease # bumps a pre-release to a release version (from 3.4.5-rc1 -> 3.4.5).
```

These bumps can be combined (i.e. `version:bumpMinor version:bumpRC` bumps 3.4.5 -> 3.5.0-rc1). Apart from these bump commands, we have commands for writing version files (`version:files`), creating the bump commit (`version:commitBump`) and the version tag (`version:tag`).

A typical release process is executed directly on the `master` branch and looks like this:

```bash
$ version=$(./mage version:bumpPatch version:current)
$ git checkout -b "release/${version}"
$ ${EDITOR:-vim} CHANGELOG.md SECURITY.md # edit CHANGELOG.md and SECURITY.md
$ git add CHANGELOG.md SECURITY.md
$ ./mage version:bumpPatch version:files version:commitBump
$ git push origin "release/${version}"
```

After this, open a pull request from `release/${version}`. After it is approved:

```bash
$ git checkout master
$ git merge --ff-only "release/${version}"
$ ./mage version:bumpPatch version:tag
$ git push origin ${version}
$ git push origin master
```

After pushing the tag, our CI system will start building the release. When this is done, you'll find a new release on the [releases page](https://github.com/TheThingsNetwork/lorawan-stack/releases). After this is done, you'll need to edit the release notes. We typically copy-paste these from `CHANGELOG.md`.
