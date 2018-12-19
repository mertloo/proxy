package proxy

import (
	"fmt"
	"strings"
)

type Config struct {
	Proto    string
	Addr     string
	Method   string
	Password string
}

func ParseConfig(expr string) (*Config, error) {
	if expr == "tcp" {
		return &Config{Proto: "tcp"}, nil
	}
	parts := strings.SplitN(expr, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("bad expr %s", expr)
	}
	conf := &Config{Proto: parts[0]}
	if conf.Proto == "socks5" {
		conf.Addr = parts[1]
		return conf, nil
	}
	if conf.Proto == "ssocks" {
		parts = strings.SplitN(parts[1], "@", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad expr %s", expr)
		}
		conf.Addr = parts[1]
		parts = strings.SplitN(parts[0], ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad expr %s", expr)
		}
		conf.Method, conf.Password = parts[0], parts[1]
		return conf, nil
	}
	return nil, fmt.Errorf("bad expr %s", expr)
}
