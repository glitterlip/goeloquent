package goeloquent

import (
	"fmt"
	"reflect"
)

type Relations string

func GetMorphMap(name string) string {
	v, ok := RegisteredMorphModelsMap.Load(name)
	if !ok {
		panic(fmt.Sprintf("no registered morph model found for %s,check your code or register it", name))
	}
	return v.(string)
}

const (
	RelationHasOne    Relations = "HasOne"
	RelationBelongsTo Relations = "BelongsTo"

	RelationHasMany Relations = "HasMany"

	RelationHasOneThrough  Relations = "HasOneThrough"
	RelationHasManyThrough Relations = "HasManyThrough"

	RelationBelongsToMany Relations = "BelongsToMany"

	RelationMorphOne  Relations = "MorphOne"
	RelationMorphMany Relations = "MorphMany"
	RelationMorphTo   Relations = "MorphTo"

	RelationMorphToMany Relations = "MorphToMany"
	RelationMorphByMany Relations = "MorphByMany"
	PivotAlias                    = "goelo_pivot_"     // used for pivot alias
	OrmPivotAlias                 = "goelo_orm_pivot_" // used for orm pivot alias
)

type Relation struct {
	SelfModel        interface{} // parent/self model pointer,usually a pointer to a struct , &User{}
	RelatedModel     interface{} // related model,usually a pointer to a struct , &User{}
	RelationTypeName Relations   // relation type name
	FieldName        string      // field name in self model corresponding to this relation
	*EloquentBuilder             // RelationBuilder

}

func (r *Relation) GetEloquentBuilder() *EloquentBuilder {
	return r.EloquentBuilder
}

type RelationI interface {
	AddEagerConstraints(models interface{})
	AddConstraints()
	GetEloquentBuilder() *EloquentBuilder

	//Match(models []interface{}, results []interface{}, relation string) //Match the eagerly loaded results to their parents.
	//GetResults() //Get the results of the relationship.
	//GetEager() //Get the relationship for eager loading.
	//Get(dest interface{}, columns ...interface{}) //Get the results of the relationship.

	//GetRelationExistenceQuery(Builder $query, Builder $parentQuery, columns = ['*']) //Get the relationship for exists query.
	//GetRelationExistenceCountQuery(Builder $query, Builder $parentQuery) //Get the relationship count of an eager load query.

	//GetKeys(models []interface{},key string) //Get all the primary keys for an array of models.
	//GetQuery() *Builder //Get the underlying query for the relation.
	//GetQualifiedParentKeyName()string //
}

func NewRelationBaseBuilder(related interface{}) *EloquentBuilder {

	if related == nil {
		return ToEloquentBuilder(NewQueryBuilder())
	}
	return NewEloquentBuilder(related)
}
func (r *Relation) GetSelfKey(key string) interface{} {
	feild := GetParsedModel(r.SelfModel).FieldsByDbName[key]
	return reflect.ValueOf(r.SelfModel).Elem().FieldByName(feild.Name).Interface()
}

/*
HasMany Define a one-to-many relationship.

if we have a user model and a post model,each user has many posts, user has a hasMany relation with post, the user(self) model has an id column as primary key, and the post model(related) has a user_id column,

post's user_id = user's id , so the relation is defined as follows:

	func (u *User) PostRelation() *goeloquent.HasManyRelation {
		return u.HasMany(u, &Post{}, "id", "user_id")
	}
*/
func (m *EloquentModel) HasMany(selfModelPointer, relatedModelPointer interface{}, selfColumn, relatedColumn string) *HasManyRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := HasManyRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationHasMany,
			EloquentBuilder:  b,
		},
		SelfColumn:    selfColumn,
		RelatedColumn: relatedColumn,
	}

	relation.AddConstraints()
	return &relation
}

