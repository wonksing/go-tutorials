package port

type Syncer interface {
	Sync() error
}
