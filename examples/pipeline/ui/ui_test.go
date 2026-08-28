package ui_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Roarge/sysml-federation/examples/pipeline/ui"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// sortableSHA256 is the digest of Sortable.min.js at tag 1.15.7 of
// SortableJS/Sortable, the same value NOTICE records.
const sortableSHA256 = "bf4241bc73fef7f11c59a283a69fe8051cdd31c6d8ff5a2b9ba219e7831fcf76"

// vendored is the one JavaScript file this repository did not write. The
// checksum above is what holds it, so nothing here parses it.
const vendored = "shared/Sortable.min.js"

func read(t *testing.T, name string) []byte {
	t.Helper()
	data, err := fs.ReadFile(ui.Files, name)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return data
}

func TestSR08_VendoredSortableChecksum(t *testing.T) {
	sum := sha256.Sum256(read(t, vendored))
	assert.Equal(t, hex.EncodeToString(sum[:]), sortableSHA256)
	licence := read(t, "shared/LICENSE.SortableJS")
	assert.True(t, bytes.HasPrefix(licence, []byte("MIT License")),
		"LICENSE.SortableJS to open with the MIT heading")
	assert.True(t, bytes.Contains(licence, []byte("Copyright (c) 2019 All contributors to Sortable")),
		"LICENSE.SortableJS to carry the vendor's copyright line")
}

// absoluteURL matches a scheme-qualified URL anywhere, and a protocol-relative
// URL where one can begin: after a quote, a backtick, an opening parenthesis
// or an equals sign. A JavaScript line comment follows none of those. There
// is no allowlist: the apps reference nothing outside their own origin.
var absoluteURL = regexp.MustCompile("https?://|[\"'`(=]//")

func TestSR10_NoResourceFromAnotherOrigin(t *testing.T) {
	var scanned []string
	err := fs.WalkDir(ui.Files, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data := read(t, path)
		scanned = append(scanned, path)
		for _, m := range absoluteURL.FindAllIndex(data, -1) {
			t.Errorf("%s: reference to another origin at byte %d: %q",
				path, m[0], data[m[0]:min(m[1]+40, len(data))])
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Contains(t, scanned, "shared/graphql.js")
	assert.Contains(t, scanned, vendored)
	assert.Contains(t, scanned, "shared/LICENSE.SortableJS")
	assert.Contains(t, scanned, "viewer/index.html")
	assert.Contains(t, scanned, "viewer/app.js")
}

// SC-06: report the size of each app, the shared module counted against the
// document app. Sortable.min.js is vendored, not hand-written, and is not
// counted. The estimate beside each row is what the plan expected, and a
// count above it is reported rather than failed: size never breaks a build.
func TestSC06_LineBudgets(t *testing.T) {
	cases := []struct {
		app      string
		files    []string
		estimate int
	}{
		{"viewer JavaScript", []string{"viewer/app.js", "viewer/tokeniser.js", "viewer/sketch.js"}, 900},
		{"viewer CSS", []string{"viewer/style.css"}, 300},
	}
	for _, c := range cases {
		t.Run(c.app, func(t *testing.T) {
			total := 0
			for _, name := range c.files {
				total += bytes.Count(read(t, name), []byte("\n"))
			}
			t.Logf("%s: %d lines, the plan estimated %d", c.app, total, c.estimate)
		})
	}
}

// TestEveryModuleParses is the only thing in the toolchain that reads the
// JavaScript at all. Nothing compiles it and no test executes it, so
// without this a syntax error would travel as far as the first browser to
// load a page. "node --check" parses a file and stops, so this is a syntax
// gate rather than a browser: it says nothing about whether the code works.
//
// Each file is written out under a .mjs name, which fixes the parse goal at
// ES module whatever the node version thinks of a bare .js file. Every file
// here is loaded by the apps as a module (AD-0017). node is not part of
// this project's toolchain, so a machine without it skips rather than
// fails, and the gate stays honest about what it did not do.
func TestEveryModuleParses(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not on the path, so the JavaScript was not parsed here")
	}
	dir := t.TempDir()
	var parsed []string
	walked := fs.WalkDir(ui.Files, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".js") || path == vendored {
			return err
		}
		out := filepath.Join(dir, strings.ReplaceAll(strings.TrimSuffix(path, ".js"), "/", "_")+".mjs")
		if err := os.WriteFile(out, read(t, path), 0o600); err != nil {
			return err
		}
		if report, err := exec.Command(node, "--check", out).CombinedOutput(); err != nil {
			t.Errorf("%s does not parse: %v\n%s", path, err, report)
		}
		parsed = append(parsed, path)
		return nil
	})
	assert.NoError(t, walked)
	// Without this the test would pass on a walk that found nothing.
	assert.Contains(t, parsed, "shared/graphql.js")
}
