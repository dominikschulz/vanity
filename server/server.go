package server

import (
	"fmt"
	"net"
	"net/http"
	"text/template"

	"github.com/go-kit/kit/log"
)

var metaTemplate = template.Must(template.New("meta").Parse(`
{{range .}}<meta name="go-import" content="{{.Prefix}} {{.VCS}} {{.URL}}">
{{end}}
`))

// Server contains all vhosts this server knows about
type Server struct {
	hosts map[string]*Host
	log   log.Logger
}

type Config struct {
	Log   log.Logger
	Hosts map[string]*Host
}

// New will create a new vanity server
func New(cfg Config) *Server {
	if cfg.Log == nil {
		cfg.Log = log.NewNopLogger()
	}

	s := &Server{
		hosts: make(map[string]*Host, len(cfg.Hosts)),
	}
	// initialize the Hosts, which have not been fully
	// initialized after loading from YAML
	for k, v := range cfg.Hosts {
		s.hosts[k] = &Host{
			Imports:  v.Imports,
			Defaults: v.Defaults,
		}
	}
	return s
}

// ServeHTTP will serve http
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
		s.log.Log("level", "error", "msg", "Error writing response", "err", err)
	}
}

// lookup will try to look up an imports definition for a given
// host and port.
func (s *Server) lookup(host string, path string) ([]Import, error) {
	if h, found := s.hosts[host]; found {
		i, err := h.getImport(host + path)
		if err != nil {
			return []Import{}, err
		}
		return []Import{i}, nil
	}
	return []Import{}, fmt.Errorf("Host %s not found", host)
}
