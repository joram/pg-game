package items

type GlowingMushroom struct{}

func (g GlowingMushroom) Name() string        { return "glowing mushroom" }
func (g GlowingMushroom) Description() string { return "It glows with a soft, magical light." }
func (g GlowingMushroom) Examine() string {
	return "You look closely at the glowing mushroom. It pulses gently like it’s alive."
}
