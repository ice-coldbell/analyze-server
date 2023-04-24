package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type Field = zapcore.Field

type ObservedLogs = observer.ObservedLogs

var (
	Any         = zap.Any
	Array       = zap.Array
	Binary      = zap.Binary
	Bool        = zap.Bool
	Boolp       = zap.Boolp
	ByteString  = zap.ByteString
	Complex128  = zap.Complex128
	Complex128p = zap.Complex128p
	Complex64   = zap.Complex64
	Complex64p  = zap.Complex64p
	Duration    = zap.Duration
	Durationp   = zap.Durationp
	Float32     = zap.Float32
	Float32p    = zap.Float32p
	Float64     = zap.Float64
	Float64p    = zap.Float64p
	Inline      = zap.Inline
	Int         = zap.Int
	Int16       = zap.Int16
	Int16p      = zap.Int16p
	Int32       = zap.Int32
	Int32p      = zap.Int32p
	Int64       = zap.Int64
	Int64p      = zap.Int64p
	Int8        = zap.Int8
	Int8p       = zap.Int8p
	Intp        = zap.Intp
	Namespace   = zap.Namespace
	Object      = zap.Object
	Reflect     = zap.Reflect
	Skip        = zap.Skip
	Stack       = zap.Stack
	StackSkip   = zap.StackSkip
	String      = zap.String
	Stringer    = zap.Stringer
	Stringp     = zap.Stringp
	Time        = zap.Time
	Timep       = zap.Timep
	Uint        = zap.Uint
	Uint16      = zap.Uint16
	Uint16p     = zap.Uint16p
	Uint32      = zap.Uint32
	Uint32p     = zap.Uint32p
	Uint64      = zap.Uint64
	Uint64p     = zap.Uint64p
	Uint8       = zap.Uint8
	Uint8p      = zap.Uint8p
	Uintp       = zap.Uintp
	Uintptr     = zap.Uintptr
	Uintptrp    = zap.Uintptrp
)
