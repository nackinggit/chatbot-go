package bcode

type BError struct {
	code    int
	message string
}

func (b BError) Code() int {
	return b.code
}

func (b BError) Message() string {
	return b.message
}

func (b BError) Error() string {
	return b.message
}

func New(code int, message string) BError {
	return BError{
		code:    code,
		message: message,
	}
}
