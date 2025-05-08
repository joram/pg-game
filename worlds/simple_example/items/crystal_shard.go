package items

type CrystalShard struct{}

func (c CrystalShard) Name() string        { return "Crystal Shard" }
func (c CrystalShard) Description() string { return "A shiny crystal shard that sparkles faintly." }
func (c CrystalShard) Examine() string {
	return "The crystal shard sparkles faintly, it seems to be a piece of a larger crystal."
}
