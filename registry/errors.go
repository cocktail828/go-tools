package registry

import "errors"

var (
	ErrNoRegister     = errors.New("no service register is set")
	ErrNoConfiger     = errors.New("no config finder is set")
	ErrNoAvailNode    = errors.New("no avaliable service instance found")
	ErrNotRegistered  = errors.New("cannot deregister for service is not registered")
	ErrMissingName    = errors.New("missing service name")
	ErrMissingVersion = errors.New("missing service version")
	ErrMissingPort    = errors.New("missing service port")
)
