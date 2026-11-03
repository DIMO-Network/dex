package web

import (
	"io/fs"
	"path"
	"regexp"
	"testing"
)

// remoteScript matches a script the login pages would fetch from another origin: a
// <script src> or an ES module import with an absolute or protocol-relative URL.
var remoteScript = regexp.MustCompile(`(?i)<script[^>]*\ssrc\s*=\s*["']?(?:[a-z][a-z0-9+.-]*:)?//|(?:\bfrom|\bimport\s*\(?)\s*["'](?:[a-z][a-z0-9+.-]*:)?//`)

// TestNoRemoteScripts keeps every script the pages run on the issuer's own origin, so a
// script-src of 'self' covers them and no CDN can change what runs on the login pages.
func TestNoRemoteScripts(t *testing.T) {
	err := fs.WalkDir(FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		switch path.Ext(p) {
		case ".html", ".js", ".mjs":
		default:
			return nil
		}
		b, err := fs.ReadFile(FS(), p)
		if err != nil {
			return err
		}
		if m := remoteScript.Find(b); m != nil {
			t.Errorf("%s loads a script from another origin: %q", p, m)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoteScriptPattern(t *testing.T) {
	for _, s := range []string{
		`import { Web3Modal } from "https://unpkg.com/@web3modal/html@2.6.2";`,
		`import {x} from'//cdn.example.com/x.js'`,
		`const m = await import("https://cdn.example.com/x.js");`,
		`<script type="module" src="https://unpkg.com/x.js"></script>`,
		`<SCRIPT src=//cdn.example.com/x.js></SCRIPT>`,
	} {
		if !remoteScript.MatchString(s) {
			t.Errorf("not matched: %s", s)
		}
	}
	for _, s := range []string{
		`<script type="module" src="../static/web3-login.js"></script>`,
		`import { a } from "./a.js";`,
		`<a href="https://example.com">x</a>`,
		`<img src="https://example.com/logo.svg">`,
	} {
		if remoteScript.MatchString(s) {
			t.Errorf("matched: %s", s)
		}
	}
}
