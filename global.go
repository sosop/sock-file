package sockfile

import "os"

const (
	DEFAULT_DIR = "/data"
)

var (
	isTrans      = false
	otherServers = os.Getenv("OTHER_SERVERS")
)
