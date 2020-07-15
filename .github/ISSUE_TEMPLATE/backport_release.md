---
name: Backport Release
about: Checklist for Backport releases

---

<!--
Please check items along as you follow the release process.
-->

#### Overview

This is a checklist for backport releases. This is filled in by both the releaser and the reviewer where necessary.

#### Preparation

- [ ] Create a release/<version> branch off the backport/<minor> branch
  ```bash
  $ git checkout backport/<minor>
  $ git checkout -b release/${version}
  ```

- [ ] Cherry pick the necessary commits.
  ```bash
  $ git cherrypick <commit>
  ```

- [ ] Update the `CHANGELOG.md` file
  - [ ] Change the **Unreleased** section to the new version and add date obtained via `date +%Y-%m-%d` (e.g. `## [3.2.1] - 2019-10-11`)
  - [ ] Check if we didn't forget anything important
  - [ ] Remove empty subsections
  - [ ] Update the list of links in the bottom of the file
  - [ ] Add new **Unreleased** section:
    ```md
    ## [Unreleased]

    ### Added

    ### Changed

    ### Deprecated

    ### Removed

    ### Fixed

    ### Security
    ```

- [ ] Once complete, you can add the file to staging
  ```bash
  $ git add CHANGELOG.md
  ```

- [ ] Bump version
  - [ ] Run the necessary `mage` bump commands based on the type of release
    ```bash
    $ tools/bin/mage version:bumpMajor   # bumps a major version (from 3.4.5 -> 4.0.0).
    $ tools/bin/mage version:bumpMinor   # bumps a minor version (from 3.4.5 -> 3.5.0).
    $ tools/bin/mage version:bumpPatch   # bumps a patch version (from 3.4.5 -> 3.4.6).
    $ tools/bin/mage version:bumpRC      # bumps a release candidate version (from 3.4.5-rc1 -> 3.4.5-rc2).
    $ tools/bin/mage version:bumpRelease # bumps a pre-release to a release version (from 3.4.5-rc1 -> 3.4.5).
    # These bumps can be combined (i.e. `version:bumpMinor version:bumpRC` bumps 3.4.5 -> 3.5.0-rc1).
    ```

  - [ ] Write the version files
    ```bash
    $ tools/bin/mage version:files
    ```

  - [ ] Commit the version bump
    ```bash
    $ tools/bin/mage version:commitBump
    ```

- [ ] Create a pull request targeting `backport/<minor>`.

#### Check 1 (for reviewers)

- [ ] The correct base branch is used.
- [ ] The Changelog is complete i.e., contains only the changes that are in the release (not more/less).
- [ ] The version files are correctly updated.

#### Release

- [ ] Once this PR is approved and merged, checkout the latest  `backport/<minor>` branch locally.
- [ ] Create a version tag
  ```bash
  $ tools/bin/mage version:bumpXXX version:tag
  # For RCs, make sure to use the same bumping combination (ex: `version:bumpXXX version:bumpYYY`) as used in the bump step above.
  ```

- [ ] Push the version tag. Once this is done, CI automatically starts building and pushing to package managers. When this is done, you'll find a new release on the [releases page](https://github.com/TheThingsNetwork/lorawan-stack/releases).
  ```bash
  $ git push origin ${version}
  ```

#### Post Release

- [ ] Edit the release notes on the Github releases page, which is typically copied from `CHANGELOG.md`.

#### Check 2 (for reviewers)

- [ ] The new release contains only the intended commits. This can be checked using `https://github.com/TheThingsNetwork/lorawan-stack/compare/v<previous-version>...v<current-version>`
