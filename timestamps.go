package goeloquent

type HasTimestamps interface {
	//Update the model's update timestamp
	Touch() bool
	//
	SetCreatedAt() Builder
	SetUpdatedAt() Builder
	GetCreatedAtColumn() Field
	GetUpdatedAtColumn() Field
}
