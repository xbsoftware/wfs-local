package wfs

import (
	"path/filepath"
	"strings"
)

// CombinedPolicy allows to join multiple policies together
type CombinedPolicy struct {
	Policies []Policy
}

// Comply method returns true only when all joined policies are complied
func (p CombinedPolicy) Comply(path string, operation Operation) bool {
	for _, el := range p.Policies {
		if el.Comply(path, operation) == false {
			return false
		}
	}
	return true
}

// ReadOnlyPolicy allows read access and blocks any modifications
type ReadOnlyPolicy struct{}

// Comply method returns true for read operations
func (p ReadOnlyPolicy) Comply(path string, operation Operation) bool {
	return operation == ReadOperation
}

// ForceRootPolicy prevents any operations outside of data root
type ForceRootPolicy struct {
	Root string
}

// Comply method returns true when path is inside of root
func (p ForceRootPolicy) Comply(path string, operation Operation) bool {
	return strings.Contains(filepath.Clean(path), p.Root)
}

// AllowPolicy allows all operations
type AllowPolicy struct{}

// Comply method returns true
func (p AllowPolicy) Comply(path string, operation Operation) bool {
	return true
}

// DenyPolicy allows all operations
type DenyPolicy struct{}

// Comply method returns false
func (p DenyPolicy) Comply(path string, operation Operation) bool {
	return false
}
