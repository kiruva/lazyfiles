package remote

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// A native ssh client does not read ~/.ssh/config, so the keys that decide where
// a connection actually goes are parsed here. This is deliberately a subset —
// HostName, User, Port, IdentityFile, ProxyJump — which is what an alias needs to
// resolve. Anything else in the file is ignored, so an alias that depends on, say,
// a custom ProxyCommand will not resolve the way `ssh` would.

type sshHostConfig struct {
	HostName      string
	User          string
	Port          string
	IdentityFiles []string
	ProxyJump     string
}

// lookupSSHConfig resolves alias through ~/.ssh/config. First match wins per key,
// which is how ssh itself reads the file. Identity files always fall back to the
// standard key names so an entry without IdentityFile still authenticates.
func lookupSSHConfig(alias string) sshHostConfig {
	cfg := parseSSHConfig(sshConfigPath(), alias, 0)
	cfg.IdentityFiles = append(cfg.IdentityFiles, defaultIdentities()...)
	return cfg
}

func sshConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "config")
}

// parseSSHConfig walks the file, collecting values from every Host block whose
// pattern matches alias. depth bounds Include recursion.
func parseSSHConfig(path, alias string, depth int) sshHostConfig {
	var cfg sshHostConfig
	if path == "" || depth > 8 {
		return cfg
	}
	f, err := os.Open(path)
	if err != nil {
		return cfg
	}
	defer f.Close()

	applies := false // are we inside a block that matches alias?
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		key, value, ok := configLine(sc.Text())
		if !ok {
			continue
		}

		switch key {
		case "host":
			applies = matchesAny(value, alias)
			continue
		case "match":
			// Match blocks use a different language; treat them as not applying
			// rather than guessing.
			applies = false
			continue
		case "include":
			for _, inc := range expandInclude(value) {
				merge(&cfg, parseSSHConfig(inc, alias, depth+1))
			}
			continue
		}
		if !applies {
			continue
		}

		switch key {
		case "hostname":
			setOnce(&cfg.HostName, value)
		case "user":
			setOnce(&cfg.User, value)
		case "port":
			setOnce(&cfg.Port, value)
		case "identityfile":
			cfg.IdentityFiles = append(cfg.IdentityFiles, expandHome(unquote(value)))
		case "proxyjump":
			setOnce(&cfg.ProxyJump, value)
		}
	}
	return cfg
}

// configLine splits "Key value" or "Key = value", lowercasing the key.
func configLine(raw string) (key, value string, ok bool) {
	line := strings.TrimSpace(raw)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	if i := strings.IndexAny(line, " \t="); i >= 0 {
		key = strings.ToLower(line[:i])
		value = strings.TrimSpace(strings.TrimLeft(line[i:], " \t="))
	} else {
		return "", "", false
	}
	if key == "" || value == "" {
		return "", "", false
	}
	return key, value, true
}

// matchesAny reports whether alias matches any pattern in a Host line. A leading
// "!" negates, and a negated match rules the whole line out, as in ssh.
func matchesAny(patterns, alias string) bool {
	matched := false
	for _, p := range strings.Fields(patterns) {
		if negated, ok := strings.CutPrefix(p, "!"); ok {
			if matchPattern(negated, alias) {
				return false
			}
			continue
		}
		if matchPattern(p, alias) {
			matched = true
		}
	}
	return matched
}

// matchPattern implements ssh's glob: "*" spans any run, "?" one character.
func matchPattern(pattern, s string) bool {
	if pattern == "*" {
		return true
	}
	return globMatch(pattern, s)
}

func globMatch(pattern, s string) bool {
	// Iterative backtracking match — no allocation, no regexp.
	var pi, si, star, mark int
	star = -1
	for si < len(s) {
		switch {
		case pi < len(pattern) && (pattern[pi] == '?' || pattern[pi] == s[si]):
			pi++
			si++
		case pi < len(pattern) && pattern[pi] == '*':
			star = pi
			mark = si
			pi++
		case star >= 0:
			pi = star + 1
			mark++
			si = mark
		default:
			return false
		}
	}
	for pi < len(pattern) && pattern[pi] == '*' {
		pi++
	}
	return pi == len(pattern)
}

// expandInclude resolves an Include directive's globs, relative to ~/.ssh.
func expandInclude(value string) []string {
	var out []string
	for _, raw := range strings.Fields(value) {
		p := expandHome(unquote(raw))
		if !filepath.IsAbs(p) {
			home, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			p = filepath.Join(home, ".ssh", p)
		}
		matches, err := filepath.Glob(p)
		if err != nil {
			continue
		}
		out = append(out, matches...)
	}
	return out
}

func setOnce(dst *string, value string) {
	if *dst == "" {
		*dst = unquote(value)
	}
}

func merge(dst *sshHostConfig, src sshHostConfig) {
	setOnce(&dst.HostName, src.HostName)
	setOnce(&dst.User, src.User)
	setOnce(&dst.Port, src.Port)
	setOnce(&dst.ProxyJump, src.ProxyJump)
	dst.IdentityFiles = append(dst.IdentityFiles, src.IdentityFiles...)
}

func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
