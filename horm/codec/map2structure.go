package codec

import (
	"github.com/horm-database/common/codec/mapstructure"
)

var DefaultMap2Structure = &Map2Structure{DefaultTag, true, true}

type Map2Structure struct {
	tagName    string
	weaklyType bool
	squash     bool
}

func (m *Map2Structure) Decode(src, dest interface{}) error {
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: m.weaklyType,
		Squash:           m.squash,
		Result:           dest,
		TagName:          m.tagName,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(src)
}
