package facades

import "github.com/nd-tools/capyvel/responses"

func Response() *responses.Response {
	return responses.Handler
}
