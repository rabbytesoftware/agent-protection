package interfacesize

type catalog struct {
	interfaces map[string]declaration
	structures map[string]structure
	methods    map[string][]string
	functions  map[string]bool
}
