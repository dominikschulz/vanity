package server

import (
	"fmt"
	"strings"
)

// Host contains the config for a single vhost
type Host struct {
	Imports  []Import `yaml:"imports"`
	Defaults []Import `yaml:"defaults"`
}

// getImport will retrieve or generate the import for a given repo
func (h *Host) getImport(repo string) (Import, error) {
	for _, i := range h.Imports {
		if strings.HasPrefix(repo, i.Prefix) {
			return i, nil
		}
	}
	i, err := h.genImport(repo)
	if err != nil {
		return Import{}, err
	}
	return i, nil
}

// genImport will try to generate an imports definition given an
// repo prefix based on our set of defaults defintions.
func (h *Host) genImport(repo string) (Import, error) {
	var pkg string
	var matchingDefault Import

	// try to find the most specific defaults definition
	for _, defn := range h.Defaults {
		prefixNoPlaceholder := strings.Replace(defn.Prefix, "{{package}}", "", -1)
		if strings.HasPrefix(repo, prefixNoPlaceholder) {
			p := strings.Split(strings.TrimPrefix(repo, prefixNoPlaceholder), "/")
			if len(p) > 0 {
				pkg = p[0]
				matchingDefault = defn
				break
			}
		}
	}

	if pkg == "" {
		return Import{}, fmt.Errorf("No default found for %s", repo)
	}

	// replace the placeholders in the default
	prefix := strings.Replace(matchingDefault.Prefix, "{{package}}", pkg, -1)
	url := strings.Replace(matchingDefault.URL, "{{package}}", pkg, -1)
	docs := strings.Replace(matchingDefault.Docs, "{{package}}", pkg, -1)
	source := strings.Replace(matchingDefault.Source, "{{package}}", pkg, -1)

	genImport := Import{
		Prefix: prefix,
		VCS:    matchingDefault.VCS,
		URL:    url,
		Docs:   docs,
		Source: source,
	}

	// save the generated import for later use
	return genImport, nil
}
