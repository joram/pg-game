package items

type ItemLantern struct {
	Attached bool
}

func (i ItemLantern) Examine() string {
	return "The lantern is small, and it is casting a small radius of light around you."
}

func (i ItemLantern) Name() string {
	return "lantern"
}

func (i ItemLantern) Description() string {
	if i.Attached {
		return "A small lantern, it is hanging on the wall beside the door. It is casting a small radius of light around you."
	}
	return "a small lantern"
}
