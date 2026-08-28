// Package ui carries the two web apps and their shared module as embedded
// files. The UI server in cmd/sysml-federation serves them from Files: the
// viewer at /viewer/, the document at /document/ and the shared module at
// /shared/ (architecture V4, AD-0017).
package ui

import "embed"

// Files holds each app directory under its own name, so that
// fs.Sub(Files, "viewer") is the viewer's document root.
//
//go:embed shared viewer document
var Files embed.FS
