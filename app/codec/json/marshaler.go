package json

import (
	"encoding/json"

	"github.com/asim/nitro/util/buf"
)

// create buffer pool with 16 instances each preallocated with 256 bytes
var bufferPool = buf.NewPool()

type Marshaler struct{}

func (j Marshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j Marshaler) Unmarshal(d []byte, v interface{}) error {
	return json.Unmarshal(d, v)
}

func (j Marshaler) String() string {
	return "json"
}
