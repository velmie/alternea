package manipulation

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrUnsupportedDataType = Error("unsupported data type")
)
