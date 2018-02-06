package entity

type Parser interface {
	ParseFile(fnames []string, fpaths []string) ([]*Package, error)
}
