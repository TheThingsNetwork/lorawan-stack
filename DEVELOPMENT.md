# The Things Stack for LoRaWAN Development

The Things Stack components are primarily built in Go, we use React for web front-ends. It is assumed that you have decent knowledge and experience with these technologies. If you want to get more familiar with Go, we strongly recommend to take [A Tour of Go](https://tour.golang.org/).

## Development Environment

The Things Network's development tooling uses [Mage](https://magefile.org/). Under the hood, `mage` calls other tools such as `git`, `go`, `yarn`, `docker` etc. Recent versions are supported; Node v12.x and Go v1.14x.

- Follow [Go's installation guide](https://golang.org/doc/install) to install Go.
- Download Node.js [from their website](https://nodejs.org) and install it.
- Follow [Yarn's installation guide](https://yarnpkg.com/en/docs/install) to install Yarn.
- Follow the guides to [install Docker](https://docs.docker.com/install/#supported-platforms) and to [install Docker Compose](https://docs.docker.com/compose/install/#install-compose).

## Cloning the Repository

If you are unfamiliar with forking projects on GitHub or cloning them locally, please [see the GitHub documentation](https://help.github.com/articles/fork-a-repo/).

## Getting Started

As most of the tasks will be managed by `make` and `mage` we will first initialize the tooling:

```bash
$ make init
```

You may want to run this commands from time to time.

Now you can initialize the development databases with some defaults.

```bash
$ ./mage dev:dbStart   # This requires Docker to be running.
$ ./mage dev:initStack
```

This starts a CockroachDB and Redis database in Docker containers, creates a database, migrates tables and creates a user `admin` with password `admin`.

## Using the CLI with the Development Environment

In order to login, you will need to use the correct OAuth Server Address:

```bash
$ export TTN_LW_OAUTH_SERVER_ADDRESS=http://localhost:1885/oauth
$ go run ./cmd/ttn-lw-cli login
```

## Managing the Development Databases

You can use the following commands to start, stop and erase databases.

```bash
$ ./mage dev:dbStart # Starts all databases in a Docker container.
$ ./mage dev:dbStop  # Stops all databases.

# The contents of the databases will be saved in .env/data

$ ./mage dev:dbErase # Stops all databases and erase storage.
```

### CockroachDB

CockroachDB is a distributed SQL database that we use in the Identity Server.

You can use `./mage dev:dbSQL` to enter an SQL shell.

### Redis

Redis is an in-memory data store that we use as a database for "hot" data.

You can use `./mage dev:dbRedisCli` to enter a Redis-CLI shell.

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

Data for generated documentation like API and glossary is stored in `doc/data`.

In order to build the documentation site with the right theme, you need to run
`./mage docs:deps` from time to time.

>Note: as a workaround for [this](https://github.com/gohugoio/hugo/issues/7083), `./mage docs:deps` also pulls the latest version of [frequency-plans.yml](https://github.com/TheThingsNetwork/lorawan-frequency-plans/).

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
├── template.go       go template module used to render the frontend HTML
```

For development purposes, the frontend can be run using `webpack-dev-server`. After following the [Getting Started](#getting-started) section to initialize The Things Stack and doing an initial build of the frontend via `./mage js:build`, it can be served using:

```bash
$ export NODE_ENV=development
$ ./mage js:serve
```

The development server runs on `http://localhost:8080` and will proxy all api calls to port `1885`. The serve command watches any changes inside `pkg/webui` and refreshes automatically.

#### Development Configuration

In order to set up The Things Stack to support running the frontend via `webpack-dev-server`, the following environment setup is needed:

```bash
NODE_ENV="development"
TTN_LW_LOG_LEVEL="debug"
TTN_LW_IS_OAUTH_UI_JS_FILE="libs.bundle.js oauth.js"
TTN_LW_CONSOLE_UI_JS_FILE="libs.bundle.js console.js"
TTN_LW_CONSOLE_UI_CANONICAL_URL="http://localhost:8080/console"
TTN_LW_CONSOLE_OAUTH_AUTHORIZE_URL="http://localhost:8080/oauth/authorize"
TTN_LW_CONSOLE_OAUTH_LOGOUT_URL="http://localhost:8080/oauth/logout"
TTN_LW_CONSOLE_OAUTH_TOKEN_URL="http://localhost:8080/oauth/token"
TTN_LW_IS_OAUTH_UI_CANONICAL_URL="http://localhost:8080/oauth"
TTN_LW_IS_EMAIL_NETWORK_IDENTITY_SERVER_URL="http://localhost:8080/oauth.js"
TTN_LW_CONSOLE_UI_ASSETS_BASE_URL="http://localhost:8080/assets"
```

#### Optional Configuration

Disable [Hot Module Replacement](https://webpack.js.org/concepts/hot-module-replacement/)

```bash
WEBPACK_DEV_SERVER_DISABLE_HMR="true"
```

Enable TLS in `webpack-dev-server`, using the key and certificate set via `TTN_LW_TLS_KEY` and `TTN_LW_TLS_CERTIFICATE` environment variables. Useful when developing functionalities that rely on TLS.

```bash
WEBPACK_DEV_SERVER_USE_TLS="true"
```
Note: To use this option, The Things Stack for LoRaWAN must be properly setup for TLS. You can obtain more information about this in the **Getting Started** section of the The Things Stack for LoRaWAN documentation.

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

We use [`revive`](http://github.com/mgechev/revive) to lint Go code and [`eslint`](https://eslint.org) to lint JavaScript code. These tools should automatically be installed when initializing your development environment.

### Documentation Site

Please respect the following guidelines for content in our documentation site:

- The title of a doc page is already rendered by the build system as a h1, don't add an extra one.
- Use title case for headings.
- A documentation page starts with an introduction, and then the first heading. The first paragraph of the introduction is typically a summary of the page. Use a `<!--more-->` to indicate where the summary ends.
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
- Is the component generated by higher order components (e.g. `withFeatureRequirement`)?
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
- `withFeatureRequirement` HOC is used to prevent access to routes that the user has no rights for
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
$ ./mage js:translations
```

> Note: When using `./mage js:serve`, this command will be run automatically after any change.

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
$ ./mage js:build
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

A new version can be released from the `master` branch or a `backport` branch. The necessary steps for each are detailed below.

> Note: To get the target version, you can run `version=$(./mage version:bumpXXX version:current)`, where xxx is the type of new release (minor/patch/RC). Check the section [Version Bump](#version-bump) for more information.

### Release From Master

1. Create a `release/${version}` branch off the `master` branch.
```bash
$ git checkout master
$ git checkout -b release/${version}
```
2. Update the `CHANGELOG.md` file as explained in the [Changelog Update](#changelog-update) section.
Once complete, you can add the file to staging
```bash
$ git add CHANGELOG.md
```
3. If releasing a new minor version, update the `SECURITY.md` file and stage it for commit.
```bash
$ git add SECURITY.md
```
4. Bump version as explained in the section [Version Bump](#version-bump).
5. Create a pull request targeting `master`.
6. Once this PR is approved and merged, checkout the latest  `master` branch locally.
7. Create a version tag as explained in the section [Version Tag](#version-tag).
8. Push the version tag. Once this is done, CI automatically starts building and pushing to package managers. When this is done, you'll find a new release on the [releases page](https://github.com/TheThingsNetwork/lorawan-stack/releases).
```bash
$ git push origin ${version}
```
9. Edit the release notes on the Github releases page, which is typically copied from `CHANGELOG.md`.
10. For non RC releases, tag the Docker latest tag as explained in the section [Docker Latest Tag](#docker-latest-tag).

### Release Backports

1. Create a `release/<version>` branch off the `backport/<minor>` branch.
```bash
$ git checkout backport/<minor>
$ git checkout -b release/${version}
```
2. Cherry pick the necessary commits.
```bash
$ git cherrypick <commit>
```
3. Update the `CHANGELOG.md` file as explained in the section [Changelog Update](#changelog-update). Once complete, you can add the file to staging.
```bash
$ git add CHANGELOG.md
```
4. Bump version as explained in the section [Version Bump](#version-bump).
5. Create a pull request targeting `backport/<minor>`.
6. Once this PR is approved and merged, checkout the latest  `backport/<minor>` branch locally.
7. Create a version tag as explained in the section [Version Tag](#version-tag).
8. Push the version tag. Once this is done, CI automatically starts building and pushing to package managers. When this is done, you'll find a new release on the [releases page](https://github.com/TheThingsNetwork/lorawan-stack/releases).
```bash
$ git push origin ${version}
```
9. Edit the release notes on the Github releases page, which is typically copied from `CHANGELOG.md`.

###  Changelog Update

Updating the `CHANGELOG.md` consists of the following steps:
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

### Version Bump

This involves the following three steps

1. Bump

Our development tooling helps with this process. The `mage` command has the following commands for version bumps:
```bash
$ ./mage version:bumpMajor   # bumps a major version (from 3.4.5 -> 4.0.0).
$ ./mage version:bumpMinor   # bumps a minor version (from 3.4.5 -> 3.5.0).
$ ./mage version:bumpPatch   # bumps a patch version (from 3.4.5 -> 3.4.6).
$ ./mage version:bumpRC      # bumps a release candidate version (from 3.4.5-rc1 -> 3.4.5-rc2).
$ ./mage version:bumpRelease # bumps a pre-release to a release version (from 3.4.5-rc1 -> 3.4.5).
```
> Note: These bumps can be combined (i.e. `version:bumpMinor version:bumpRC` bumps 3.4.5 -> 3.5.0-rc1).
2. Write the version files

There are a few files that need to contain the latest version. The new version can be written using
```bash
$ ./mage version:files
```
3. Commit the version bump

A bump commit can be created by running
```bash
$ ./mage version:commitBump
```

> Note: The steps above can be combined to a single command (i.e., `$ ./mage version:bumpPatch version:files version:commitBump`).

### Version Tag

To tag a new version run
```bash
$ ./mage version:bumpXXX version:tag
```

For RCs, make sure to use the same bumping combination (ex: `version:bumpXXX version:bumpYYY`) as used in the bump step above.

### Docker Latest Tag

When the CI system pushed the Docker image, it gets tagged as the current minor and patch version. If this release is not a backport but a latest stable one, you should manually tag and push `latest`:

```bash
$ versionDockerTag=${version#"v"} # v3.6.1 -> 3.6.1
$ docker pull thethingsnetwork/lorawan-stack:${versionDockerTag}
$ docker tag thethingsnetwork/lorawan-stack:{versionDockerTag} thethingsnetwork/lorawan-stack:latest
$ docker push thethingsnetwork/lorawan-stack:latest
```

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

##### Missing restart

The stack has not been restarted after the Console bundle has changed. In production mode, The Things Stack will access the bundle via a filename that contains a content-hash, which is set during the build process of the Console. The hash cannot be updated during runtime and will take effect only after a restart.

##### Possible solution

  1. Restart the The Things Stack

##### Accidentally deleted bundle files

The bundle files have been deleted. This might happen e.g. when a mage target encountered an error and quit before running through.

##### Possible solution

  1. Rebuild the Console `./mage js:clean js:build`
  2. Restart the The Things Stack

##### Mixing up production and development builds

If you switch between production and development builds of the Console, you might forget to re-run the build process and to restart The Things Stack. Likewise, you might have arbitrary config options set that are specific to a respective build type.

##### Possible solution

  1. Double check whether you have set the correct environment: `echo $NODE_ENV`, it should be either `production` or `development`
  2. Double check whether [your The Things Stack config](#development-configuration) is set correctly (especially `TTN_LW_CONSOLE_UI_JS_FILE`, `TTN_LW_CONSOLE_UI_CANONICAL_URL` and similar settings)
  3. Make sure to rebuild the Console `./mage js:clean js:build`
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

or

```
ERROR in ./node_modules/redux-logic/node_modules/rxjs/operators/index.js Module not found: Error: Can't resolve '../internal/operators/audit' in '/lorawan-stack/node_modules/redux-logic/node_modules/rxjs/operators'
```

#### Possible causes

##### Bundle using old JS SDK

The bundle integrates an old version of the JS SDK. This is likely a caching/linking issue of the JS SDK dependency.

##### Possible solutions

- Re-establish a proper module link between the Console and the JS SDK
  - Run `./mage js:cleanDeps js:deps`
  - Check whether the `ttn-lw` symlink exists inside `node_modules` and whether it points to the right destination: `lorawan-stack/sdk/js/dist`
    - If you have cloned multiple `lorawan-stack` forks in different locations, `yarn link` might associate the JS SDK module with the SDK on another ttn repository
  - Rebuild the Console and (only after the build has finished) restart The Things Stack

##### Broken yarn or npm cache

##### Possible solutions

- Clear yarn cache: `yarn cache clear`
- Clear npm cache: `npm cache clear`
- Clean and reinstall dependencies: `./mage js:cleanDeps js:deps`

#### Problem: The build crashes without showing any helpful error message

#### Cause: Not running mage in verbose mode

`./mage` runs in silent mode by default. In verbose mode, you might get more helpful error messages

#### Solution

Run mage in verbose mode: `./mage -v {target}`

#### Problem: Browser displays error: `Cannot GET /`

##### Cause: No endpoint is exposed at root

##### Solution:

Console is typically exposed at `http://localhost:8080/console`,
API at `http://localhost:8080/console`,
OAuth at `http://localhost:8080/oauth`,
etc

#### Problem: Browser displays error: `Error occurred while trying to proxy to: localhost:8080/console`

##### Cause: Stack is not available or not running

For development, remember to run the stack with `go run`:

```bash
$ go run ./cmd/ttn-lw-stack start
```

#### General advice

A lot of problems during build stem from fragmented, incomplete runs of mage targets (due to arbitrary errors happening during a run). Oftentimes, it then helps to build the entire Web UI from scratch: `./mage jsSDK:cleanDeps jsSDK:clean js:cleanDeps js:clean js:build`, and (re-)start The Things Stack after running this.
