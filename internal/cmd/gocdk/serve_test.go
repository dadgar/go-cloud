// Copyright 2019 The Go Cloud Development Kit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBuildForServe(t *testing.T) {
	t.Run("NoWire", func(t *testing.T) {
		testBuildForServe(t, map[string]string{
			"main.go": `package main
import "fmt"
func main() { fmt.Println("Hello, World!") }`,
		})
	})
	t.Run("Wire", func(t *testing.T) {
		if _, err := exec.LookPath("wire"); err != nil {
			// Wire tool is run unconditionally. Needed for both tests.
			t.Skip("wire not found:", err)
		}
		// TODO(light): This test is not hermetic because it brings
		// in an external module.
		testBuildForServe(t, map[string]string{
			"main.go": `package main
import "fmt"
func main() { fmt.Println(greeting()) }`,
			"setup.go": `// +build wireinject

package main
import "github.com/google/wire"
func greeting() string {
	wire.Build(wire.Value("Hello, World!"))
	return ""
}`,
		})
	})
}

func testBuildForServe(t *testing.T, files map[string]string) {
	ctx := context.Background()
	pctx, cleanup, err := newTestProject(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for name, content := range files {
		if err := ioutil.WriteFile(filepath.Join(pctx.workdir, name), []byte(content), 0666); err != nil {
			t.Fatal(err)
		}
	}
	exePath := filepath.Join(pctx.workdir, "hello")
	if runtime.GOOS == "windows" {
		exePath += ".EXE"
	}

	// Build program.
	if err := buildForServe(ctx, pctx, pctx.workdir, exePath); err != nil {
		t.Error("buildForServe(...):", err)
	}

	// Run program and check output to ensure correctness.
	got, err := exec.Command(exePath).Output()
	if err != nil {
		t.Fatal(err)
	}
	const want = "Hello, World!\n"
	if string(got) != want {
		t.Errorf("program output = %q; want %q", got, want)
	}
}
