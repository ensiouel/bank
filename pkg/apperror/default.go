package apperror

var (
	Unknown       = New("unknown error")
	Internal      = New("internal error")
	NotFound      = New("not found")
	AlreadyExists = New("already exists")
	BadRequest    = New("bad request")
	Unauthorized  = New("unauthorized")
	Forbidden     = New("forbidden")
)