/*
BelongsTo Define an inverse one-to-one or many relationship.

if we have a user model and a post model,each post belongs to a user,user has a belongsTo relation with post,

the user(related) model has an id column as primary key, and the post model(self) has a user_id column,

post's user_id = user's id , so the relation is defined as follows:

	func (p *Post) UserRelation() *goeloquent.BelongsToRelation {
		return p.BelongsTo(p, &User{}, "user_id", "id")
	}
*/
func (m *EloquentModel) BelongsTo(selfModelPointer, relatedModelPointer interface{}, selfColumn, relatedColumn string) *BelongsToRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := BelongsToRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationBelongsTo,
			EloquentBuilder:  b,
		},
		SelfColumn:    selfColumn,
		RelatedColumn: relatedColumn,
	}
	relation.AddConstraints()

	return &relation

}

/*
BelongsToMany Define a many-to-many relationship.

if we have a user model and a role model,each user have many roles,each role also belongs to many users,user(related) has a belongsToMany relation with role(self),

in this case, we need a pivot table to store the relationship between user and role, the pivot table(let's say user_roles) has user_id and role_id columns,

user_roles.user_id = users.uid, user_roles.role_id = roles.rid, so the relation is defined as follows:

	func (r *Role) UsersRelation() *goeloquent.BelongsToManyRelation {
		return r.BelongsToMany(r, &User{}, "role_users", "role_id", "user_id", "rid", "uid")
	}

usually we use id as priamry key for both users,roles table, here we use uid,rid as primary key for users,roles table only for demonstration purpose.
*/
func (m *EloquentModel) BelongsToMany(selfModelPointer, relatedModelPointer interface{}, pivotTable, pivotSelfColumn, pivotRelatedColumn, selfModelColumn, relatedModelColumn string) *BelongsToManyRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := BelongsToManyRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationBelongsToMany,
			EloquentBuilder:  b,
		},
		PivotTable:         pivotTable,
		PivotSelfColumn:    pivotSelfColumn,
		PivotRelatedColumn: pivotRelatedColumn,
		SelfColumn:         selfModelColumn,
		RelatedColumn:      relatedModelColumn,
	}
	relatedModel := GetParsedModel(relatedModelPointer)
	b.Join(relation.PivotTable, relation.PivotTable+"."+relation.PivotRelatedColumn, "=", relatedModel.Table+"."+relation.RelatedColumn)
	b.Select(relatedModel.Table + "." + "*")
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotSelfColumn, OrmPivotAlias, relation.PivotSelfColumn))
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotRelatedColumn, OrmPivotAlias, relation.PivotRelatedColumn))
	relation.AddConstraints()
	return &relation

}

/*
HasOne Define a one-to-one relationship.

let's say we have a User model and a Phone model,each user has a phone , user has a hasOne relation with phone,

the user model(self) has an id column as primary key, and the phone model(related) has a user_id column,

phone's user_id = user's id , so the relation is defined as follows:

	func (u *User) PhoneRelation() *goeloquent.HasOneRelation {
		return u.HasOne(u, &Phone{}, "id", "user_id")
	}
*/
func (m *EloquentModel) HasOne(selfModelPointer, relatedModelPointer interface{}, selfColumn, relatedColumn string) *HasOneRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := HasOneRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationHasOne,
			EloquentBuilder:  b,
		},
		RelatedColumn: relatedColumn,
		SelfColumn:    selfColumn,
	}
	relation.AddConstraints()
	return &relation
}

/*
MorphTo Create a new morph to relationship instance.

let's say we have a post model , a video model and a comment model,each post/video may have many comments,

and each comment belongs to a post or a video, so a comment have a morphTo relation with post/video model ,

we use comments table to store the comments ,in this case we would need a column(commentable_type) in comments table to
decide the type of the related model(post/video) and another column(commentable_id) to store the id of the related model(post/video),

	func (c *Comment) CommentableRelation() *goeloquent.MorphToRelation {
		return c.MorphTo(c, "commentable_id", "commentable_type", "id")
	}
*/
func (m *EloquentModel) MorphTo(selfModelPointer interface{}, selfRelatedIdColumn, selfRelatedTypeColumn, relatedModelIdColumn string) *MorphToRelation {

	builder := NewRelationBaseBuilder(nil)
	relation := MorphToRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     nil,
			RelationTypeName: RelationMorphTo,
			EloquentBuilder:  builder,
		},
		SelfRelatedIdColumn:   selfRelatedIdColumn,
		RelatedModelIdColumn:  relatedModelIdColumn,
		SelfRelatedTypeColumn: selfRelatedTypeColumn,
	}

	relation.AddConstraints()

	return &relation

}

