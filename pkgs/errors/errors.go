package errors

import "fmt"

// ContextW should be []interface{"key", "value"}
type FmtError struct{
	responseMsg string
	impact string
	cause string
	err error
	contextInfo interface{}
}

func NewFmtError(respMsg, impact, cause string, err error, ctxInfo ...interface{}) *FmtError{
	return &FmtError{
		responseMsg: respMsg, impact: impact, cause: cause, err: err, contextInfo: ctxInfo,
	}
}

func (er *FmtError) Error() string {
	return fmt.Sprintf("{ Impact: %s, Cause: %s, Error: %v, contextInfo: %v }",
	er.impact, er.cause, er.err, er.contextInfo)
}

func (er *FmtError) GetResponseMsg() string {
	return er.responseMsg
}
