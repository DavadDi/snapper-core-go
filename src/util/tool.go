package util

// IDToSource ...
func IDToSource(ID string) string {
	switch ID[0:2] {
	case "a.":
		return "android"
	case "i.":
		return "ios"
	}
	return "web"
}
