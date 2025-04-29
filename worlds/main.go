package worlds

import (
	"github.com/veilstream/psql-text-based-adventure/core/interfaces"
	"github.com/veilstream/psql-text-based-adventure/worlds/simple_example"
)

var Worlds = []interfaces.World{
	simple_example.SimpleExampleWorld,
}
