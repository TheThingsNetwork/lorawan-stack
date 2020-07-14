---
name: Release
about: Checklist for releases

---

<!--
Please check items along as you follow the release process.
-->

#### Overview

This is a checklist for releases. This is filled in by both the releaser and the reviewer where necessary.

#### Preparation

- [ ] Create a `release/${version}` branch off the `master` branch.
```bash
$ git checkout master
$ git checkout -b release/${version}
```
- [ ] Update the `CHANGELOG.md` file as explained in the [Changelog Update](https://github.com/TheThingsNetwork/lorawan-stack/blob/master/DEVELOPMENT.md#changelog-update) section.
Once complete, you can add the file to staging
```bash
$ git add CHANGELOG.md
```
- [ ] If releasing a new minor version, update the `SECURITY.md` file and stage it for commit.
```bash
$ git add SECURITY.md
```
- [ ] Bump version as explained in the section [Version Bump](https://github.com/TheThingsNetwork/lorawan-stack/blob/master/DEVELOPMENT.md#version-bump).
- [ ] Create a pull request targeting `master`.

#### Check 1 (for reviewers)
- [ ] The Changelog is complete i.e., contains only the changes that are in the release (not more/less).
- [ ] `SECURITY.md` is updated.
- [ ] The version files are correctly updated.

#### Release
- [ ] Once this PR is approved and merged, checkout the latest  `master` branch locally.
- [ ] Create a version tag as explained in the section [Version Tag](https://github.com/TheThingsNetwork/lorawan-stack/blob/master/DEVELOPMENT.md#version-tag).
- [ ] Push the version tag. Once this is done, CI automatically starts building and pushing to package managers. When this is done, you'll find a new release on the [releases page](https://github.com/TheThingsNetwork/lorawan-stack/releases).
```bash
$ git push origin ${version}
```

### Post Release

- [ ] Edit the release notes on the Github releases page, which is typically copied from `CHANGELOG.md`.
- [ ] For non RC releases, tag the Docker latest tag as explained in the section [Docker Latest Tag](https://github.com/TheThingsNetwork/lorawan-stack/blob/master/DEVELOPMENT.md#docker-latest-tag).

#### Check 2 (for reviewers)

- [ ] The new release contains only the intended commits. This can be checked using `https://github.com/TheThingsNetwork/lorawan-stack/compare/v<previous-version>...v<current-version>`
- [ ] The Docker latest tag is up to date.
