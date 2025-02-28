package models

type ErrNotFound struct{ Err error }

func (e ErrNotFound) Error() string { return e.Err.Error() }

type ErrClientError struct{ Err error }

func (e ErrClientError) Error() string { return e.Err.Error() }

func IsClientError(e error) bool {
	if e == nil {
		return false
	}

	if _, ok := e.(*ErrClientError); ok {
		return true
	}
	return false
}

func IsNotFound(e error) bool {
	if e == nil {
		return false
	}

	if cerr, ok := e.(*ErrClientError); ok {
		if _, ok := cerr.Err.(*ErrNotFound); ok {
			return true
		}
	}
	return false
}
