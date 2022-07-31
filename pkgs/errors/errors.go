package errors

import "fmt"

// ContextW should be []interface{"key", "value"}
type FmtError struct{
	Impact string
	Context string
	Err error
	ContextInfo interface{}
}

func NewFmtError(contx, impact string, err error, ctxInfo []interface{}) *FmtError{
	return &FmtError{
		Impact: impact, Context: contx ,Err: err, ContextInfo: ctxInfo,
	}
}

func (er *FmtError) Error() string {
	return fmt.Sprintf("{ Error: %s, Context: %s, Impact: %s, contextInfo: %v }",
	er.Err.Error(), er.Context, er.Impact, er.ContextInfo)
}
