package graph

import (
	"fmt"

	appErrors "abacus/internal/errors"
)

func cyclicDependencyError(path []string) error {
	return appErrors.New(appErrors.CodeCyclicDependency, fmt.Sprintf("cyclic dependency detected: %v", path), nil)
}
