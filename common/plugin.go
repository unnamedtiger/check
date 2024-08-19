package common

type Plugin struct {
	Name       string
	Doc        string
	Extensions []string
	Run        func(analysis *Analysis) error
}

func (p *Plugin) handlesExtension(ext string) bool {
	for _, e := range p.Extensions {
		if e == ext {
			return true
		}
	}
	return false
}
