package errorx

import "errors"

func Wrap(err error) *customError {
	return wrapDepth(err, 4)
}

func wrapDepth(err error, depth int) *customError {
	if err == nil {
		return nil
	}
	return &customError{
		err:   err,
		stack: callers(depth),
		data:  make(map[string]any),
	}
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}
