package headers

type Headers map[string]string

func NewHeaders() *Headers {
	return &Headers{}
}

// Parse will parse a byte array and return:
// the number of bytes read and a done bool
func (h Headers) Parse(data []byte) (int, bool, error) {
}
