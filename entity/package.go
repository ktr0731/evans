package entity

type Package struct {
	Name     string
	Services []Service
	Messages []Message
}

func NewPackage(name string, msgs []Message, svcs []Service) *Package {
	return &Package{
		Name:     name,
		Services: svcs,
		Messages: msgs,
	}
}
