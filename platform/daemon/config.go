package daemon

var (
	address string = "127.0.0.1"
	port    uint16 = 8080
)

func Address() string {
	return address
}

func Port() uint16 {
	return port
}

func SetAddress(addr string) {
	address = addr
}

func SetPort(newPort uint16) {
	port = newPort
}
