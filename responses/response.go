package responses

var (
	Handler *Response
)

func Boot() {
	Handler = &Response{
		Api: Api{},
	}
}

type Response struct {
	Api Api
}