/*
MorphOne Define a polymorphic one-to-one relationship.

let's say we have a user model and an image model(save all other models related images ,post images , video screenshots,user avatars),each user has an avatar image, user has a morphOne relation with image,

the user model(self) has an id column as primary key, and the image model(related) has an imageable_id column and an imageable_type column,

image's imageable_id = user's id , image's imageable_type = user's model name, so the relation is defined as follows:

	func (u *User) AvatarRelation() *goeloquent.MorphOneRelation {
		return u.MorphOne(u, &Image{}, "id", "imageable_id", "imageable_type", "user")
	}
*/
func (m *EloquentModel) MorphOne(selfModelPointer, relatedModelPointer interface{}, selfColumn, relatedModelIdColumn, relatedModelTypeColumn string, morphType ...string) *MorphOneRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := MorphOneRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationMorphOne,
			EloquentBuilder:  b,
		},
		RelatedModelIdColumn:   relatedModelIdColumn,
		SelfColumn:             selfColumn,
		RelatedModelTypeColumn: relatedModelTypeColumn,
	}
	selfModel := GetParsedModel(selfModelPointer)
	if len(morphType) > 0 {
		relation.RelatedModelTypeColumnValue = morphType[0]
	} else {
		relation.RelatedModelTypeColumnValue = GetMorphMap(selfModel.Name)
	}
	relation.AddConstraints()

	return &relation

}

/*
MorphMany Define a polymorphic one-to-many relationship.

let's say we have a post model and an image model,each post has many images, post has a morphMany relation with image,

the post model(self) has an id column as primary key, and the image model(related) has an imageable_id column and an imageable_type column,

image's imageable_id = post's id , image's imageable_type = post's model name, so the relation is defined as follows:

	func (p *Post) ImageRelation() *goeloquent.MorphManyRelation {
		return p.MorphMany(p, &Image{}, "id", "imageable_id", "imageable_type")
	}

	func (u *User) ImageRelation() *goeloquent.MorphManyRelation {
		return u.MorphMany(u, &Image{}, "id", "imageable_id", "imageable_type")
	}
*/
func (m *EloquentModel) MorphMany(selfModelPointer, relatedModelPointer interface{}, selfColumn, relatedModelIdColumn, relatedModelTypeColumn string, relatedModelTypeColumnValue ...string) *MorphManyRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := MorphManyRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationMorphOne,
			EloquentBuilder:  b,
		},
		RelatedModelIdColumn:   relatedModelIdColumn,
		SelfColumn:             selfColumn,
		RelatedModelTypeColumn: relatedModelTypeColumn,
	}
	selfModel := GetParsedModel(selfModelPointer)

	if len(relatedModelTypeColumnValue) > 0 {
		relation.RelatedModelTypeColumnValue = relatedModelTypeColumnValue[0]
	} else {
		relation.RelatedModelTypeColumnValue = GetMorphMap(selfModel.Name)
	}

	relation.AddConstraints()

	return &relation

}

