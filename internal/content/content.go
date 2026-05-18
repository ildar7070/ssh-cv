// Package content loads the personal bio shown to SSH visitors.
package content

import (
	"fmt"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

type Profile struct {
	Name     string    `toml:"name"`
	Tagline  string    `toml:"tagline"`
	About    About     `toml:"about"`
	CV       []CVEntry `toml:"cv"`
	Projects []Project `toml:"projects"`
	Contact  Contact   `toml:"contact"`
}

type About struct {
	Lines []string `toml:"lines"`
}

type CVEntry struct {
	Role    string   `toml:"role"`
	Company string   `toml:"company"`
	Period  string   `toml:"period"`
	Summary string   `toml:"summary"`
	Bullets []string `toml:"bullets"`
}

type Project struct {
	Name    string   `toml:"name"`
	Tagline string   `toml:"tagline"`
	URL     string   `toml:"url"`
	Details []string `toml:"details"`
}

type Contact struct {
	Email    string `toml:"email"`
	GitHub   string `toml:"github"`
	LinkedIn string `toml:"linkedin"`
	Instagram string `toml:"instagram"`
}

func Load(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var p Profile
	if err := toml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &p, nil
}
