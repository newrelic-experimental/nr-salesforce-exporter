package connect

import "fmt"

type ConnectError struct {
	Err     error
	ErrCode int
}

func (ce *ConnectError) Error() string {
	if ce.Err != nil {
		return fmt.Sprintf("connection error: code %d: %s", ce.ErrCode, ce.Err.Error())
	}
	return fmt.Sprintf("connection error: code %d", ce.ErrCode)
}

func MakeConnectErr(err error, errCode int) *ConnectError {
	return &ConnectError{
		Err:     err,
		ErrCode: errCode,
	}
}

type Connector interface {
	SetConfig(any)
	Request() ([]byte, error)
	ConnectorID() string
}
