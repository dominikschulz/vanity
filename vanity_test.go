package main

import (
	"sync"
	"testing"
)

func TestDefaults(t *testing.T) {
	h := &Host{
		Imports: []Import{
			Import{
				Prefix: "example.org/path/to/a/package",
				VCS:    "hg",
				URL:    "https://code.google.com/p/package",
			},
		},
		Defaults: []Import{
			Import{
				Prefix: "example.org/x/user1/{{package}}",
				VCS:    "git",
				URL:    "https://github.com/user1/{{package}}",
			},
			Import{
				Prefix: "example.org/x/user2/{{package}}",
				VCS:    "git",
				URL:    "https://github.com/user2/{{package}}",
			},
			Import{
				Prefix: "example.org/x/{{package}}",
				VCS:    "git",
				URL:    "https://bitbucket.org/org1/{{package}}.git",
			},
			Import{
				Prefix: "ex.io/{{package}}",
				VCS:    "git",
				URL:    "https://bitbucket.org/org1/ex-{{package}}.git",
			},
		},
		mutex:     &sync.Mutex{},
		generated: make([]Import, 0),
	}

	var tests = []struct {
		Repo string
		URL  string
	}{
		{Repo: "example.org/path/to/a/package", URL: "https://code.google.com/p/package"},
		{Repo: "example.org/x/user1/package1", URL: "https://github.com/user1/package1"},
		{Repo: "example.org/x/user1/package2", URL: "https://github.com/user1/package2"},
		{Repo: "example.org/x/user2/package3", URL: "https://github.com/user2/package3"},
		{Repo: "example.org/x/foo", URL: "https://bitbucket.org/org1/foo.git"},
		{Repo: "ex.io/bar", URL: "https://bitbucket.org/org1/ex-bar.git"},
	}

	for _, test := range tests {
		is := h.getImports(test.Repo)
		found := false
		for _, i := range is {
			if i.Prefix == test.Repo && i.URL == test.URL {
				t.Logf("Matching import found: %s - %s", i.Prefix, i.URL)
				found = true
			}
		}
		if !found {
			t.Errorf("Found no valid import for %s -> %s", test.Repo, test.URL)
		}
	}
}

func TestGenImportPath(t *testing.T) {
	h := &Host{
		Imports: []Import{},
		Defaults: []Import{
			Import{
				Prefix: "example.org/x/{{package}}",
				VCS:    "git",
				URL:    "git@github.com:example/{{package}}.git",
				Docs:   "godoc.org/github.com/example/{{package}}",
				Source: "example.org/docs/{{package}}",
			},
		},
		mutex:     &sync.Mutex{},
		generated: make([]Import, 0),
	}

	prefix := "example.org/x/foo"
	url := "git@github.com:example/foo.git"
	i, err := h.genImport(prefix)
	if err != nil {
		t.Fatalf("Failed to generate import for %s: %s", prefix, err)
	}
	if i.URL != url {
		t.Errorf("URL should be %s not %s", url, i.URL)
	}

}

func TestGenImportNoPath(t *testing.T) {
	h := &Host{
		Imports: []Import{},
		Defaults: []Import{
			Import{
				Prefix: "example.org/{{package}}",
				VCS:    "git",
				URL:    "git@bitbucket.org:example/{{package}}.git",
				Docs:   "godoc.org/bitbucket.org/example/{{package}}",
				Source: "example.org/docs/{{package}}",
			},
		},
		mutex:     &sync.Mutex{},
		generated: make([]Import, 0),
	}

	prefix := "example.org/foo"
	url := "git@bitbucket.org:example/foo.git"
	i, err := h.genImport(prefix)
	if err != nil {
		t.Fatalf("Failed to generate import for %s: %s", prefix, err)
	}
	if i.URL != url {
		t.Errorf("URL should be %s not %s", url, i.URL)
	}

}
