// Copyright The Helm Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !windows

package kubehclpath

import (
	"runtime"
	"testing"

	"kubehcl.sh/kubehcl/internal/kubehclpath/xdg"
)

func TestKubehclHome(t *testing.T) {
	t.Setenv(xdg.CacheHomeEnvVar, "/cache")
	t.Setenv(xdg.ConfigHomeEnvVar, "/config")
	t.Setenv(xdg.DataHomeEnvVar, "/data")
	isEq := func(t *testing.T, got, expected string) {
		t.Helper()
		if expected != got {
			t.Error(runtime.GOOS)
			t.Errorf("Expected %q, got %q", expected, got)
		}
	}

	isEq(t, CachePath(), "/cache/kubehcl")
	isEq(t, ConfigPath(), "/config/kubehcl")
	isEq(t, DataPath(), "/data/kubehcl")

	// test to see if lazy-loading environment variables at runtime works
	t.Setenv(xdg.CacheHomeEnvVar, "/cache2")

	isEq(t, CachePath(), "/cache2/kubehcl")
}
