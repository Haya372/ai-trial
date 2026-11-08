package tag

import (
	"github.com/google/uuid"
)

// MaxTagsPerUser is the number of tags one user may own (SPEC-008).
const MaxTagsPerUser = 50

type Tag interface {
	ID() uuid.UUID
	UserID() uuid.UUID
	Name() TagName
	Color() TagColor
}

type tagEntity struct {
	id     uuid.UUID
	userID uuid.UUID
	name   TagName
	color  TagColor
}

func New(id, userID uuid.UUID, name TagName, color TagColor) Tag {
	return &tagEntity{id: id, userID: userID, name: name, color: color}
}

func (t *tagEntity) ID() uuid.UUID     { return t.id }
func (t *tagEntity) UserID() uuid.UUID { return t.userID }
func (t *tagEntity) Name() TagName     { return t.name }
func (t *tagEntity) Color() TagColor   { return t.color }
