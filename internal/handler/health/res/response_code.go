package res

import "comb-dockerfile/internal/pkg/response"

const (
	ErrUnhealth    = 10001
)

func RegisterCode() {
	response.Register(ErrUnhealth, "unhealth")
}

