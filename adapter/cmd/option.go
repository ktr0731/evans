package cmd

import (
	"flag"
	"fmt"
)

type optStrSlice []string

func (o *optStrSlice) String() string {
	return fmt.Sprintf("%v", *o)
}

func (o *optStrSlice) Set(v string) error {
	*o = append(*o, v)
	return nil
}

var usageFormat = `
Usage: %s [options ...] [PROTO [PROTO ...]]

Positional arguments:
	PROTO			.proto files

Options:
	--edit, -e		%s
	--repl			%s
	--cli			%s
	--silent, -s		%s
	--host HOST		%s
	--port PORT, -p PORT	%s
	--package PACKAGE	%s
	--service SERVICE	%s
	--call CALL		%s
	--file FILE, -f FILE	%s
	--path PATH		%s
	--header HEADER		%s
	--web			%s
	--reflection, -r	%s

	--help, -h		%s
	--version, -v		%s
`

func (c *Command) parseFlags(args []string) *options {
	const (
		edit       = "edit config file using by $EDITOR"
		repl       = "start as REPL mode"
		cli        = "start as CLI mode"
		silent     = "hide splash"
		host       = "gRPC server host"
		port       = "gRPC server port"
		pkg        = "default package"
		service    = "default service"
		call       = "call specified RPC by CLI mode"
		file       = "the script file which will be executed by (used only CLI mode)"
		path       = "proto file paths"
		header     = "default headers which set to each requests (example: foo=bar)"
		web        = "use gRPC Web protocol"
		reflection = "use gRPC reflection"

		version = "display version and exit"
		help    = "display this help and exit"
	)

	f := flag.NewFlagSet("main", flag.ExitOnError)
	f.Usage = func() {
		c.Version()
		fmt.Fprintf(
			c.ui.Writer(),
			usageFormat,
			c.name,
			edit,
			repl,
			cli,
			silent,
			host,
			port,
			pkg,
			service,
			call,
			file,
			path,
			header,
			web,
			reflection,
			help,
			version,
		)
	}

	var opts options

	f.BoolVar(&opts.editConfig, "edit", false, edit)
	f.BoolVar(&opts.editConfig, "e", false, edit)
	f.BoolVar(&opts.repl, "repl", false, repl)
	f.BoolVar(&opts.cli, "cli", false, cli)
	f.BoolVar(&opts.silent, "silent", false, silent)
	f.BoolVar(&opts.silent, "s", false, silent)
	f.StringVar(&opts.host, "host", "", host)
	f.StringVar(&opts.port, "port", "50051", port)
	f.StringVar(&opts.port, "p", "50051", port)
	f.StringVar(&opts.pkg, "package", "", pkg)
	f.StringVar(&opts.service, "service", "", service)
	f.StringVar(&opts.call, "call", "", call)
	f.StringVar(&opts.file, "file", "", file)
	f.StringVar(&opts.file, "f", "", file)
	f.Var(&opts.path, "path", path)
	f.Var(&opts.header, "header", header)
	f.BoolVar(&opts.web, "web", false, web)
	f.BoolVar(&opts.reflection, "reflection", false, reflection)
	f.BoolVar(&opts.reflection, "r", false, reflection)
	f.BoolVar(&opts.version, "version", false, version)
	f.BoolVar(&opts.version, "v", false, version)

	// ignore error because flag set mode is ExitOnError
	_ = f.Parse(args)

	c.flagSet = f

	return &opts
}

type options struct {
	// mode options
	editConfig bool

	// config options
	repl       bool
	cli        bool
	silent     bool
	host       string
	port       string
	pkg        string
	service    string
	call       string
	file       string
	path       optStrSlice
	header     optStrSlice
	web        bool
	reflection bool

	// meta options
	version bool
}
