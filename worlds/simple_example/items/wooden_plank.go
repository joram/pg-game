package items

type WoodenPlank struct{}

func (p WoodenPlank) Examine() string {
	return "A long, flat plank — perfect for patching something."
}

func (p WoodenPlank) Name() string { return "wooden plank" }
func (p WoodenPlank) Description() string {
	return "A long, flat plank — perfect for patching something."
}
