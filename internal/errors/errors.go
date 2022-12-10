package errors

type Kind uint8

const (
	Other Kind = iota
	Forbidden
	Validation
	NotExist
)

type AppError struct {
	Kind Kind
	Err  error
	msg  string
}

func (e AppError) Error() string {
	return e.msg
}

func E(args ...interface{}) AppError {
	e := AppError{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Kind:
			e.Kind = arg
		case string:
			e.msg = arg
		case error:
			e.Err = arg
		}
	}

	return e
}
