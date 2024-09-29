// Copyright 2015 KhulnaSoft, Ltd
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

package kengine

import (
	"os"
	"path/filepath"
	"runtime"
)

// AssetsPath returns the path to the folder
// where the application may store data. If
// KENGINEPATH env variable is set, that value
// is used. Otherwise, the path is the result
// of evaluating "$HOME/.kengine".
func AssetsPath() string {
	if kenginePath := os.Getenv("KENGINEPATH"); kenginePath != "" {
		return kenginePath
	}
	return filepath.Join(userHomeDir(), ".kengine")
}

// userHomeDir returns the user's home directory according to
// environment variables.
//
// Credit: http://stackoverflow.com/a/7922977/1048862
func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}