package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// defaultConfigPath returns ~/.ssh/knockr.conf. The .ssh directory is chosen
// deliberately: it is the one location users and dotfiles tooling reliably
// treat as secret, reducing the chance a knock sequence is accidentally
// published in a dotfiles repository.
func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".ssh", "knockr.conf"), nil
}

// checkPerms refuses profile files readable by group or others, matching
// ssh's treatment of private keys. Skipped on Windows, where POSIX permission
// bits are not meaningful.
func checkPerms(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("%s: not a regular file", path)
	}
	if runtime.GOOS == "windows" {
		return nil
	}
	if perm := fi.Mode().Perm(); perm&0o077 != 0 {
		return fmt.Errorf("%s: permissions %04o are too open; run: chmod 600 %s", path, perm, path)
	}

	return nil
}

// loadProfiles parses an ssh_config-style profile file:
//
//	profile router
//	    host 203.0.113.1
//	    ports 1234,8923,1233
//	    protocol udp
//	    delay 150ms
//	    timeout 2s
//	    verify 22
//
// Keywords are case-insensitive; blank lines and # comments are ignored.
func loadProfiles(path string) (map[string]map[string]string, error) {
	if err := checkPerms(path); err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	profiles := map[string]map[string]string{}
	var current map[string]string
	lineNo := 0

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("%s:%d: expected 'keyword value', got %q", path, lineNo, line)
		}
		key, value := strings.ToLower(fields[0]), fields[1]

		if key == "profile" {
			if _, dup := profiles[value]; dup {
				return nil, fmt.Errorf("%s:%d: duplicate profile %q", path, lineNo, value)
			}
			current = map[string]string{}
			profiles[value] = current

			continue
		}

		if current == nil {
			return nil, fmt.Errorf("%s:%d: %q appears before any profile", path, lineNo, key)
		}

		switch key {
		case "host", "ports", "protocol", "delay", "timeout", "silent", "verify":
			if _, dup := current[key]; dup {
				return nil, fmt.Errorf("%s:%d: duplicate key %q", path, lineNo, key)
			}
			current[key] = value
		default:
			return nil, fmt.Errorf("%s:%d: unknown key %q", path, lineNo, key)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

// fromProfile loads the named profile and applies it to cfg, returning the
// host and ports to knock.
func fromProfile(cfg *config, name string, explicit map[string]bool) (string, []int, error) {
	path := cfg.config
	if path == "" {
		var err error
		path, err = defaultConfigPath()
		if err != nil {
			return "", nil, err
		}
	}

	profiles, err := loadProfiles(path)
	if err != nil {
		return "", nil, err
	}

	p, ok := profiles[name]
	if !ok {
		names := make([]string, 0, len(profiles))
		for n := range profiles {
			names = append(names, n)
		}
		sort.Strings(names)

		return "", nil, fmt.Errorf("profile %q not found in %s (available: %s)",
			name, path, strings.Join(names, ", "))
	}

	return applyProfile(cfg, p, explicit)
}

// applyProfile fills cfg from profile values. Flags set explicitly on the
// command line take precedence over profile values.
func applyProfile(cfg *config, p map[string]string, explicit map[string]bool) (string, []int, error) {
	host, ok := p["host"]
	if !ok {
		return "", nil, errors.New("profile missing required key: host")
	}
	portsRaw, ok := p["ports"]
	if !ok {
		return "", nil, errors.New("profile missing required key: ports")
	}
	ports, err := parsePorts(portsRaw)
	if err != nil {
		return "", nil, err
	}

	if v, ok := p["protocol"]; ok && !explicit["n"] {
		cfg.protocol = v
	}
	if v, ok := p["delay"]; ok && !explicit["d"] {
		d, err := time.ParseDuration(v)
		if err != nil {
			return "", nil, fmt.Errorf("profile delay: %w", err)
		}
		cfg.delay = d
	}
	if v, ok := p["timeout"]; ok && !explicit["t"] {
		d, err := time.ParseDuration(v)
		if err != nil {
			return "", nil, fmt.Errorf("profile timeout: %w", err)
		}
		cfg.timeout = d
	}
	if v, ok := p["silent"]; ok && !explicit["s"] {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return "", nil, fmt.Errorf("profile silent: %w", err)
		}
		cfg.silent = b
	}
	if v, ok := p["verify"]; ok && !explicit["verify"] {
		cfg.verify = v
	}

	return host, ports, nil
}
