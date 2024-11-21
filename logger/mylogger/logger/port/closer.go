package port

type Closer interface {
	Close() error
}
