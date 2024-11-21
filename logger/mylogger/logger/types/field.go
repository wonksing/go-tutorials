package types

type FieldType uint8

const (
	StringType FieldType = iota + 1
	BytesType
	Int32Type
	Int64Type
	Uint32Type
	Uint64Type
	AnyType
	UnknownType
)

type Field struct {
	Type        FieldType
	Name        string
	ValueString string
	ValueBytes  []byte
	ValueInt32  int32
	ValueInt64  int64
	ValueUint32 uint32
	ValueUint64 uint64
	ValueAny    interface{}
}

func WithStringField(fieldName string, fieldValue string) Field {
	return Field{
		Type:        StringType,
		Name:        fieldName,
		ValueString: fieldValue,
	}
}

func WithBytesField(fieldName string, fieldValue []byte) Field {
	return Field{
		Type:       BytesType,
		Name:       fieldName,
		ValueBytes: fieldValue,
	}
}

func WithInt32Field(fieldName string, fieldValue int32) Field {
	return Field{
		Type:       Int32Type,
		Name:       fieldName,
		ValueInt32: fieldValue,
	}
}

func WithInt64Field(fieldName string, fieldValue int64) Field {
	return Field{
		Type:       Int64Type,
		Name:       fieldName,
		ValueInt64: fieldValue,
	}
}

func WithUint32Field(fieldName string, fieldValue uint32) Field {
	return Field{
		Type:        Uint32Type,
		Name:        fieldName,
		ValueUint32: fieldValue,
	}
}

func WithUint64Field(fieldName string, fieldValue uint64) Field {
	return Field{
		Type:        Uint64Type,
		Name:        fieldName,
		ValueUint64: fieldValue,
	}
}

func WithAnyField(fieldName string, fieldValue any) Field {
	return Field{
		Type:     AnyType,
		Name:     fieldName,
		ValueAny: fieldValue,
	}
}
