package pb

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

// SubtypeJSON is the gRPC content-subtype used by this project.
const SubtypeJSON = "json"

type jsonCodec struct{}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return SubtypeJSON
}

func init() {
	encoding.RegisterCodec(jsonCodec{})
}
