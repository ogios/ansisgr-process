package process

const (
	ESCAPE_SEQUENCE     = '\x1b'
	SGR_FUNC            = byte('\x6d')
	ESCAPE_SEQUENCE_END = string(ESCAPE_SEQUENCE) + "[0" + string(SGR_FUNC)
)

func IsEscEnd(c rune) bool {
	return (c >= 0x40 && c <= 0x5a) || (c >= 0x61 && c <= 0x7a)
}

func IsSGR(data []byte) bool {
	return data[len(data)-1] == SGR_FUNC
}

func IsEndOfSGR(data []byte) bool {
	return string(data) == ESCAPE_SEQUENCE_END
}
