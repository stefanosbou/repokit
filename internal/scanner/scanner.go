package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindRepos returns absolute paths of every directory that contains a `.git`
// child, descending up to maxDepth levels below root. Hidden directories
// (starting with ".") and `node_modules` are skipped. The walk does not
// descend into a repo's own `.git` directory.
func FindRepos(root string, maxDepth int) ([]string, error) {
	var found []string
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	err = filepath.WalkDir(rootAbs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsPermission(walkErr) {
				return nil
			}
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(rootAbs, path)
		depth := 0
		if rel != "." {
			depth = strings.Count(rel, string(os.PathSeparator)) + 1
		}
		base := filepath.Base(path)

		if depth > 0 && (base == ".git" || base == "node_modules" || (strings.HasPrefix(base, ".") && base != ".")) {
			return filepath.SkipDir
		}
		if info, err := os.Stat(filepath.Join(path, ".git")); err == nil && (info.IsDir() || info.Mode().IsRegular()) {
			found = append(found, path)
			return filepath.SkipDir
		}
		if depth >= maxDepth {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}
