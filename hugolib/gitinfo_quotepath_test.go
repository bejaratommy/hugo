// Copyright 2026 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugolib

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bep/gitmap"
	qt "github.com/frankban/quicktest"
)

// A content path with a byte above 0x80 must still receive GitInfo when Git's
// core.quotepath is left at its default. See https://github.com/gohugoio/hugo/issues/9810.
func TestGitMapKeepsNonASCIIPathsWhenQuotePathIsDisabled(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	c := qt.New(t)
	dir := t.TempDir()
	contentDir := filepath.Join(dir, "content")
	c.Assert(os.MkdirAll(contentDir, 0o755), qt.IsNil)

	const name = "hugö.md"
	c.Assert(os.WriteFile(filepath.Join(contentDir, name), []byte("hi\n"), 0o644), qt.IsNil)

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Hugo Test",
			"GIT_AUTHOR_EMAIL=hugo@example.com",
			"GIT_COMMITTER_NAME=Hugo Test",
			"GIT_COMMITTER_EMAIL=hugo@example.com",
		)
		out, err := cmd.CombinedOutput()
		c.Assert(err, qt.IsNil, qt.Commentf("%s", out))
	}

	git("init")
	git("config", "core.quotepath", "true")
	git("add", filepath.Join("content", name))
	git("commit", "-m", "add non-ascii path")

	repo, err := gitmap.Map(gitmap.Options{
		Repository: dir,
		GetGitCommandFunc: func(stdout, stderr io.Writer, args ...string) (gitmap.Runner, error) {
			cmd := exec.Command("git", gitArgsDisableQuotePath(args)...)
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			return cmd, nil
		},
	})
	c.Assert(err, qt.IsNil)

	info, ok := repo.Files["content/"+name]
	c.Assert(ok, qt.IsTrue)
	c.Assert(info.Subject, qt.Equals, "add non-ascii path")
	c.Assert(info.Hash, qt.Not(qt.Equals), "")
}
