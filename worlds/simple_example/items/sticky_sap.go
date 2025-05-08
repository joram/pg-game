package items

type StickySap struct{}

func (s StickySap) Examine() string {
	return "A thick, sticky sap that could glue things together."
}

func (s StickySap) Name() string { return "sticky sap" }
func (s StickySap) Description() string {
	return "A thick, sticky sap that could glue things together."
}
