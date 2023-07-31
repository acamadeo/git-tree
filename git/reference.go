package gitutil

import (
	"encoding/hex"

	git "github.com/libgit2/git2go/v34"
)

func ReferenceShortHash(ref *git.Reference) string {
	oid := ref.Target()
	oidString := hex.EncodeToString(oid[:])
	return oidString[:shortHashLength]
}