/*
MorphToMany Define a polymorphic many-to-many relationship.

let's say we have a video model and a tag model,each video has many tags,each tag also belongs to many videos,video has a morphToMany relation with tag,

the video model(self) has an id column as primary key, and the tag model(related) has a id column, we need a pivot table to store the relationship between video and tag,

the pivot table(let's say tagables) has tag_id , tagable_id , tagable_type columns(to decide the type of the related model(post/video))

tagables.tag_id = tags.id, tagables.tagable_id = videos.id/posts.id,tagable_type = post/video, so the relation is defined as follows:

for post model

	func (p *Post) TagsRelation() *goeloquent.MorphToManyRelation {
		return p.MorphToMany(p, &Tag{}, "id", "id", "tagables", "tagable_id", "tagable_type", "tag_id", "post")
	}

for video model

	func (v *Video) TagsRelation() *goeloquent.MorphToManyRelation {
		return v.MorphToMany(v, "tagable_id", "tagable_type", "tagables", "tag_id", "tag_id", "id", "id")
	}
*/
func (m *EloquentModel) MorphToMany(selfModelPointer, relatedModelPointer interface{}, selfIdColumn, relatedIdColumn, pivotTable, pivotSelfIdColumn, pivotSelfTypeColumn, pivotRelatedIdColumn string, morphType ...string) *MorphToManyRelation {

	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := MorphToManyRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationMorphToMany,
			EloquentBuilder:  b,
		},
		PivotTable:           pivotTable,
		PivotSelfIdColumn:    pivotSelfIdColumn,
		PivotSelfTypeColumn:  pivotSelfTypeColumn,
		SelfIdColumn:         selfIdColumn,
		RelatedIdColumn:      relatedIdColumn,
		PivotRelatedIdColumn: pivotRelatedIdColumn,
	}
	relatedModel := GetParsedModel(relatedModelPointer)
	selfModel := GetParsedModel(selfModelPointer)
	b.Join(relation.PivotTable, relation.PivotTable+"."+relation.PivotRelatedIdColumn, "=", relatedModel.Table+"."+relation.RelatedIdColumn)
	b.Select(relatedModel.Table + "." + "*")
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotRelatedIdColumn, OrmPivotAlias, relation.PivotRelatedIdColumn))
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotSelfIdColumn, OrmPivotAlias, relation.PivotSelfIdColumn))
	var selfModelTypeColumnValue string

	if len(morphType) > 0 {
		selfModelTypeColumnValue = morphType[0]
	} else {
		selfModelTypeColumnValue = GetMorphMap(selfModel.Name)
	}
	relation.SelfModelTypeColumnValue = selfModelTypeColumnValue
	relation.AddConstraints()
	return &relation

}

/*
MorphByMany Define a polymorphic many-to-many relationship.

let's say we have a tag model and a post model,each tag has many posts,each post also belongs to many tags,tag has a morphByMany relation with post,

the tag model(self) has an id column as primary key, and the post model(related) has a id column, we need a pivot table to store the relationship between tag and post,

the pivot table(let's say tagables) has tag_id , tagable_id , tagable_type columns(to decide the type of the related model(post/video))

tagables.tag_id = tags.id, tagables.tagable_id = videos.id/posts.id,tagable_type = post/video, so the relation is defined as follows:

	func (t *Tag) PostsRelation() *goeloquent.MorphByManyRelation {
		return t.MorphByMany(t, &Post{}, "tagables", "id", "id", "tag_id", "tagable_id", "tagable_type")
	}

	func (t *Tag) VideosRelation() *goeloquent.MorphByManyRelation {
		return t.MorphByMany(t, &Video{}, "tagables", "id", "id", "tag_id", "tagable_id", "tagable_type")
	}
*/
func (m *EloquentModel) MorphByMany(selfModelPointer, relatedModelPointer interface{}, pivotTable, selfColumn, relatedIdColumn, pivotSelfColumn, pivotRelatedIdColumn, pivotRelatedTypeColumn string) *MorphByManyRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := MorphByManyRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationMorphByMany,
			EloquentBuilder:  b,
		},
		PivotTable:             pivotTable,
		PivotSelfColumn:        pivotSelfColumn,
		PivotRelatedIdColumn:   pivotRelatedIdColumn,
		SelfColumn:             selfColumn,
		RelatedIdColumn:        relatedIdColumn,
		PivotRelatedTypeColumn: pivotRelatedTypeColumn,
	}
	selfModel := GetParsedModel(selfModelPointer)
	relatedModel := GetParsedModel(relatedModelPointer)
	modelMorphName := GetMorphMap(selfModel.Name)

	b.Join(relation.PivotTable, relation.PivotTable+"."+relation.PivotRelatedIdColumn, "=", relatedModel.Table+"."+relation.RelatedIdColumn)
	b.Select(relatedModel.Table + "." + "*")
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotRelatedIdColumn, OrmPivotAlias, relation.PivotRelatedIdColumn))
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.SelfColumn, OrmPivotAlias, relation.PivotSelfColumn))
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.RelationTypeName, OrmPivotAlias, relation.RelationTypeName))

	relation.RelatedModelTypeColumnValue = modelMorphName
	relation.AddConstraints()
	return &relation

}
