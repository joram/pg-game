package items

type ItemOldSword struct{}

func (i ItemOldSword) Name() string { return "old sword" }
func (i ItemOldSword) Description() string {
	return "A heavy, time‑worn longsword. Its edge is still sharp enough to hack through thick vines."
}
func (i ItemOldSword) Examine() string {
	return "A heavy, time‑worn longsword. Its edge is still sharp enough to hack through thick vines. or scare off small creatures"
}
