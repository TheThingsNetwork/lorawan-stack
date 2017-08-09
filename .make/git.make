# Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

# This makefile contains variables related to the git commit, branch and tags
# as well as rules to execute git hooks
#
# You can override the following variables from this file:
# PRE_COMMIT: space-separated list of rules that need to be executed on every commit
# PRE_PUSH: space-separated list of rules that need to be executed on every push
# COMMIT_MSG: space-separated list of rules that need to be executed to check commit messages
# PREFIXES: space-separated list of allow commit message topics

.PHONY: git.hooks

GIT_COMMIT = `git rev-parse HEAD 2>/dev/null`
GIT_BRANCH = `git rev-parse --abbrev-ref HEAD 2>/dev/null`
GIT_TAG = $(shell git describe --abbrev=0 --tags 2>/dev/null)
BUILD_DATE = `date -u +%Y-%m-%dT%H:%M%SZ`

GIT_RELATIVE_DIR=git rev-parse --show-prefix

only_existing = (xargs ls -d 2>/dev/null || true)
dot_prefixed = sed 's:^:./:'

# All files that are not ignored by git
ALL_FILES ?= (git ls-files . && git ls-files . --exclude-standard --others) | $(only_existing) | $(dot_prefixed)

# Get all files that are currently staged, except for deleted files
STAGED_FILES = git diff --staged --name-only --diff-filter=d --relative=$$($(GIT_RELATIVE_DIR)) | $(only_existing) | $(dot_prefixed)

# filter only staged files from a list
only_staged = sort | comm -12 - <($(STAGED_FILES) | sort)


_hooks = pre-commit pre-push commit-msg

# install git hooks
git.hooks:
	@for hook in $(_hooks); do \
		touch ".git/hooks/$$hook"; \
		chmod u+x ".git/hooks/$$hook"; \
		grep "make" ".git/hooks/$$hook" >/dev/null || $(log) "installing git hook: $$hook" && echo 'ARGS="$$1" make git.'$$hook >> ".git/hooks/$$hook"; \
	done

# remove git hooks
git.hooks.remove:
	@for hook in $(_hooks); do \
		rm ".git/hooks/$$hook" 2>/dev/null && $(log) "removed git hook: $$hook" || true; \
	done

# pre-commit
_git.pre-commit-noop:
	@$(warn) "warning: no pre-commit hooks set, add them by overriding PRE_COMMIT in your makefile"

PRE_COMMIT ?= _git.pre-commit-noop

git.pre-commit: $(PRE_COMMIT)

# pre-push
_git.pre-push-noop:
	@$(warn) "warning: no pre-push hooks set, add them by overriding PRE_PUSH in your makefile"

PRE_PUSH ?= _git.pre-push-noop

git.pre-push: $(PRE_PUSH)

# commit-msg
_git.commit-msg-noop:
	@$(warn) "warning: no commit-msg hooks set, add them by overriding COMMIT_MSG in your makefile"

COMMIT_MSG ?= _git.commit-msg-noop

git.commit-msg: $(COMMIT_MSG)

# prefixes for commit messages
PREFIXES ?= gs ns as is webui util doc make vendor dev

# the args of the commit hook
ARGS ?= /dev/null

# check the commit message to have a prefix
git.commit-msg-log:
	@$(log) "checking commit message"

git.commit-msg-prefix:
	@ok=0; \
	for prefix in `echo $(PREFIXES)`; do \
		cat $(ARGS) | grep -q '^\(fixup! \)\?\(.*,\)\?'$$prefix'\(,.*\)\?: ' && ok=1 || true; \
	done; \
	if [[ $$ok -ne 1 ]]; then \
		$(err) "commit messages should start with a topic from: $(PREFIXES)"; \
		exit 1; \
	fi

# check the commit message to be no longer thant 50 chars
git.commit-msg-length:
	@if [[ `head -n 1 $(ARGS) | sed 's/fixup! //' | wc -c` -gt 50 ]]; then \
		$(err) "commit messages should be shorter than 50 characters"; \
	fi

# check the commit message to not be empty
git.commit-msg-empty:
	@if [[ `head -n 1 $(ARGS) | wc -c` -le 0 ]]; then \
		$(err) "commit messages cannot be empty"; \
	fi

# check if the commit message ends with punctuation
git.commit-msg-phrase:
	@grep -q '[.,?]$$' $(ARGS) && $(warn) "commit messages should not end with punctuation" || true

# check if the commit message begins with a capital letter
git.commit-msg-casing:
	@grep -q '.*: [A-Z]' $(ARGS) && $(warn) "commit messages should be lower case" || true

# vim: ft=make
