package codec

import (
	"reflect"
	"time"

	"github.com/horm-database/common/codec/mapstructure"
	"github.com/horm-database/common/json"
)

var DefaultMap2Structure = &Map2Structure{DefaultTag, true, true}

type Map2Structure struct {
	tagName    string
	weaklyType bool
	squash     bool
}

var typeTime = reflect.TypeOf(time.Time{})
var typeString = reflect.TypeOf("")

func (m *Map2Structure) Decode(src, dest interface{}) error {
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: m.weaklyType,
		Squash:           m.squash,
		Result:           dest,
		TagName:          m.tagName,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if t == typeTime && f == typeString {
					tt := time.Time{}
					err := json.Api.Unmarshal([]byte(`"`+data.(string)+`"`), &tt)
					return tt, err
				}

				return data, nil
			},
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(src)
}
