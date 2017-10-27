package repl

import (
	"github.com/ktr0731/evans/env"
	"github.com/pkg/errors"
)

type PackageCommand struct {
	env *env.Env
}

func (c *PackageCommand) Synopsis() string {
	return "Set the package as the current selected package"
}

func (c *PackageCommand) Help() string {
	return "Usage: package <package name>"
}

func (c *PackageCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "package name")
	}
	return nil
}

func (c *PackageCommand) Run(args []string) (string, error) {
	if err := c.env.UsePackage(args[0]); err != nil {
		return "", errors.Wrapf(err, "file %s", args[0])
	}
	return "", nil
}
