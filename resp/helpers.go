package resp

func NewSimpleString(s string) Value {
	return Value{Typ: "string", Str: s}
}

func NewBulkString(s string) Value {
	return Value{Typ: "bulk", Bulk: s}
}

func NewError(msg string) Value {
	return Value{Typ: "error", Str: msg}
}

func NewNull() Value {
	return Value{Typ: "null"}
}

func NewInteger(n int) Value {
	return Value{Typ: "int", Num: n}
}
