package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessDirectory(t *testing.T) {
	// create temp input dir
	inDir := t.TempDir()
	outDir := t.TempDir()

	// create default template
	tmpl := `<!DOCTYPE html>
<html>
<head><title>test</title></head>
<body>{{ body . }}</body>
</html>`
	if err := os.WriteFile(filepath.Join(inDir, "default.tmpl"), []byte(tmpl), 0644); err != nil {
		t.Fatal(err)
	}

	// create a .lit file with yaml frontmatter
	lit1 := `<!--yaml
title: Hello
-->

¶ ⦊Hello world⦉
`
	if err := os.WriteFile(filepath.Join(inDir, "index.lit"), []byte(lit1), 0644); err != nil {
		t.Fatal(err)
	}

	// create nested dir
	if err := os.MkdirAll(filepath.Join(inDir, "posts"), 0755); err != nil {
		t.Fatal(err)
	}

	// create nested .lit file without frontmatter
	lit2 := `¶ ⦊A post⦉
`
	if err := os.WriteFile(filepath.Join(inDir, "posts", "one.lit"), []byte(lit2), 0644); err != nil {
		t.Fatal(err)
	}

	// set flags and run
	*in = inDir
	*out = outDir
	*outmode = "html"
	processDirectory(inDir, outDir)

	// check index.html exists and has content
	indexHTML, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("reading index.html: %v", err)
	}
	if !strings.Contains(string(indexHTML), "Hello world") {
		t.Errorf("index.html missing content, got: %s", indexHTML)
	}
	if !strings.Contains(string(indexHTML), "<!DOCTYPE html>") {
		t.Errorf("index.html missing doctype from template, got: %s", indexHTML)
	}
	// should not contain yaml block
	if strings.Contains(string(indexHTML), "title: Hello") {
		t.Errorf("index.html should not contain yaml frontmatter, got: %s", indexHTML)
	}

	// check posts/one.html exists
	postHTML, err := os.ReadFile(filepath.Join(outDir, "posts", "one.html"))
	if err != nil {
		t.Fatalf("reading posts/one.html: %v", err)
	}
	if !strings.Contains(string(postHTML), "A post") {
		t.Errorf("posts/one.html missing content, got: %s", postHTML)
	}
}

func TestProcessDirectoryCustomTemplate(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// create custom template
	customTmpl := `<article>{{ body . }}</article>`
	if err := os.WriteFile(filepath.Join(inDir, "custom.tmpl"), []byte(customTmpl), 0644); err != nil {
		t.Fatal(err)
	}

	// create .lit file specifying custom template
	lit1 := `<!--yaml
template: custom.tmpl
-->

¶ ⦊Custom content⦉
`
	if err := os.WriteFile(filepath.Join(inDir, "page.lit"), []byte(lit1), 0644); err != nil {
		t.Fatal(err)
	}

	*in = inDir
	*out = outDir
	*outmode = "html"
	processDirectory(inDir, outDir)

	pageHTML, err := os.ReadFile(filepath.Join(outDir, "page.html"))
	if err != nil {
		t.Fatalf("reading page.html: %v", err)
	}
	if !strings.Contains(string(pageHTML), "<article>") {
		t.Errorf("page.html should use custom template, got: %s", pageHTML)
	}
	if !strings.Contains(string(pageHTML), "Custom content") {
		t.Errorf("page.html missing content, got: %s", pageHTML)
	}
}

func TestProcessDirectoryNoTemplate(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// no template files - should output raw html
	lit1 := `¶ ⦊Raw content⦉
`
	if err := os.WriteFile(filepath.Join(inDir, "raw.lit"), []byte(lit1), 0644); err != nil {
		t.Fatal(err)
	}

	*in = inDir
	*out = outDir
	*outmode = "html"
	processDirectory(inDir, outDir)

	rawHTML, err := os.ReadFile(filepath.Join(outDir, "raw.html"))
	if err != nil {
		t.Fatalf("reading raw.html: %v", err)
	}
	if !strings.Contains(string(rawHTML), "Raw content") {
		t.Errorf("raw.html missing content, got: %s", rawHTML)
	}
}
