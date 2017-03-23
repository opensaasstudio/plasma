package event

import (
	"encoding/json"

	"go.uber.org/zap/zapcore"
)

type MetaData struct {
	Type string `json:"type"`
}

type Payload struct {
	Meta MetaData        `json:"meta"`
	Data json.RawMessage `json:"data"`
}

func (p Payload) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", p.Meta.Type)
	enc.AddString("data", string(p.Data))

	return nil
}
