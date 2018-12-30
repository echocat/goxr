package usagescanner

import (
	"path/filepath"
)

type Usages map[string][]string

func (instance Usages) Resolve() []string {
	if instance == nil {
		return []string{}
	}

	buffer := make(map[string]bool)
	for source, usages := range instance {
		for _, usage := range usages {
			usage = filepath.Clean(filepath.FromSlash(usage))
			if !filepath.IsAbs(usage) {
				usage = filepath.Join(filepath.Dir(source), usage)
				usage = filepath.Clean(usage)
			}
			buffer[usage] = true
		}
	}

	result := make([]string, len(buffer))
	i := 0
	for usage := range buffer {
		result[i] = usage
		i++
	}

	return result
}
