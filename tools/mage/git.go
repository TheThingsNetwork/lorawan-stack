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

package ttnmage

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TheThingsIndustries/magepkg/git"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Git namespace.
type Git mg.Namespace

func (Git) installHook(name string) (err error) {
	if mg.Verbose() {
		fmt.Printf("Installing %s hook\n", name)
	}
	return os.WriteFile(
		filepath.Join(".git", "hooks", name),
		[]byte(fmt.Sprintf(
			`STDIN="$(cat /dev/stdin)" ARGS="$@" make git.%s`,
			name,
		)),
		0o755,
	)
}

var gitHooks = []string{"pre-commit", "commit-msg", "pre-push"}

// InstallHooks installs git hooks that help developers follow our best practices.
func (g Git) InstallHooks() error {
	for _, hook := range gitHooks {
		if err := g.installHook(hook); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	initDeps = append(initDeps, Git.InstallHooks)
}

// UninstallHooks uninstalls git hooks.
func (g Git) UninstallHooks() error {
	for _, hook := range gitHooks {
		if mg.Verbose() {
			fmt.Printf("Uninstalling %s hook\n", hook)
		}
		if err := os.Remove(filepath.Join(".git", "hooks", hook)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (Git) selectStaged() error {
	staged, err := git.StagedFiles()
	if err != nil {
		return err
	}
	selectedFiles, selectedDirs = make(map[string]bool), make(map[string]bool)
	for _, file := range staged {
		selectedFiles[file] = true
		selectedDirs[filepath.Dir(file)] = true
	}
	return nil
}

var preCommitChecks []interface{}

func (g Git) preCommit() error {
	if mg.Verbose() {
		fmt.Println("Running pre-commit hook")
	}
	mg.Deps(g.selectStaged)
	mg.Deps(preCommitChecks...)
	return nil
}

var gitCommitPrefixes = []string{
	"all",
	"api",
	"as",
	"ci",
	"cli",
	"console",
	"data",
	"dcs",
	"dev",
	"dr",
	"dtc",
	"gcs",
	"gs",
	"is",
	"js",
	"ns",
	"account",
	"pba",
	"qrg",
	"util",
}

func (Git) commitMsg(messageFile string) error {
	if mg.Verbose() {
		fmt.Println("Running commit-msg hook")
	}
	if messageFile == "" {
		messageFile = ".git/COMMIT_EDITMSG"
	}
	f, err := os.Open(messageFile)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Scan()
	commitMsg := s.Text()

	if commitMsg == "" {
		return errors.New("commit message must not be empty")
	}

	if strings.HasPrefix(commitMsg, "fixup! ") || strings.HasPrefix(commitMsg, "Merge ") {
		return nil
	}

	// Check length:
	switch {
	case len(commitMsg) > 72:
		return errors.New("commit message must be shorter than 72 characters")
	case len(commitMsg) > 50:
		// TODO: Warn.
	}

	// Check topics: Message structure:
	split := strings.SplitN(commitMsg, ": ", 2)
	if len(split) != 2 {
		return fmt.Errorf("commit message must contain topics from %s",
			strings.Join(gitCommitPrefixes, ","))
	}

	// Check topics:
	topics := strings.Split(split[0], ",")
	var unknownTopics []string
nextTopic:
	for _, topic := range topics {
		for _, allowed := range gitCommitPrefixes {
			if strings.TrimSpace(topic) == allowed {
				continue nextTopic
			}
		}
		unknownTopics = append(unknownTopics, topic)
	}
	if len(unknownTopics) > 0 {
		return fmt.Errorf("commit messages must only topics from %s (and not %s)",
			strings.Join(gitCommitPrefixes, ","),
			strings.Join(unknownTopics, ","))
	}

	words := strings.Fields(split[1])

	// Casing:
	if words[0][0] < 'A' || words[0][0] > 'Z' {
		return fmt.Errorf("commit messages must start with a capital letter (and %s doesn't)", words[0])
	}

	// Punctuation:
	if strings.HasSuffix(commitMsg, ".") {
		return fmt.Errorf("commit messages must not end with punctuation")
	}

	// Imperative
	if strings.HasSuffix(words[0], "ed") || strings.HasSuffix(words[0], "ing") {
		// TODO: Warn that Commit messages should use imperative mood
	}

	return nil
}

func (g Git) prePush(stdin string, args ...string) error {
	if mg.Verbose() {
		fmt.Println("Running pre-push hook")
	}
	if stdin == "" {
		fmt.Println("Standard input is empty, skip pre-push hook")
		return nil
	}
	var (
		ref  string
		head string
	)
	if ss := strings.Fields(stdin); len(ss) == 4 {
		ref = ss[0]
		head = ss[1]
	} else {
		return fmt.Errorf("expected pre-push hook standard input to contain 4 fields, got: %d(`%s`)", len(ss), ss)
	}
	if len(args) != 2 {
		return fmt.Errorf("pre-push hook expected to get 2 arguments, got: %s", args)
	}
	if head == "0000000000000000000000000000000000000000" {
		// Remote branch is being deleted
		return nil
	}
	const ttiMarkerHash = "f3df41ad99f4acdcb2b038da9a15671023bc827c" // Hash of the first proprietary commit.
	switch err := exec.Command("git", "merge-base", "--is-ancestor", ttiMarkerHash, head).Run().(type) {
	case nil:
	case *exec.ExitError:
		switch n := err.ExitCode(); n {
		case 1:
			return nil
		case 128:
			if mg.Verbose() {
				fmt.Println("Unable to check presence of TTI marker commit: hash not found")
			}
			return nil
		default:
			return fmt.Errorf("expected exit code of 1, got %d", n)
		}
	default:
		return fmt.Errorf("failed to check presence of TTI marker commit `%s`: %s", ttiMarkerHash, err)
	}
	if s := os.Getenv("TTI_REMOTES"); s != "" {
		for _, remote := range strings.Fields(s) {
			if args[1] == remote {
				return nil
			}
		}
	} else {
		switch args[1] {
		case "git@github.com:TheThingsIndustries/lorawan-stack",
			"git@github.com:TheThingsIndustries/lorawan-stack.git",
			"https://github.com/TheThingsIndustries/lorawan-stack",
			"https://github.com/TheThingsIndustries/lorawan-stack.git",
			"ssh://git@github.com:TheThingsIndustries/lorawan-stack",
			"ssh://git@github.com:TheThingsIndustries/lorawan-stack.git":
			return nil
		}
	}
	return fmt.Errorf("trying to push TTI head `%s`(`%s`) to unverified remote `%s`, abort", head, ref, args[1])
}

// RunHook runs the Git hook for $HOOK.
// - standard input contents are taken from $STDIN
// - arguments are taken from $ARGS
func (g Git) RunHook() error {
	hook, stdin, args := os.Getenv("HOOK"), os.Getenv("STDIN"), strings.Fields(os.Getenv("ARGS"))
	switch hook {
	case "pre-commit":
		return g.preCommit()
	case "commit-msg":
		var messageFile string
		if len(args) > 0 {
			messageFile = args[0]
		}
		return g.commitMsg(messageFile)
	case "pre-push":
		if mg.Verbose() {
			fmt.Println("Running pre-push hook with", args)
		}
		return g.prePush(stdin, args...)
	default:
		return fmt.Errorf("Unknown hook %s", hook)
	}
}

// Diff returns error if `git diff` is not empty
func (Git) Diff() error {
	if mg.Verbose() {
		fmt.Println("Checking git diff")
	}
	output, err := sh.Output("git", "diff")
	if err != nil {
		return err
	}
	if output != "" {
		return fmt.Errorf("Previous operations have created changes that were not recorded in the repository. Please make those changes on your local machine before pushing them to the repository:\n%s", output)
	}
	return nil
}

// UpdateSubmodules updates submodules, and initializes them when necessary.
func (Git) UpdateSubmodules() error {
	if mg.Verbose() {
		fmt.Println("Updating submodules")
	}
	_, err := sh.Exec(nil, os.Stdout, os.Stderr, "git", "submodule", "update", "--init")
	return err
}

func init() {
	initDeps = append(initDeps, Git.UpdateSubmodules)
}

// PullSubmodules pulls in submodule updates.
func (Git) PullSubmodules() error {
	if mg.Verbose() {
		fmt.Println("Updating submodules")
	}
	_, err := sh.Exec(nil, os.Stdout, os.Stderr, "git", "submodule", "update", "--init", "--remote")
	return err
}
