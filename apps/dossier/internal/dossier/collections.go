package dossier

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ListCollections returns a copy of configured collections (never nil).
func ListCollections() []Collection {
	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		return []Collection{}
	}
	if cfg.Collections == nil {
		return []Collection{}
	}
	out := make([]Collection, len(cfg.Collections))
	copy(out, cfg.Collections)
	return out
}

// CreateCollection creates an empty named collection.
func CreateCollection(name string) (*Collection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("collection name required")
	}
	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		cfg = &AppConfig{}
	}
	col := Collection{
		ID:    uuid.NewString(),
		Name:  name,
		Paths: []string{},
	}
	cfg.Collections = append(cfg.Collections, col)
	if err := SaveConfig(cfg); err != nil {
		return nil, err
	}
	return &col, nil
}

// DeleteCollection removes a collection by id.
func DeleteCollection(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("id required")
	}
	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		return fmt.Errorf("no config")
	}
	out := cfg.Collections[:0]
	found := false
	for _, c := range cfg.Collections {
		if c.ID == id {
			found = true
			continue
		}
		out = append(out, c)
	}
	if !found {
		return fmt.Errorf("collection not found")
	}
	cfg.Collections = out
	return SaveConfig(cfg)
}

// RenameCollection sets the display name of a collection.
func RenameCollection(id, name string) (*Collection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name required")
	}
	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		return nil, fmt.Errorf("no config")
	}
	for i := range cfg.Collections {
		if cfg.Collections[i].ID == id {
			cfg.Collections[i].Name = name
			if err := SaveConfig(cfg); err != nil {
				return nil, err
			}
			c := cfg.Collections[i]
			return &c, nil
		}
	}
	return nil, fmt.Errorf("collection not found")
}

// AddPathToCollection adds a dossier folder path to a collection (deduped).
func AddPathToCollection(id, path string) (*Collection, error) {
	path = WorkspaceKey(path)
	if path == "" || path == "." {
		return nil, fmt.Errorf("path required")
	}
	if !IsDossierFolder(path) {
		return nil, fmt.Errorf("not a dossier folder: %s", path)
	}
	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		return nil, fmt.Errorf("no config")
	}
	for i := range cfg.Collections {
		if cfg.Collections[i].ID != id {
			continue
		}
		for _, p := range cfg.Collections[i].Paths {
			if strings.EqualFold(WorkspaceKey(p), path) {
				c := cfg.Collections[i]
				return &c, nil
			}
		}
		cfg.Collections[i].Paths = append(cfg.Collections[i].Paths, path)
		if err := SaveConfig(cfg); err != nil {
			return nil, err
		}
		c := cfg.Collections[i]
		return &c, nil
	}
	return nil, fmt.Errorf("collection not found")
}

// RemovePathFromCollection removes a path from a collection.
func RemovePathFromCollection(id, path string) (*Collection, error) {
	path = WorkspaceKey(path)
	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		return nil, fmt.Errorf("no config")
	}
	for i := range cfg.Collections {
		if cfg.Collections[i].ID != id {
			continue
		}
		out := cfg.Collections[i].Paths[:0]
		for _, p := range cfg.Collections[i].Paths {
			if strings.EqualFold(WorkspaceKey(p), path) {
				continue
			}
			out = append(out, p)
		}
		cfg.Collections[i].Paths = out
		if err := SaveConfig(cfg); err != nil {
			return nil, err
		}
		c := cfg.Collections[i]
		return &c, nil
	}
	return nil, fmt.Errorf("collection not found")
}

// CollectionContaining returns the first collection that contains path, or nil.
func CollectionContaining(path string) *Collection {
	path = WorkspaceKey(path)
	for _, c := range ListCollections() {
		for _, p := range c.Paths {
			if strings.EqualFold(WorkspaceKey(p), path) {
				cc := c
				return &cc
			}
		}
	}
	return nil
}

// CollectionMemberViews builds portfolio rows for a collection (name, path, last opened).
func CollectionMemberViews(col Collection) []map[string]string {
	cfg, _ := LoadConfig()
	if cfg == nil {
		cfg = &AppConfig{}
	}
	out := make([]map[string]string, 0, len(col.Paths))
	for _, p := range col.Paths {
		p = filepath.Clean(p)
		row := map[string]string{
			"path":        p,
			"name":        cfg.DisplayNameFor(p),
			"folderName":  filepath.Base(p),
			"lastOpened":  cfg.LastOpenedFor(p),
			"isDossier":   "false",
		}
		if IsDossierFolder(p) {
			row["isDossier"] = "true"
		}
		out = append(out, row)
	}
	return out
}
