package handlers

type HandlerResponse struct {
	Err error
	ErrCode int
	StatusCode int
}

func MakeHandlerResponse(err error, errCode, statusCode int) *HandlerResponse {
	return &HandlerResponse{
		Err: err, ErrCode: errCode, StatusCode: statusCode,
	}
}

// error codes
const (
	NoError int = 0
	FailedSetDS1 int = 1
	FailedPartiallySetDS1 int = 2
	FailedAddBlockUsrDS1 int = 3
	FailedCreateNewUsrDS1 int = 4
	FailedRmBlockUserDs1 int = 5
	FailedCreateNewUsrDS2 int = 6
	SomeErrCreateNewUsrDS2 int = 7 // unsuccessful history table creation, some errors occuried
	FailedMvLastReadDS2 int = 8
)