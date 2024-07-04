package goeloquent

type HasManyThrough struct {
	Relation
	ThroughParent  interface{}
	FarParent      interface{}
	FirstKey       string
	SecondKey      string
	LocalKey       string
	SecondLocalKey string
}
