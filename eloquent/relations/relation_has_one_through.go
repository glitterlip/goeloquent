package relations

import "github.com/glitterlip/goeloquent/eloquent"

type HasOneThrough struct {
	eloquent.Relation
	ThroughParent  interface{}
	FarParent      interface{}
	FirstKey       string
	SecondKey      string
	LocalKey       string
	SecondLocalKey string
}
