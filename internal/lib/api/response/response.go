package response

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Id     int64  `json:"id,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func Id(id int64) Response {
	return Response{
		Status: StatusOK,
		Id:     id,
	}
}
