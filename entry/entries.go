package entry

import (
	"os"
	"path"
	"path/filepath"
)

type Entries map[string]Entry

func (instance *Entries) Find(pathname string) *Entry {
	if instance == nil || *instance == nil {
		return nil
	}

	cleanedPath := CleanPath(pathname)
	if entry, ok := (*instance)[cleanedPath]; ok {
		return &entry
	} else {
		return nil
	}
}

func (instance *Entries) Get(pathname string) (Entry, error) {
	if instance == nil || *instance == nil {
		return Entry{}, os.ErrNotExist
	}

	cleanedPath := CleanPath(pathname)
	if entry, ok := (*instance)[cleanedPath]; !ok {
		return Entry{}, os.ErrNotExist
	} else {
		return entry, nil
	}
}

func (instance *Entries) Add(pathname string, entry Entry) error {
	if instance == nil || *instance == nil {
		*instance = make(Entries)
	}

	cleanedPath := CleanPath(pathname)
	if _, alreadyContained := (*instance)[cleanedPath]; alreadyContained {
		return os.ErrExist
	}
	(*instance)[cleanedPath] = entry
	return nil
}

func (instance *Entries) Replace(pathname string, entry Entry) error {
	if instance == nil || *instance == nil {
		*instance = make(Entries)
	}

	cleanedPath := CleanPath(pathname)
	if _, alreadyContained := (*instance)[cleanedPath]; !alreadyContained {
		return os.ErrNotExist
	}
	(*instance)[cleanedPath] = entry
	return nil
}

func (instance Entries) Filter(predicate Predicate) (Entries, error) {
	if instance == nil {
		return Entries{}, nil
	}

	result := make(Entries)

	for p, entry := range instance {
		if predicate == nil {
			result[p] = entry
		} else if matches, err := predicate(p, entry); err != nil {
			return Entries{}, err
		} else if matches {
			result[p] = entry
		}
	}

	return result, nil
}

func CleanPath(in string) string {
	if len(in) <= 0 {
		return in
	}
	if in[0] == '/' {
		in = in[1:]
	}
	return path.Clean(filepath.ToSlash(in))
}
