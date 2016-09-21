// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go-import-redirector-gae is an HTTP server for a custom Go import domain.
// Unlike the original (located here: rsc.io/go-import-redirector) this version
// runs on Google App Engine. Please set the values in the first var block to
// the desired values.
//
// It responds to requests in a given import path root with a meta tag
// specifying the source repository for the ``go get'' command and an
// HTML redirect to the godoc.org documentation page for that package.
package go_import_redirector_gae

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
)

var (
	vcs        = "git"
	importPath = "m8k.in/*"
	repoPath   = "https://github.com/mikewiacek/*"
	wildcard   bool
)

func init() {
	if !strings.Contains(repoPath, "://") {
		panic("repo path must be full URL")
	}
	if strings.HasSuffix(importPath, "/*") != strings.HasSuffix(repoPath, "/*") {
		panic("either both import and repo must have /* or neither")
	}
	if strings.HasSuffix(importPath, "/*") {
		wildcard = true
		importPath = strings.TrimSuffix(importPath, "/*")
		repoPath = strings.TrimSuffix(repoPath, "/*")
	}
	http.HandleFunc(strings.TrimSuffix(importPath, "/")+"/", redirect)
}

var tmpl = template.Must(template.New("main").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.ImportRoot}} {{.VCS}} {{.VCSRoot}}">
<meta http-equiv="refresh" content="0; url=https://godoc.org/{{.ImportRoot}}{{.Suffix}}">
</head>
<body>
Nothing to see here; <a href="https://godoc.org/{{.ImportRoot}}{{.Suffix}}">move along</a>.
</body>
</html>
`))

type data struct {
	ImportRoot string
	VCS        string
	VCSRoot    string
	Suffix     string
}

func redirect(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(req.Host+req.URL.Path, "/")
	var importRoot, repoRoot, suffix string
	if wildcard {
		if path == importPath {
			http.Redirect(w, req, "https://godoc.org/"+importPath, 302)
			return
		}
		if !strings.HasPrefix(path, importPath+"/") {
			http.NotFound(w, req)
			return
		}
		elem := path[len(importPath)+1:]
		if i := strings.Index(elem, "/"); i >= 0 {
			elem, suffix = elem[:i], elem[i:]
		}
		importRoot = importPath + "/" + elem
		repoRoot = repoPath + "/" + elem
	} else {
		if path != importPath && !strings.HasPrefix(path, importPath+"/") {
			http.NotFound(w, req)
			return
		}
		importRoot = importPath
		repoRoot = repoPath
		suffix = path[len(importPath):]
	}
	d := &data{
		ImportRoot: importRoot,
		VCS:        vcs,
		VCSRoot:    repoRoot,
		Suffix:     suffix,
	}
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, d)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(buf.Bytes())
}
