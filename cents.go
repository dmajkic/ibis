package ibis

// Cent is an interface that represents one ibis cent
// ibis Cent is a plugin interface to application
type Cent interface {
	Init(server *Server)
}

var cents = map[string]Cent{}

// RegisterCent should be used from cent package implementation
func RegisterCent(name string, cent Cent) {
	cents[name] = cent
}
