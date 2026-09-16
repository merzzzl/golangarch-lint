package dto

type Scope struct {
	Root            string
	Files           map[string][]int
	Directories     map[string][]int
	FileDirectories map[string][]int
	Subdirs         map[string][]int
}
