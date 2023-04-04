package errorx

func Join(errs ...error) customErrors {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	var cErrs customErrors
	for _, err := range errs {
		switch tErr := err.(type) {
		case nil:
			continue
		case customErrors:
			cErrs = append(cErrs, tErr...)
		case *customError:
			cErrs = append(cErrs, tErr)
		default:
			cErrs = append(cErrs, wrapDepth(err, 4))
		}
	}
	return cErrs
}

func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}
