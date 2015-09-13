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

func NewServer(h map[string]Host) *Server {
	s := &Server{
		hosts: make(map[string]*Host, len(h)),
	}
	for k, v := range h {
		s.hosts[k] = &Host{
			Imports:   v.Imports,
			Default:   v.Default,
			mutex:     &sync.Mutex{},
			generated: make([]Import, 0),
		}
	}
	return s
}

func (s *Server) lookup(host string, path string) ([]Import, error) {
	if h, found := s.hosts[host]; found {
		return h.getImports(host + path), nil
	}
	return []Import{}, fmt.Errorf("Host %s not found", host)
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

func (h *Host) getImports(repo string) []Import {
	log.Printf("getting imports for %s - %s", repo, h.Default.Prefix)

	h.mutex.Lock()
	imports := make([]Import, 0, len(h.Imports)+len(h.generated)+1)
	for _, i := range h.Imports {
		log.Printf("copying import: %s", i.Prefix)
		imports = append(imports, i)
	}
	for _, i := range h.generated {
		imports = append(imports, i)
	}
	h.mutex.Unlock()

	for _, i := range imports {
		log.Printf("getImports() - Matching %s against %s", i.Prefix, repo)
		if strings.HasPrefix(repo, i.Prefix) {
			log.Println("MATCHED")
			return imports
		}
	}

	gen, err := h.genImport(repo)
	if err == nil {
		log.Printf("getImports() - generated import: %s - %s", gen.Prefix, gen.URL)
		imports = append(imports, gen)
	}
	return imports
}

func (h *Host) genImport(repo string) (Import, error) {
	var cleanRepo string
	var matchingDefault Import
	for _, defn := range h.Defaults {
		log.Printf("genImport(%s) - Trying default %s", repo, defn.Prefix)
		prefixNoPlaceholder := strings.Replace(defn.Prefix, "{{package}}", "", -1)
		if strings.HasPrefix(repo, prefixNoPlaceholder) {
			cleanRepo = strings.TrimPrefix(repo, prefixNoPlaceholder)
			matchingDefault = defn
			log.Printf("genImport(%s) - Match: %s", repo, cleanRepo)
			break
		}
	}
	p := strings.Split(cleanRepo, "/")
	if len(p) < 1 || p[0] == "" {
		return Import{}, fmt.Errorf("Invalid path")
	}
	pkg := p[0]
	prefix := strings.Replace(matchingDefault.Prefix, "{{package}}", pkg, -1)
	url := strings.Replace(matchingDefault.URL, "{{package}}", pkg, -1)
	docs := strings.Replace(matchingDefault.Docs, "{{package}}", pkg, -1)
	source := strings.Replace(matchingDefault.Source, "{{package}}", pkg, -1)

	def := Import{
		Prefix: prefix,
		VCS:    matchingDefault.VCS,
		URL:    url,
		Docs:   docs,
		Source: source,
	}

	log.Printf("auto-generated import: %s - %s - %s - %s - %s", def.Prefix, def.VCS, def.URL, def.Docs, def.Source)
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.generated = append(h.generated, def)
	return def, nil
}
