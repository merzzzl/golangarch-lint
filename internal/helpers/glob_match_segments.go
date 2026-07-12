package helpers

import "path"

func GlobMatchSegments(pat, name []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			pat = pat[1:]
			if len(pat) == 0 {
				return true
			}

			for i := range len(name) + 1 {
				if GlobMatchSegments(pat, name[i:]) {
					return true
				}
			}

			return false
		}

		if len(name) == 0 {
			return false
		}

		matched, _ := path.Match(pat[0], name[0])
		if !matched {
			return false
		}

		pat = pat[1:]
		name = name[1:]
	}

	return len(name) == 0
}
