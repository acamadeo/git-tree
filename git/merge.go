package gitutil

import (
	"fmt"

	git "github.com/libgit2/git2go/v34"
)

// TODO: DO NOT SUBMIT UNTIL deleting this whole thing!!
// MergeBaseOctopus() seems to actually work, it was a (skill)-issue...

// Find the best common ancestor of all supplied commits.
//
// Use this instead of libgit2's MergeBaseOctopus, which does not actually
// provide an ancestor across *all* commits.
func MergeBaseOctopus(repo *git.Repository, oids []*git.Oid) (*git.Oid, error) {
	if len(oids) == 0 {
		return nil, fmt.Errorf("MergeBaseOctopus received empty oids")
	}

	// TODO: remove debug print
	fmt.Println("MergeBaseOctopus: Oids [")
	for _, oid := range oids {
		fmt.Printf("\t%s,\n", oid.String())
	}
	fmt.Println("]")

	// Find the common ancestor by successively getting the ancestor of each
	// pair of Oids.
	for len(oids) > 1 {
		// TODO: remove debug print
		fmt.Println("MergeBaseOctopus: Oids [")
		for _, oid := range oids {
			fmt.Printf("\t%s,\n", oid.String())
		}
		fmt.Println("]")
		// END: debug print
		newOids := []*git.Oid{}

		// Find the common ancestor of each pair of Oids.
		for i := 0; i < len(oids)/2; i++ {
			base, err := repo.MergeBase(oids[2*i], oids[2*i+1])
			if err != nil {
				return nil, err
			}
			newOids = append(newOids, base)
		}

		// Always include the last Oid, if there's an odd number.
		if len(oids)%2 == 1 {
			newOids = append(newOids, oids[len(oids)-1])
		}
		oids = newOids
	}
	return oids[0], nil
}
