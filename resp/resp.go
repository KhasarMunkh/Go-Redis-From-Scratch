package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING = '+'
	ERROR  = '-'
	INT    = ':'
	BULK   = '$'
	ARRAY  = '*'
)

type Value struct {
	Typ   string
	Str   string
	Num   int
	Bulk  string
	Array []Value
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

func (r *Resp) readInt() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i), n, nil
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	// case STRING:
	// case ERROR:
	// case INT:
	default:
		fmt.Println("Unkown type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) readArray() (Value, error) {
	v := Value{Typ: "array"}
	length, _, err := r.readInt()
	if err != nil {
		return v, err
	}
	v.Array = make([]Value, length)
	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		v.Array[i] = val
	}
	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{Typ: "bulk"}
	length, _, err := r.readInt()
	if err != nil {
		return v, err
	}
	if length == -1 {
		v.Typ = "null"
		return v, nil
	}
	if length < 0 {
		return v, fmt.Errorf("invalid bulk length: %d", length)
	}
	bulk := make([]byte, length)
	// ensure we read exactly 'length' bytes
	if _, err := io.ReadFull(r.reader, bulk); err != nil {
		return v, fmt.Errorf("failed to read bulk data: %w", err)
	}
	// r.reader.Read(bulk)
	v.Bulk = string(bulk)
	// read trailing CLRF
	if _, _, err := r.readLine(); err != nil {
		return v, err
	}

	return v, nil
}

func (v Value) Marshal() []byte {
	switch v.Typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshalNull()
	case "error":
		return v.marshalError()
	case "int":
		return v.marshalInt()
	default:
		return []byte{}
	}
}

// recursive marshal for array
func (v Value) marshalArray() []byte {
	l := len(v.Array)
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = strconv.AppendInt(bytes, int64(l), 10)
	bytes = append(bytes, '\r', '\n')
	for _, elem := range v.Array {
		bytes = append(bytes, elem.Marshal()...)
	}

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = strconv.AppendInt(bytes, int64(len(v.Bulk)), 10)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.Bulk...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

func (v Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalInt() []byte {
	var bytes []byte
	bytes = append(bytes, INT)
	bytes = strconv.AppendInt(bytes, int64(v.Num), 64)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	bytes := v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
