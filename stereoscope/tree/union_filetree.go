package tree

import (
	"github.com/anchore/stereoscope/stereoscope/file"
)

type UnionFileTree struct {
	trees []*FileTree
}

func NewUnionTree() *UnionFileTree {
	return &UnionFileTree{
		trees: make([]*FileTree, 0),
	}
}

func (u *UnionFileTree) PushTree(t *FileTree) {
	u.trees = append(u.trees, t)
}

func (u *UnionFileTree) Squash() (*FileTree, error) {
	switch len(u.trees) {
	case 0:
		return NewFileTree(), nil
	case 1:
		return u.trees[0].Copy()
	}

	var squashedTree *FileTree
	var err error
	for layerIdx, refTree := range u.trees {
		if layerIdx == 0 {
			squashedTree, err = refTree.Copy()
			if err != nil {
				return nil, err
			}
			continue
		}

		conditions := WalkConditions{
			ShouldContinueBranch: refTree.ConditionFn(func(f *file.File) bool {
				return !f.Path.IsWhiteout()
			}),
		}

		visitor := refTree.VisitorFn(func(f *file.File) {
			if f.Path.IsWhiteout() {
				lowerPath, err := f.Path.UnWhiteoutPath()
				if err != nil {
					// TODO: replace
					panic(err)
				}

				err = squashedTree.RemovePath(lowerPath)
				if err != nil {
					// TODO: replace
					panic(err)
				}
			} else {
				if !squashedTree.HasPath(f.Path) {
					_, err := squashedTree.AddPath(f.Path)
					if err != nil {
						// TODO: replace
						panic(err)
					}
				}
				err := squashedTree.SetFile(f)
				if err != nil {
					// TODO: replace
					panic(err)
				}
			}
		})

		w := NewDepthFirstWalkerWithConditions(refTree.Reader(), visitor, conditions)
		w.WalkAll()

	}
	return squashedTree, nil
}