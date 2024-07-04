package goeloquent

type HasOneThrough struct {
	Relation
	ThroughParent  interface{}
	FarParent      interface{}
	FirstKey       string
	SecondKey      string
	LocalKey       string
	SecondLocalKey string
}
