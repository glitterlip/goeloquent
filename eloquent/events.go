package eloquent

const (
	EventSaving     = "Saving"
	EventSaved      = "Saved"
	EventCreating   = "Creating"
	EventCreated    = "Created"
	EventUpdating   = "Updating"
	EventUpdated    = "Updated"
	EventDeleteing  = "Deleting"
	EventDeleted    = "Deleted"
	EventRetrieved  = "Retrieved" //TODO: EventRetrieved EventRetrieving
	EventRetrieving = "Retrieving"
	EventALL        = "ALL"
	EventBooting    = "Booting"
	EventBooted     = "Booted"
	TagName         = "goelo"
	EloquentName    = "EloquentModel"
)

type ISaving interface {
	Saving(model interface{}) error
}
type ISaved interface {
	Saved(model interface{}) error
}
type ICreating interface {
	Creating(model interface{}) error
}
type ICreated interface {
	Created(model interface{}) error
}
type IUpdating interface {
	Updating(model interface{}) error
}
type IUpdated interface {
	Updated(model interface{}) error
}
type IDeleting interface {
	Deleting(model interface{}) error
}
type IDeleted interface {
	Deleted(model interface{}) error
}
type IRetrieved interface {
	Retrieved(model interface{}) error
}
