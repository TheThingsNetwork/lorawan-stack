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

package util

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// NormalizedFlagSet returns a flagset with a NormalizeFunc that replaces underscores to dashes.
func NormalizedFlagSet() *pflag.FlagSet {
	fs := &pflag.FlagSet{}
	fs.SetNormalizeFunc(NormalizeFlags)
	return fs
}

// DeprecateFlag deprecates a CLI flag.
func DeprecateFlag(flagSet *pflag.FlagSet, old string, new string) {
	if newFlag := flagSet.Lookup(new); newFlag != nil {
		deprecated := *newFlag
		deprecated.Name = old
		deprecated.Usage = strings.Replace(deprecated.Usage, old, new, -1)
		deprecated.Deprecated = fmt.Sprintf("use the %s flag", new)
		deprecated.Hidden = true
		flagSet.AddFlag(&deprecated)
	}
}

// DeprecateWithoutForwarding deprecates a CLI flag without forwarding it to a new flag.
func DeprecateWithoutForwarding(flagSet *pflag.FlagSet, flag string, reason string) {
	if flag := flagSet.Lookup(flag); flag != nil {
		flag.Deprecated = reason
		flag.Hidden = true
	}
}

// ForwardFlag forwards the flag value of old to new if new is not set while old is.
func ForwardFlag(flagSet *pflag.FlagSet, old string, new string) {
	if oldFlag := flagSet.Lookup(old); oldFlag != nil && oldFlag.Changed {
		if newFlag := flagSet.Lookup(new); newFlag != nil && !newFlag.Changed {
			flagSet.Set(new, oldFlag.Value.String())
		}
	}
}

// HideFlag hides the provided flag from the flag set.
func HideFlag(flagSet *pflag.FlagSet, name string) {
	if flag := flagSet.Lookup(name); flag != nil {
		flag.Hidden = true
	}
}

// HideFlagSet hides the flags from the provided flag set.
func HideFlagSet(flagSet *pflag.FlagSet) *pflag.FlagSet {
	flagSet.VisitAll(func(f *pflag.Flag) {
		f.Hidden = true
	})
	return flagSet
}

var (
	toDash       = strings.NewReplacer("_", "-")
	toUnderscore = strings.NewReplacer("-", "_")
)

// NormalizePaths converts arguments to field mask paths, replacing '-' with '_'
func NormalizePaths(paths []string) []string {
	normalized := make([]string, len(paths))
	for i, path := range paths {
		normalized[i] = toUnderscore.Replace(path)
	}
	return normalized
}

func NormalizeFlags(f *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(toDash.Replace(name))
}

func SelectFieldMask(cmdFlags *pflag.FlagSet, fieldMaskFlags ...*pflag.FlagSet) (paths []string) {
	if all, _ := cmdFlags.GetBool("all"); all {
		for _, fieldMaskFlags := range fieldMaskFlags {
			fieldMaskFlags.VisitAll(func(flag *pflag.Flag) {
				paths = append(paths, toUnderscore.Replace(flag.Name))
			})
		}
		return
	}
	cmdFlags.Visit(func(flag *pflag.Flag) {
		flagName := toUnderscore.Replace(flag.Name)
		for _, fieldMaskFlags := range fieldMaskFlags {
			if b, err := fieldMaskFlags.GetBool(flag.Name); err == nil && b {
				paths = append(paths, flagName)
				return
			}
		}
	})
	return
}

func UpdateFieldMask(cmdFlags *pflag.FlagSet, fieldMaskFlags ...*pflag.FlagSet) (paths []string) {
	cmdFlags.Visit(func(flag *pflag.Flag) {
		flagName := toUnderscore.Replace(flag.Name)
		for _, fieldMaskFlags := range fieldMaskFlags {
			if fieldMaskFlags.Lookup(flagName) != nil {
				paths = append(paths, flagName)
				return
			}
		}
	})
	return
}

// SelectAllFlagSet returns a flagset with the --all flag
func SelectAllFlagSet(what string) *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Bool("all", false, fmt.Sprintf("select all %s fields", what))
	return flagSet
}

// UnsetFlagSet returns a flagset with the --unset flag
func UnsetFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringSlice("unset", []string{}, "list of fields to unset")
	return flagSet
}
