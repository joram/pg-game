package items

type Screwdriver struct{}

func (s Screwdriver) Examine() string {
	return "The screwdriver is a bit rusty, it's a flathead, and it still seems to works."
}

func (s Screwdriver) Name() string        { return "screwdriver" }
func (s Screwdriver) Description() string { return "A flathead screwdriver, a bit rusty but usable." }
