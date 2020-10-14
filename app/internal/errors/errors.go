package errors

import "golang.org/x/xerrors"

var (
	InternalError = xerrors.New("Internal error")
	NotFound      = xerrors.New("Not found")
	AlreadyExist  = xerrors.New("Already exist")
)
