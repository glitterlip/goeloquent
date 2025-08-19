package goeloquent

type EventName string

const (
	EventConnectionOpened      EventName = "connection:opened"
	EventQueryExecuted         EventName = "query:executed"
	EventTransactionBegin      EventName = "transaction:begin"
	EventTransactionCommitting EventName = "transaction:committing"
	EventTransactionCommitted  EventName = "transaction:committed"
	EventTransactionRolledback EventName = "transaction:rolledback"
	EventError                 EventName = "error"
)

type ModelEventName string

const (
	EventModelBooting       ModelEventName = "model:booting"
	EventModelBoot          ModelEventName = "model:boot"
	EventModelBooted        ModelEventName = "model:booted"
	EventModelCreating      ModelEventName = "model:creating"
	EventModelCreated       ModelEventName = "model:created"
	EventModelSaving        ModelEventName = "model:saving"
	EventModelSaved         ModelEventName = "model:saved"
	EventModelUpdating      ModelEventName = "model:updating"
	EventModelUpdated       ModelEventName = "model:updated"
	EventModelDeleting      ModelEventName = "model:deleting"
	EventModelDeleted       ModelEventName = "model:deleted"
	EventModelTrashed       ModelEventName = "model:trashed"
	EventModelForceDeleting ModelEventName = "model:force_deleting"
	EventModelForceDeleted  ModelEventName = "model:force_deleted"
	EventModelRestoring     ModelEventName = "model:restoring"
	EventModelRestored      ModelEventName = "model:restored"
	EventModelRetrieved     ModelEventName = "model:retrieved"
	EventModelReplicating   ModelEventName = "model:replicating"
)
