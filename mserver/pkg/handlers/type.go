package handlers

import (
	"fmt"

	ds2repo "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/data_source2"
)

type HandlerResponse struct {
	Err        error
	ErrCode    int
	StatusCode int
}

func MakeHandlerResponse(err error, errCode, statusCode int) *HandlerResponse {
	return &HandlerResponse {
		Err: err, ErrCode: errCode, StatusCode: statusCode,
	}
}

type RequestIgnore struct {
	reason string
}

func (ri RequestIgnore) Error() string {
	return fmt.Sprintf("request is ignore due to %s", ri.reason)
}

func IsRequestIgnore(err error) bool {
	_, ok := err.(RequestIgnore)
	return ok
}

type Info struct {
	Metadata *ds2repo.HistTbMetadata
	HistId string
	Prop string
}

// for intermidiate calculation. 
type SmPairHistInfo struct {
	UserId string  // toUserId
	TxLink *Info
	RxLink *Info
}

// error codes
const (
	NoError                int = 0
	FailedSetDS1           int = 1
	FailedPartiallySetDS1  int = 2
	FailedAddBlockUsrDS1   int = 3
	FailedCreateNewUsrDS1  int = 4
	FailedRmBlockUserDs1   int = 5
	FailedCreateNewUsrDS2  int = 6
	SomeErrCreateNewUsrDS2 int = 7 // unsuccessful history table creation, some errors occuried
	FailedMvLastReadDS2    int = 8
	FailedUpdateLastMsgDs2 int = 9
	FailedDeleteMsgDs2     int = 10
	NoJobToDo              int = 11  // leads to a request ignore
	FailedMsgsLoad 		   int = 12
	FaliedListPairHistTbsDS1 int = 13
)
