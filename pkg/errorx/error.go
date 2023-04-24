package errorx

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(message string) *customError {
	return &customError{
		err:   errors.New(message),
		stack: callers(3),
		data:  make(errorData),
	}
}

type errorData map[string]any

func (ed errorData) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range ed {
		switch value := v.(type) {
		case interface{ Int() int }:
			enc.AddInt(k, value.Int())
		case interface{ String() string }:
			enc.AddString(k, value.String())
		default:
			if err := enc.AddReflected(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

type customError struct {
	err   error
	stack *stack
	data  errorData
}

func (e *customError) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if e.err == nil {
		return nil
	}
	switch tErr := e.err.(type) {
	case *customError:
		zap.Object("error", tErr).AddTo(enc)
	case customErrors:
		zap.Array("errors", tErr).AddTo(enc)
	default:
		enc.AddString("message", e.Error())
	}
	enc.AddString("stack", fmt.Sprintf("%+v", e.stack))
	if len(e.data) > 0 {
		zap.Object("data", e.data).AddTo(enc)
	}
	return nil
}

func (e *customError) Error() string {
	return e.err.Error()
}

func (e *customError) Unwrap() error {
	return e.err
}

func (e *customError) With(key string, data any) *customError {
	e.data[key] = data
	return e
}

type customErrors []*customError

func (e customErrors) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, err := range e {
		if err == nil {
			continue
		}
		if err := enc.AppendObject(err); err != nil {
			return err
		}
	}
	return nil
}

func (e customErrors) Error() string {
	var b []byte
	for i, err := range e {
		if i > 0 {
			b = append(b, '\n')
		}
		b = append(b, err.Error()...)
	}
	return string(b)
}

func (e customErrors) Unwrap() []error {
	errors := make([]error, 0, len(e))
	for i := range e {
		errors = append(errors, e[i])
	}
	return errors
}
