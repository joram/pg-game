package items

import "fmt"

type ItemBowl struct {
	Full     bool
	Contents string
}

func (b ItemBowl) Name() string {
	if b.Full {
		return "bowl full of " + b.Contents
	}
	return "an empty bowl"
}

func (b ItemBowl) Description() string {
	if b.Full {
		return fmt.Sprintf("Bowl full of %s", b.Contents)
	}

	return fmt.Sprintf("A bowl, it looks like it could hold something. Probably %s.", b.Contents)
}
func (b ItemBowl) Examine() string {
	if b.Full {
		return fmt.Sprintf("A bowl full of %s", b.Contents)
	}
	return "An empty bowl."
}
