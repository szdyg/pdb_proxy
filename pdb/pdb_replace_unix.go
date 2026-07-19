//go:build !windows

package pdb

import "os"

func replaceCachedFile(source, target string) error {
	return os.Rename(source, target)
}
