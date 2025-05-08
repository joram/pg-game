package actions

import (
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/veilstream/psql-text-based-adventure/core/world"
)

type LookAction struct{}

func (l *LookAction) Execute(backend *pgproto3.Backend, world *world.World) {
	loc := *world.CurrentLocation
	description := loc.Describe()
	backend.Send(&pgproto3.NoticeResponse{
		Severity: "",
		Message:  description,
	})
}
