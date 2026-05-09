package registry

import (
	"github.com/stefanosbou/repokit/internal/config"
)

// Registry is a thin facade over a *config.Config that provides repo CRUD
// All mutations operate directly on the underlying
// config; callers are responsible for persisting via config.Save.
type Registry struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Registry {
	return &Registry{cfg: cfg}
}

// Add appends a repo entry. Returns false if the path is already registered.
func (r *Registry) Add(e config.RepoEntry) bool {
	for _, existing := range r.cfg.Repos {
		if existing.Path == e.Path {
			return false
		}
	}
	r.cfg.Repos = append(r.cfg.Repos, e)
	return true
}

// Remove deletes the entry with the given name. Returns true if anything was removed.
func (r *Registry) Remove(name string) bool {
	for i, e := range r.cfg.Repos {
		if e.Name == name {
			r.cfg.Repos = append(r.cfg.Repos[:i], r.cfg.Repos[i+1:]...)
			return true
		}
	}
	return false
}

// All returns a copy of every registered repo.
func (r *Registry) All() []config.RepoEntry {
	out := make([]config.RepoEntry, len(r.cfg.Repos))
	copy(out, r.cfg.Repos)
	return out
}

// ByName returns the repo with the given name.
func (r *Registry) ByName(name string) (config.RepoEntry, bool) {
	for _, e := range r.cfg.Repos {
		if e.Name == name {
			return e, true
		}
	}
	return config.RepoEntry{}, false
}
