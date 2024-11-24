package codec

import (
	"github.com/json-iterator/go"
)

type Marshal func(v interface{}) ([]byte, error)
type Unmarshal func(data []byte, v interface{}) error

type Codec interface {
	SetMarshal(m Marshal, um Unmarshal) // marshal/unmarshal for single item
	GetMarshal() Marshal                // get marshal
	GetUnmarshal() Unmarshal            // get unmarshal

	SetTag(tag string) // set tag
	GetTag() string    // get tag

	Encode(typ int, v interface{}) (interface{}, error)        // encode
	Decode(typ int, src interface{}, dest []interface{}) error // decode
}

var (
	DefaultTag   = "orm"
	DefaultCodec = NewDefaultCodec(jsonMarshal, jsonUnmarshal, DefaultTag, true, true, true)
)

var marshalConfig = jsoniter.Config{
	EscapeHTML: true,
	TagKey:     "json",
}.Froze()

var (
	jsonMarshal   = func(v interface{}) ([]byte, error) { return marshalConfig.Marshal(v) }
	jsonUnmarshal = func(data []byte, v interface{}) error { return marshalConfig.Unmarshal(data, v) }
)

func NewDefaultCodec(m Marshal, um Unmarshal, tag string, omitEmpty, weaklyType, squash bool) Codec {
	c := defaultCodec{}
	c.m = m
	c.um = um
	c.tag = tag
	c.omitEmpty = omitEmpty

	c.map2structure = &Map2Structure{
		tagName:    tag,
		squash:     squash,
		weaklyType: weaklyType,
	}

	return &c
}

// SetDefaultTag change the default tag
func SetDefaultTag(tag string) {
	DefaultTag = tag

	marshalConfig = jsoniter.Config{
		EscapeHTML: true,
		TagKey:     tag,
	}.Froze()

	jsonMarshal = func(v interface{}) ([]byte, error) { return marshalConfig.Marshal(v) }
	jsonUnmarshal = func(data []byte, v interface{}) error { return marshalConfig.Unmarshal(data, v) }

	DefaultCodec = NewDefaultCodec(jsonMarshal, jsonUnmarshal, tag, true, true, true)
}
