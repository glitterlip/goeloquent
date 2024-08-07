package goeloquent

const (
	EventSaving      = "EloquentSaving"
	EventSaved       = "EloquentSaved"
	EventCreating    = "EloquentCreating"
	EventCreated     = "EloquentCreated"
	EventUpdating    = "EloquentUpdating"
	EventUpdated     = "EloquentUpdated"
	EventDeleteing   = "EloquentDeleting"
	EventDeleted     = "EloquentDeleted"
	EventRetrieved   = "EloquentRetrieved"
	EventRetrieving  = "EloquentRetrieving"
	EventALL         = "EloquentEventALL"
	EventInitialized = "EventInitialized"
	EventBooting     = "EloquentBooting"
	EventBoot        = "EventBoot"
	EventBooted      = "EloquentBooted"

	EventOpened            = "EventOpened"
	EventConnectionCreated = "EventConnectionCreated"

	EventExecuted             = "EventExecuted"
	EventStatementPrepared    = "EventStatementPrepared"
	EventTransactionBegin     = "EventTransactionBegin"
	EventTransactionCommitted = "EventTransactionCommitted"
	EventTransactionRollback  = "EventTransactionRollback"
	EventErrorArised          = "EventErrorArised"
)

type ISaving interface {
	EloquentSaving() error
}
type ISaved interface {
	EloquentSaved() error
}
type ICreating interface {
	EloquentCreating() error
}
type ICreated interface {
	EloquentCreated() error
}
type IUpdating interface {
	EloquentUpdating() error
}
type IUpdated interface {
	EloquentUpdated() error
}
type IDeleting interface {
	EloquentDeleting() error
}
type IDeleted interface {
	EloquentDeleted() error
}
type IRetrieving interface {
	EloquentRetrieving() error
}
type IRetrieved interface {
	EloquentRetrieved() error
}
type IBooting interface {
	EloquentBooting() error
}
type IBooted interface {
	EloquentBooted() error
}
type IInitialized interface {
	EloquentInitialized() error
}
