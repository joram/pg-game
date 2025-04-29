package actions

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/core/verbs"
)

type Action interface {
	Execute(backend *pgproto3.Backend, world *interfaces.World)
}

type ListCommandsAction struct {
}

func (h *ListCommandsAction) Execute(backend *pgproto3.Backend, world *interfaces.World) {
	introLines := []string{
		" ",
		"***********************************************",
		"Welcome to your very own Text-Based Adventure!",
		"***********************************************",
	}
	for _, verb := range verbs.AllVerbs {
		line := fmt.Sprintf(" - \"%s\" ", verb.Name)
		for {
			if len(line) > 14 {
				break
			}
			line += " "
		}
		line += verb.Description
		introLines = append(introLines, line)
	}
	introLines = append(introLines, " ")

	for _, line := range introLines {
		backend.Send(&pgproto3.NoticeResponse{
			Severity: "",
			Message:  line,
		})
	}
}
