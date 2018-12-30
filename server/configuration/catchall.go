package configuration

import (
	"fmt"
	"github.com/blaubaer/goxr"
	"os"
	"regexp"
)

type Catchall struct {
	Target   *string   `yaml:"target,omitempty"`
	Includes *[]string `yaml:"includes,omitempty"`
	Excludes *[]string `yaml:"excludes,omitempty"`

	includesRegexpCache *[]*regexp.Regexp
	excludesRegexpCache *[]*regexp.Regexp
}

func (instance Catchall) GetTarget() string {
	r := instance.Target
	if r == nil {
		return ""
	}
	return *r
}

func (instance Catchall) GetIncludes() []string {
	r := instance.Includes
	if r == nil {
		return []string{}
	}
	return *r
}

func (instance Catchall) GetExcludes() []string {
	r := instance.Excludes
	if r == nil {
		return []string{
			regexp.QuoteMeta("/favicon.ico"),
			regexp.QuoteMeta("/robots.txt"),
		}
	}
	return *r
}

func (instance *Catchall) IsEligible(candidate string) (bool, error) {
	includes := instance.includesRegexpCache
	excludes := instance.excludesRegexpCache
	for i := 0; i < 100 && includes == nil; i++ {
		if errs := instance.rebuildIncludesCache(); len(errs) > 0 {
			return false, errs[0]
		}
		includes = instance.includesRegexpCache
	}
	for i := 0; i < 100 && excludes == nil; i++ {
		if errs := instance.rebuildExcludesCache(); len(errs) > 0 {
			return false, errs[0]
		}
		excludes = instance.excludesRegexpCache
	}

	if includes != nil && len(*includes) > 0 {
		foundMatch := false
		for _, r := range *includes {
			if r.MatchString(candidate) {
				foundMatch = true
			}
		}
		if !foundMatch {
			return false, nil
		}
	}

	if excludes != nil && len(*excludes) > 0 {
		foundMatch := false
		for _, r := range *excludes {
			if r.MatchString(candidate) {
				foundMatch = true
			}
		}
		if foundMatch {
			return false, nil
		}
	}

	return true, nil
}

func (instance *Catchall) Validate(using goxr.Box) (errors []error) {
	errors = append(errors, instance.validateTarget(using)...)
	errors = append(errors, instance.rebuildIncludesCache()...)
	errors = append(errors, instance.rebuildExcludesCache()...)
	return
}

func (instance *Catchall) validateTarget(using goxr.Box) (errors []error) {
	r := instance.GetTarget()
	if r == "" {
		return
	}
	if _, err := using.Info(r); os.IsNotExist(err) {
		errors = append(errors, fmt.Errorf(`paths.catchall.target = "%s" - path does not exist in box`, r))
	} else if err != nil {
		errors = append(errors, fmt.Errorf(`paths.catchall.target = "%s" - cannot read path information: %v`, r, err))
	}
	return
}

func (instance *Catchall) rebuildIncludesCache() (errors []error) {
	r := instance.GetIncludes()
	rs := make([]*regexp.Regexp, len(r))
	for i, pattern := range r {
		if crx, err := regexp.Compile(pattern); err != nil {
			errors = append(errors, fmt.Errorf(`paths.catchall.includes[%d]= "%s" - pattern invalid: %v`, i, pattern, err))
		} else {
			rs[i] = crx
		}
	}
	instance.includesRegexpCache = &rs
	return
}

func (instance *Catchall) rebuildExcludesCache() (errors []error) {
	r := instance.GetExcludes()
	rs := make([]*regexp.Regexp, len(r))
	for i, pattern := range r {
		if crx, err := regexp.Compile(pattern); err != nil {
			errors = append(errors, fmt.Errorf(`paths.catchall.excludes[%d]= "%s" - pattern invalid: %v`, i, pattern, err))
		} else {
			rs[i] = crx
		}
	}
	instance.excludesRegexpCache = &rs
	return
}
