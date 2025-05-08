package items

type LogPile struct{}

func (l LogPile) Examine() string {
	return "The log pile is neatly stacked and ready for use."
}

func (l LogPile) Name() string        { return "log pile" }
func (l LogPile) Description() string { return "A neatly stacked pile of firewood beside the house." }
