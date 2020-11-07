package parameter

import (
	"encoding/json"
	"io"
)

// RunParam is a set of parameters for run command
type RunParam struct {
	Template string `json:"template"`
}

func Decode(body io.Reader, param interface{}) {
	json.NewDecoder(body).Decode(&param)
}
