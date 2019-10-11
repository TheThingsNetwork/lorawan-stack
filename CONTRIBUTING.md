# Contributing to The Things Stack for LoRaWAN

Thank you for your interest in building this thing together with us. We're really happy with our active community and are glad that you're a part of it. There are many ways to contribute to our project, but given the fact that you're on Github looking at the code for The Things Stack for LoRaWAN, you're probably here for one of the following reasons:

* **Requesting a new feature**: If you have a great idea or think some functionality is missing, we want to know! The only thing you have to do for that is to [create an issue](https://github.com/TheThingsNetwork/lorawan-stack/issues) if it doesn't exist yet. Please use the issue template and fill out all sections.
* **Reporting an issue**: If you notice that a component of The Things Stack is not behaving as it should, there may be a bug in our systems. In this case you should [create an issue](https://github.com/TheThingsNetwork/lorawan-stack/issues) if it doesn't exist yet. Please use the issue template and fill out all sections. For sensitive (security) issues, you can [contact us directly](#security-issues).
* **Implementing a new feature or fixing a bug**: If you see an [open issue](https://github.com/TheThingsNetwork/lorawan-stack/issues) that you would like to work on, let the other contributors know by commenting in the issue.
* **Writing documentation**: If you see that our documentation is lacking or incorrect, it would be great if you could help us improve it. This will help users and fellow contributors understand better how to work with our stack, and will prevent making mistakes and introducing bugs. Our documentation is spread across a number of places. Code documentation obviously lives together with the code, and is therefore probably in this repository. More general documentation lives in [our `docs` repo](https://github.com/TheThingsNetwork/docs), that is published to our [official documentation pages](https://www.thethingsnetwork.org/docs).

Although The Things Network forums and community Slack are great ways to communicate with other community members, please use GitHub issues and pull requests to discuss matters related to the code base.

If you'd like to contribute by writing code, you'll find [here](DEVELOPMENT.md) how to set up your development environment. We also have some guidelines that describe how to make contributions that are consistent our way of working.

+ [Git branching workflow](#branching)
+ [Commit conventions](#commit)
+ [Coding conventions](#code)

## <a name="branching"></a>Branching

### Naming

All branches shall have one of these names.

- `master`: the default branch. This is a clean branch where reviewed, approved and CI passed pull requests are merged into. Merging to this branch is restricted to project maintainers
- `fix/#-short-name` or `fix/short-name`: refers to a fix, preferably with issue number. The short name describes the bug or issue
- `feature/#-short-name` or `feature/short-name`: (main) feature branch, preferably with issue number. The short name describes the feature
  - `feature/#-short-name-part`: a sub scope of the feature in a separate branch, that is intended to merge into the main feature branch before the main feature branch is merged into `master`
- `issue/#-short-name`: anything else that refers to an issue but is not clearly a fix nor a feature

### Scope

A fix, feature or issue branch should be **small and focused** and should be scoped to a **single specific task**. Do not combine new features and refactoring of existing code.

### Pull requests and rebasing

Pull requests shall close or reference issues. Please file an issue first before submitting a pull request. When submitting a pull request, please fill out all the sections in the pull request template.

+ **Before** a reviewer is assigned, rebasing the branch to reduce the number of commits is highly advised. We recommend self-reviewing your own pull request: making the [commit](#commit) history clean, checking for typos or incoherences, and making sure Continuous Integration passes.

+ **During** a pull request's review, do not squash commits: it makes it harder for reviewers to read the evolution of a pull request. Making the commit history denser to answer reviewers' comments is acceptable at that point.

+ Once a pull request **has been approved** by the reviewers, it can be rebased on top of its target branch before it is merged. This is an opportunity for the contributor to clean up the commit history. A reviewer can also ask specifically for a rebase.

Keep the commits to be merged clean: adhere to the commit message format defined below and instead of adding and deleting files within a pull request, drop or fix the concerning commit that added the file.

Interactive rebase (`git rebase -i`) can be used to rewrite commit messages that do not follow these contribution guidelines.

## <a name="commit"></a>Commit Messages

The first line of a commit message is the subject. The commit message may contain a body, separated from the subject by an empty line.

### Subject

The subject contains the concerning component or topic and a concise message in [the imperative mood](https://chris.beams.io/posts/git-commit/#imperative), starting with a capital. The subject may also contain references to issues or other resources.

The component or topic is typically a few characters long and should always be present. Component names are:

- `api`: API, typically protos
- `gs`: Gateway Server
- `ns`: Network Server
- `as`: Application Server
- `is`: Identity Server
- `console`: Console
- `cli`: Command-line Interface
- `util`: utilities
- `ci`: CI instructions, e.g. Travis file
- `doc`: documentation
- `dev`: other non-functional development changes, e.g. Makefile, .gitignore, editor config
- `all`: changes affecting all code, e.g. primitive types

Changes that affect multiple components can be comma separated.

Good commit messages:
- `ns: Fix MIC check`
- `dev: Set version from git tag, closes #123`
- `ns,as,gs: Fix TLS check`

Make sure that commits are scoped to something meaningful and could, potentially, be merged individually.

### Body

The body may contain a more detailed description of the commit, explaining what it changes and why. The "how" is less relevant, as this should be obvious from the diff.

## <a name="code"></a>Code

### Formatting

We want our code to be consistent across our projects, so we'll have to agree on a number of formatting rules. These rules should usually usually be applied by your editor. Make sure to install the [editorconfig](https://editorconfig.org) plugin for your editor.

Go code can be automatically formatted using the [`gofmt`](https://golang.org/cmd/gofmt/) tool. There are many editor plugins that call `gofmt` when you save your files.

#### General

We use **utf-8**, **LF** line endings, a **final newline** and we **trim whitespace** from the end of the line (except in Markdown).

#### Tabs vs Spaces

Many developers have strong opinions about using tabs vs spaces. We apply the following rules:

- All `.go` files are indented using **tabs**
- The `Makefile` and all `.make` files are indented using **tabs**
- All other files are indented using **two spaces**

#### Line length

- If a line is longer than 80 columns, try to find a "natural" break
- If a line is longer than 120 columns, insert a line break
- In very special cases, longer lines are tolerated

### Linting

We use [`golint`](github.com/golang/lint/golint) to lint `.go` files and [`eslint`](https://eslint.org) to lint `.js` and `.vue` files. These tools should automatically be installed when initializing your development environment.

### API methods naming

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

### Variable naming

Variable names should be short and concise.

We follow the [official go guidelines](https://github.com/golang/go/wiki/CodeReviewComments#variable-names) and try to be consistent with Go standard library as much as possible, everything not defined in the tables below should follow Go standard library naming scheme. In general, variable names are English and descriptive, omitting abbreviations as much as possible (except for the tables below), as well as putting adjectives and adverbs before the noun and verb respectively.

#### Single-word entities

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
| user                 | user    |                                                               |
| transmit             | tx / Tx |                                                               |
| receive              | rx / Rx |                                                               |

The EUI naming scheme can be found in the well-known variable names section bellow.

#### 2-word entities

In case both of the words have an implementation-specific meaning, the variable name is the combination of first letter of each word.

| entity                                                  | name    |
| :-----------------------------------------------------: | :-----: |
| wait group                                              | wg      |
| Gateway Server                                          | gs      |
| Network Server                                          | ns      |
| Join Server                                             | js      |
| Application Server                                      | as      |
| Identity Server                                         | is      |

In case one of the words specifies the meaning of the variable in a specific language construct context, the variable name is the combination of abbrevations of the words.

#### Well-known variable names

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
| user ID                         | userID  |

### Events

Events are defined with `events.Define("event_name", "event description")`

The event name is usually of the form `component.entity.action`. Examples are `ns.up.receive_duplicate` and `is.user.update`. We have some exceptions, such as `ns.up.join.forward`, which is specifically used for join messages. The Device Registry (which is used by both the NS, AS and JS) currently publishes events such as `device.create`. In the future we may also add component prefixes there. See below for naming conventions.

The event description describes the event in simple English. The description is capitalized by the frontend, so the message should be lowercase, and typically doesn't end with a period.

### Errors

Errors are defined with ``errors.Define<Type>("error_name", "error description with `{attribute_value}`")``

Error definitions must be defined as close to the return statements as possible; in the same package, and preferably above the concerning function(s). Do not export the error definitions unless they are meaningful to other packages, i.e. for testing the exact error definition.

Prefer using a specific error type, i.e. `errors.DefineInvalidArgument()`. If you are using a cause (using `WithCause()`), you may use `Define()` to fallback to the cause's type.

The error name in snake case is a short and unique identifier of the error within the package. There is no need to append `_failed` or `_error` or prepend `failed_to_` as an error already indicates something went wrong. Be consistent in wording (i.e. prefer the more descriptive `missing_field` over `no_field`), order (i.e. prefer the more clear `missing_field` over `field_missing`) and do not use entity abbreviations.

The error description in lower case, with only names in title case, is a concise plain English text that is human readable and understandable. Do not end the description with a dot. You may use attributes, in snake case, in the description defined between backticks (`` ` ``) by putting the key in curly braces (`{ }`). See below for naming conventions. Only provide primitive types as attribute values using `WithAttributes()`.

### Logging field keys, event and error names, error attribute names and task identifiers

Any `name` defined in the following statements:

- Logging field key: `logger.WithField("name", "value")`
- Event name: `events.Define("name", "description")`
- Error name: `errors.Define("name", "description")`
- Error attribute: ``errors.Define("example", "description `{name}`")``
- Task identifier: `c.RegisterTask("name", ...)`

Shall be snake case, optionally having an event name prepended with a dotted namespace, see above. The spacer `_` shall be used in LoRaWAN terms: `DevAddr` is `dev_addr`, `AppSKey` is `app_s_key`, etc.

### Comments

Code should be as self-explanatory as possible. However, comments should be used to respect Go formatting guidelines, to generate insightful documentation with [Godoc](https://godoc.org/go.thethings.network/lorawan-stack), and to explain what can not be expressed by pure code. Comments should be English sentences, and documentation-generating comments should be closed by a period. Comments can also be used to indicate steps to take in the future (*TODOs*), if they reference the GitHub issue to track this *TODO*.

+ In **Go files**, comments should be added according to `golint` requirements and [Effective Go guidelines](https://golang.org/doc/effective_go.html#commentary), especially in regards to commenting exported packages, types and variables.

+ In **protocol buffer files**, fields should be concisely commented.

+ **Protobuf-generated files**, including Go files, are not tracked by `golint`. These do not need to respect the same comments guidelines as Go files, especially in regards to exported values.

## Translations

We do our best to make all text that could be visible to users available for translation. This means that all text of the console's user interface, as well as all text that it may forward from the backend, needs to be defined in such a way that it can be translated into other languages than English.

### Backend Translations

In the API, the enum descriptions, error messages and event descriptions available for translation. Enum descriptions are defined in `pkg/ttnpb/i18n.go`. Error messages and event descriptions are defined with `errors.Define(...)` and `events.Define(...)` respectively.

These messages are then collected in the `config/messages.json` file, which will be processed in the frontend build process, but may also be used by other (native) user interfaces. When you define new enums, errors or events or when you change them, the messages need to be updated into the `config/messages.json` file.

```sh
./mage go:messages
```

If you forget to do so, this will cause a CI failure.

Adding translations of messages to other languages than English is a matter of adding key/value pairs to `translations` in `config/messages.json`.

### Frontend Translations

Translations of frontend messages are located in `pkg/webui/locales`.

## Release notes
We maintain a changelog at `CHANGELOG.md` using a format based on [Keep a Changelog].

Any notes that we need to include in the Release Notes for the next release should be added under the `Unreleased` section.

Please consult documentation at [Keep a Changelog] on how to write these notes.

Some key points:

- Notes are formatted as bullet points, written in imperative tense and finish with a dot (`.`).
- There are six possible types of changes, always listed in following order:
  - `Added` for new features.
  - `Changed` for changes in existing functionality.
  - `Deprecated` for soon-to-be removed features.
  - `Removed` for now removed features.
  - `Fixed` for any bug fixes.
  - `Security` in case of vulnerabilities.

As part of the release process `Unreleased` section is renamed to the released version. For example, if a version v3.2.1 is released on 2019-10-09, the `Unreleased` section should be renamed to `v3.2.1 - 2019-10-09`, empty change subsections removed and a new `Unreleased` section should be added above looking like this:

```md
## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security
```

## <a name="security-issues"></a>Security Issues

We do our utmost best to build secure systems, but we're human too, so we sometimes make mistakes. If you find any vulnerability in our systems, please contact us directly. We can be reached on Slack, by email and a number of other communication platforms.

- Johan Stokking - [keybase.io/johanstokking](https://keybase.io/johanstokking) `EE80D01EB2BE7EC8`
- Hylke Visser - [keybase.io/htdvisser](https://keybase.io/htdvisser) `A115FF80DC8A2270`

Our email addresses follow the pattern `<firstname>@thethingsnetwork.org`.

## Legal

The Things Stack for LoRaWAN is Apache 2.0 licensed.

[Keep a Changelog]: https://keepachangelog.com/en/1.0.0/
