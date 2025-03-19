package parser

import "errors"

type errorManager struct {
	errors []error
}

func newErrorManager() *errorManager {
	return &errorManager{}
}

func (e *errorManager) add(err error) {
	if err == nil {
		return
	}
	e.errors = append(e.errors, err)
}

func (e *errorManager) error() error {
	if len(e.errors) == 0 {
		return nil
	}
	return errors.Join(e.errors...)
}
