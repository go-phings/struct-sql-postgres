package structsqlpostgres

// ErrStructSQL wraps original error with operation/step where the error occurred and optionally with a tag when
// parsing "crud" failed
type ErrStructSQL struct {
	Op  string
	Tag string
	Err error
}

func (e ErrStructSQL) Error() string {
	return e.Err.Error()
}

func (e ErrStructSQL) Unwrap() error {
	return e.Err
}
