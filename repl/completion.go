package repl

import (
	"fmt"
	"strings"
)

// TODO: まともにする
func (r *REPL) GetCompletion(line string) []string {
	// completion for target
	sp := strings.Split(line, " ")
	switch sp[0] {
	case "show":
		if len(sp) > 2 {
			return nil
		}
		c := []string{"package", "service", "message", "rpc"}
		for i, e := range c {
			c[i] = fmt.Sprintf("%s %s ", sp[0], e)
		}
		return c

	case "call":
		if len(sp) > 2 {
			return nil
		}
		rpc, err := r.env.GetRPCs()
		if err != nil {
			return nil
		}
		r := make([]string, len(rpc))
		for i, rpc := range rpc {
			r[i] = fmt.Sprintf("%s %s ", sp[0], rpc.Name)
		}
		return r

	case "desc":
		if len(sp) > 2 {
			return nil
		}
		msg, err := r.env.GetMessages()
		if err != nil {
			return nil
		}
		m := make([]string, len(msg))
		for i, msg := range msg {
			m[i] = msg.Name
		}
		return m

	case "package":
		if len(sp) > 2 {
			return nil
		}
		pkg := r.env.GetPackages()
		p := make([]string, len(pkg))
		for i, pkg := range pkg {
			p[i] = pkg.Name
		}
		return p

	case "service":
		if len(sp) > 2 {
			return nil
		}
		svc, err := r.env.GetServices()
		if err != nil {
			return nil
		}
		s := make([]string, len(svc))
		for i, svc := range svc {
			s[i] = svc.Name
		}
		return s
	}

	// Completion for command name
	switch {
	case strings.HasPrefix(line, "sh"):
		return []string{"show "}

	case strings.HasPrefix(line, "c"):
		return []string{"call "}

	case strings.HasPrefix(line, "d"):
		return []string{"desc "}

	case strings.HasPrefix(line, "p"):
		return []string{"package "}

	case strings.HasPrefix(line, "s"):
		return []string{"service "}
	}

	return nil
}
