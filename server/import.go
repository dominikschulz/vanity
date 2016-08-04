package server

// Import a single go-get import configuration item
type Import struct {
	Prefix string `yaml:"prefix"`
	VCS    string `yaml:"vcs"`
	URL    string `yaml:"url"`
	Docs   string `yaml:"docs"`
	Source string `yaml:"source"`
}
