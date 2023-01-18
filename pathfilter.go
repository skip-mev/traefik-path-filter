package traefik_path_filter

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
)

type Config struct {
	Allowlist []string `json:"allowlist,omitempty"`
	Blocklist []string `json:"blocklist,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		Allowlist: make([]string, 0),
		Blocklist: make([]string, 0),
	}
}

type PathFilter struct {
	next      http.Handler
	allowlist []*regexp.Regexp
	blocklist []*regexp.Regexp
	name      string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Allowlist) == 0 && len(config.Blocklist) == 0 {
		return nil, fmt.Errorf("both allowlist and blocklist cannot be empty")
	}

	if len(config.Allowlist) != 0 && len(config.Blocklist) != 0 {
		return nil, fmt.Errorf("both allowlist and blocklist cannot be populated")
	}

	allowlist := make([]*regexp.Regexp, len(config.Allowlist))
	blocklist := make([]*regexp.Regexp, len(config.Blocklist))

	for i, regex := range config.Allowlist {
		re, err := regexp.Compile(regex)

		if err != nil {
			return nil, fmt.Errorf("cannot compile regex in allowlist %q: %w", regex, err)
		}

		allowlist[i] = re
	}

	for i, regex := range config.Blocklist {
		re, err := regexp.Compile(regex)

		if err != nil {
			return nil, fmt.Errorf("cannot compile regex in blocklist %q: %w", regex, err)
		}

		blocklist[i] = re
	}

	return &PathFilter{
		allowlist: allowlist,
		blocklist: blocklist,
		next:      next,
		name:      name,
	}, nil
}

func (pf *PathFilter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	currentPath := req.URL.EscapedPath()

	if len(pf.blocklist) > 0 {
		for _, re := range pf.blocklist {
			if re.MatchString(currentPath) {
				http.Error(rw, "This path is blocked", http.StatusForbidden)
				return
			}
		}
	}

	if len(pf.allowlist) > 0 {
		for _, re := range pf.allowlist {
			if !re.MatchString(currentPath) {
				http.Error(rw, "This path is blocked", http.StatusForbidden)
				return
			}
		}
	}

	pf.next.ServeHTTP(rw, req)
}
