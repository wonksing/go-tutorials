package port

type Roller interface {
	Write(p []byte) (n int, err error)
	Close() error
	Rotate() error
	Sync() error

	GetPath() string
}
