//go:build !windows

package conf

func defaultPdbDir() string {
	return "/opt/pdb"
}
