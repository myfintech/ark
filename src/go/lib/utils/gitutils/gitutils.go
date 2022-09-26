package gitutils

import (
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// FileDiff is a map with a key of the file name, and a value of the action taken on that file
type FileDiff map[string]string

// CompareHeadToBranch compare the head of the repo to a specific branch and return a FileDiff
func commonAncestors(head, branch *object.Commit) ([]*object.Commit, error) {
	var ancestors []*object.Commit
	switch {
	case head.Hash == branch.Hash:
		parent, err := head.Parent(0)
		if err != nil {
			return ancestors, err
		}
		ancestors = append(ancestors, parent)
	default:
		mergeAncestors, err := head.MergeBase(branch)
		if err != nil {
			return ancestors, err
		}
		ancestors = mergeAncestors
	}
	return ancestors, nil
}
func commitFromRev(repo *git.Repository, rev string) (*object.Commit, error) {
	hash, err := repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return nil, err
	}
	return repo.CommitObject(*hash)
}

var emptyChange = object.ChangeEntry{}

func getPathToChange(change *object.Change) string {
	if change.From != emptyChange {
		return change.From.Name
	}
	return change.To.Name
}
