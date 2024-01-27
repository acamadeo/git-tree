package gitutil

import (
	git "github.com/libgit2/git2go/v34"
)

func OidShortHash(oid git.Oid) string {
	return oid.String()[:shortHashLength]
}

// Returns 0 if Oid values are the same. Otherwise, returns -1 if a < b, or +1 if a > b.
func compareOids(a git.Oid, b git.Oid) int {
	for i := 0; i < 20; i++ {
		byteA, byteB := a[i], b[i]
		if byteA != byteB {
			if int(byteA)-int(byteB) < 0 {
				return -1
			}
			return 1
		}
	}
	return 0
}
