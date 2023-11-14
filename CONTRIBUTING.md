# Contributing to The Things Stack for LoRaWAN

Thank you for your interest in building this thing together with us. We're really happy with our active community and are glad that you're a part of it. There are many ways to contribute to our project, but given the fact that you're on Github looking at the code for The Things Stack for LoRaWAN, you're probably here for one of the following reasons:

- **Asking a question**: If you have questions, please use the [forum](https://www.thethingsnetwork.org/forum/). We have a special [category for The Things Stack](https://www.thethingsnetwork.org/forum/c/ttn-network/v3/90).
- **Requesting a new feature**: If you have a great idea or think some functionality is missing, we want to know! The only thing you have to do for that is to [create an issue](https://github.com/TheThingsNetwork/lorawan-stack/issues) if it doesn't exist yet. Please use the issue template and fill out all sections.
- **Reporting an issue**: If you notice that a component of The Things Stack is not behaving as it should, there may be a bug in our systems. In this case you should [create an issue](https://github.com/TheThingsNetwork/lorawan-stack/issues) if it doesn't exist yet. Please use the issue template and fill out all sections. For sensitive (security) issues, you can [contact us directly](#security-issues).
- **Implementing a new feature or fixing a bug**: If you see an [open issue](https://github.com/TheThingsNetwork/lorawan-stack/issues) that you would like to work on, let the other contributors know by commenting in the issue.
- **Writing documentation**: If you see that our documentation is lacking or incorrect, it would be great if you could help us improve it. This will help users and fellow contributors understand how to better work with our stack. Better documentation helps prevent making mistakes and introducing new bugs. Our documentation is spread across a number of places. Code documentation obviously lives together with the code, and is therefore probably in this repository. User documentation that is published on [the official The Things Stack documentation page](https://thethingsstack.io), is built from the [`lorawan-stack-docs` repository](https://github.com/TheThingsIndustries/lorawan-stack-docs). More general documentation can be found on [The Things Network's official documentation pages](https://www.thethingsnetwork.org/docs). The source files for that documentation can be found in [the `docs` repository](https://github.com/TheThingsNetwork/docs).

If you'd like to contribute by writing code, you'll find [here](DEVELOPMENT.md) how to set up your development environment.

Style guidelines for contributing code and writing documentation can be found [here](DEVELOPMENT.md#code-style).

## Compatibility Commitment

The [compatibility commitment](README.md#commitments-and-releases) ensures that users can update confidently and with little hassle to patch releases.

If, according to the compatibility commitment, we need to bump the minor version, a pull request should target the `v3.n+1` branch (where `n` is the current minor version). Only if the changes in a pull request are fully compatible in terms of API, storage, command-line and configuration, the target can be set to `v3.n` (where `n` is the current minor version).

We use [`CODEOWNERS`](CODEOWNERS) to enforce the compatibility commitment. This requires configuration structure to be defined in `config.go` files, i.e. `/pkg/<subsystem>/config.go`.

## Branching

When contributing code or documentation to this repository, we follow a number of guidelines.

### Branch Naming

All branches shall have one of these names.

- `v3.n`: the default branch for the current minor version `n`. This is a clean branch where reviewed, approved and CI passed pull requests are merged into. Merging to this branch is restricted to project maintainers.
- `v3.n+1`: the default branch for the next minor version `n` + 1. As with `v3.n`, this is a clean branch where reviewed, approved and CI passed pull requests are merged into. Merging to this branch is restricted to project maintainers. The difference with the `v3.n` branch is that this branch may contain "breaking" features that need to go to a different minor version according to our compatibility commitment.
- `fix/#-short-name` or `fix/short-name`: refers to a fix, preferably with issue number. The short name describes the bug or issue.
- `feature/#-short-name` or `feature/short-name`: (main) feature branch, preferably with issue number. The short name describes the feature.
  - `feature/#-short-name-part`: a sub scope of the feature in a separate branch, that is intended to merge into the main feature branch before the main feature branch is merged into a `v3.*` branch.
- `issue/#-short-name`: anything else that refers to an issue but is not clearly a fix nor a feature.

### Scope

A fix, feature or issue branch should be **small and focused** and should be scoped to a **single specific task**. Do not combine new features and refactoring of existing code.

### Pull requests and rebasing

Pull requests shall close or reference issues. Please file an issue first before submitting a pull request. When submitting a pull request, please fill out all the sections in the pull request template.

- **Before** a reviewer is assigned, rebasing the branch to reduce the number of commits is highly advised. We recommend self-reviewing your own pull request: making the [commit](#commit) history clean, checking for typos or incoherences, and making sure Continuous Integration passes.
- **During** a pull request's review, do not squash commits: it makes it harder for reviewers to read the evolution of a pull request. Making the commit history denser to answer reviewers' comments is acceptable at that point.
- Once a pull request **has been approved** by the reviewers, it can be rebased on top of its target branch before it is merged. This is an opportunity for the contributor to clean up the commit history. A reviewer can also ask specifically for a rebase.

## Commits

Keep the commits to be merged clean: adhere to the commit message format defined below and instead of adding and deleting files within a pull request, drop or fix the concerning commit that added the file.

Interactive rebase (`git rebase -i`) can be used to edit or rewrite commits that do not follow these contribution guidelines.

## <a name="commit"></a>Commit Messages

The first line of a commit message is the subject. The commit message may contain a body, separated from the subject by an empty line.

### Commit Subject

The subject contains the concerning topic and a concise message in [the imperative mood](https://chris.beams.io/posts/git-commit/#imperative), starting with a capital. The subject may also contain references to issues or other resources.

The topic is typically a few characters long and should always be present. Accepted topics are:

- `account`: The Account App including backend and frontend
- `all`: Changes affecting all code, e.g. primitive types
- `api`: API, typically protos
- `as`: The Application Server component
- `ci`: Continuous Integration tooling
- `cli`: Command-line Interface
- `console`: The Console component including backend and frontend
- `cs`: The Crypto Server component
- `data`: Data from external repositories
- `dcs`: The Device Claiming Server component
- `dev`: Development tooling
- `dr`: The Device Repository component
- `dtc`: The Device Template Converter component
- `es`: The Events Server component
- `gcs`: The Gateway Configuration Server component
- `gs`: The Gateway Server component
- `is`: The Identity Server component
- `js`: The Join Server component
- `noc`: The Network Operations Center component
- `ns`: The Network Server component
- `pba`: The Packet Broker Agent component
- `qrg`: The QR Code Generator component
- `tbs`: The Tenant Billing Server component
- `util`: Utilities

Changes that affect multiple components can be comma separated.

Good commit messages:

- `ns: Fix MIC check`
- `dev: Set version from git tag`
- `ns,as,gs: Fix TLS check`

Make sure that commits are scoped to something meaningful and could potentially be cherry-picked individually.

### Commit Body

The body may contain a more detailed description of the commit, explaining what it changes and why. The "how" is less relevant, as this should be obvious from the diff.

## Release notes

We maintain a changelog at `CHANGELOG.md` using a format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

Any notes that we need to include in the Release Notes for the next release should be added under the `Unreleased` section.

Please consult documentation at [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) on how to write these notes.

Some key points:

- Notes are formatted as bullet points, written in imperative tense and finish with a dot (`.`).
- There are six possible types of changes, always listed in following order:
  - `Added` for new features.
  - `Changed` for changes in existing functionality.
  - `Deprecated` for soon-to-be removed features.
  - `Removed` for now removed features.
  - `Fixed` for any bug fixes.
  - `Security` in case of vulnerabilities.

## <a name="security-issues"></a>Security Issues

We do our utmost best to build secure systems, but we're human too, so we sometimes make mistakes. If you find any vulnerability in our systems, please report it via our [responsible disclosure process](https://www.thethingsnetwork.org/responsible-disclosure)

## Legal

The Things Stack for LoRaWAN is Apache 2.0 licensed.
