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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheThingsIndustries/magepkg/git"
	"github.com/magefile/mage/mg"
	"golang.org/x/xerrors"
)

// Git namespace.
type Git mg.Namespace

func (Git) installHook(name string) (err error) {
	if mg.Verbose() {
		fmt.Printf("Installing %s hook\n", name)
	}
	f, err := os.OpenFile(filepath.Join(".git", "hooks", name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	if _, err = fmt.Fprintf(f, "ARGS=\"$@\" make git.%s\n", name); err != nil {
		return err
	}
	return nil
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
	"dev",
	"doc",
	"dtc",
	"gcs",
	"gs",
	"is",
	"js",
	"ns",
	"oauth",
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
		return xerrors.New("commit message must not be empty")
	}

	if strings.HasPrefix(commitMsg, "fixup! ") || strings.HasPrefix(commitMsg, "Merge ") {
		return nil
	}

	// Check length:
	switch {
	case len(commitMsg) > 72:
		return xerrors.New("commit message must be shorter than 72 characters")
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

// RunHook runs the Git hook for $HOOK, arguments are taken from $ARGS.
func (g Git) RunHook() error {
	hook, args := os.Getenv("HOOK"), strings.Fields(os.Getenv("ARGS"))
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
		return nil
	default:
		return fmt.Errorf("Unknown hook %s", hook)
	}
}
