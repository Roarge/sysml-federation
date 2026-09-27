package ui_test

import (
	"bytes"
	"io/fs"
	"os"
	"testing"

	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/examples/pipeline/ui"
	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^the repository$`, func() {})
	sc.When(`^its root is read$`, w.theRootIsRead)
	sc.Then(`^its NOTICE names the Cosmo router and SortableJS$`, w.theNoticeNamesBoth)
	sc.Then(`^the MIT text of SortableJS sits beside the vendored file, with the vendor's copyright line$`, w.theMITTextIsInPlace)
	sc.Given(`^the files the demo embeds for both apps$`, func() {})
	sc.When(`^they are scanned for absolute and protocol-relative URLs$`, w.theFilesAreScanned)
	sc.Then(`^none is found$`, w.noneIsFound)
	sc.Then(`^the shared module, the vendored library and the viewer's page were among the files scanned$`, w.theKeyFilesWereScanned)
}

// world is what one scenario of this package knows.
type world struct {
	t       *testing.T
	notice  []byte
	scanned []string
	found   []string
}

// theRootIsRead reads the NOTICE at the repository root, three directories up
// from this package.
func (w *world) theRootIsRead() {
	raw, err := os.ReadFile("../../../NOTICE")
	w.notice = assert.Must(w.t, raw, err)
}

func (w *world) theNoticeNamesBoth() {
	for _, work := range []string{"Cosmo router", "SortableJS"} {
		assert.True(w.t, bytes.Contains(w.notice, []byte(work)), "the NOTICE names "+work)
	}
}

func (w *world) theMITTextIsInPlace() {
	licence := read(w.t, "shared/LICENSE.SortableJS")
	assert.True(w.t, bytes.HasPrefix(licence, []byte("MIT License")), "the licence opens with the MIT heading")
	assert.True(w.t, bytes.Contains(licence, []byte("Copyright (c) 2019 All contributors to Sortable")),
		"the licence carries the vendor's copyright line")
	_, err := fs.Stat(ui.Files, vendored)
	assert.NoError(w.t, err)
}

func (w *world) theFilesAreScanned() {
	err := fs.WalkDir(ui.Files, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data := read(w.t, path)
		w.scanned = append(w.scanned, path)
		for _, m := range absoluteURL.FindAllIndex(data, -1) {
			w.found = append(w.found, path+": "+string(data[m[0]:min(m[1]+40, len(data))]))
		}
		return nil
	})
	assert.NoError(w.t, err)
}

func (w *world) noneIsFound() {
	for _, found := range w.found {
		w.t.Errorf("a reference to another origin: %s", found)
	}
}

func (w *world) theKeyFilesWereScanned() {
	for _, name := range []string{"shared/graphql.js", vendored, "viewer/index.html"} {
		assert.Contains(w.t, w.scanned, name)
	}
}
