package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"text/template"
)

var metaTemplate = template.Must(template.New("meta").Parse(`
{{range .}}<meta name="go-import" content="{{.Prefix}} {{.VCS}} {{.URL}}">
{{end}}
`))

type Server struct {
	hosts map[string]*Host
}

// NewServer will create a new vanity server
func NewServer(h map[string]Host) *Server {
	s := &Server{
		hosts: make(map[string]*Host, len(h)),
	}
	// initialize the Hosts, which have not been fully
	// initialized after loading from YAML
	for k, v := range h {
		s.hosts[k] = &Host{
			Imports:   v.Imports,
			Defaults:  v.Defaults,
			mutex:     &sync.Mutex{},
			generated: make([]Import, 0),
		}
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "github.com/dominikschulz/vanity")

	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}

	if r.FormValue("go-get") != "1" {
		http.Redirect(w, r, "http://godoc.org/"+host+r.URL.Path, http.StatusFound)
		return
	}

	imports, err := s.lookup(host, r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := metaTemplate.Execute(w, imports); err != nil {
		log.Println("Error writing response:", err)
	}
}

// lookup will try to look up an imports definition for a given
// host and port.
func (s *Server) lookup(host string, path string) ([]Import, error) {
	if h, found := s.hosts[host]; found {
		return h.getImports(host + path), nil
	}
	return []Import{}, fmt.Errorf("Host %s not found", host)
}

// getImports will try to look up an imports
// definition from the static and cached generated
// imports or try to generate one if not already present.
func (h *Host) getImports(repo string) []Import {
	h.mutex.Lock()
	imports := make([]Import, 0, len(h.Imports)+len(h.generated)+1)
	for _, i := range h.Imports {
		imports = append(imports, i)
	}
	for _, i := range h.generated {
		imports = append(imports, i)
	}
	h.mutex.Unlock()

	for _, i := range imports {
		if strings.HasPrefix(repo, i.Prefix) {
			return imports
		}
	}

	gen, err := h.genImport(repo)
	if err == nil {
		imports = append(imports, gen)
	} else {
		log.Printf("No default generated for %s: %s", repo, err)
	}
	return imports
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
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.generated = append(h.generated, genImport)
	return genImport, nil
}
