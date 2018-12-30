package usagescanner

import (
	"os"
	"path/filepath"
)

func ScanForUsages(root string) (result Usages, err error) {
	result = make(Usages)
	if root, err = filepath.Abs(root); err != nil {
		return Usages{}, err
	}
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		visitor := newVisitor(path)
		if err := visitor.Run(); err != nil {
			return err
		}
		if len(visitor.Boxes) > 0 {
			result[path] = visitor.Boxes
		}
		return nil
	}); err != nil {
		return Usages{}, err
	}

	return result, nil
}
