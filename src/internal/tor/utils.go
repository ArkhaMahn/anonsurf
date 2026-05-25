package tor

import (
	"os"
)

func IsAnonsurfRunning() bool {
	_, err := os.Lstat("/run/systemd/units/invocation:anonsurfd.service")
	return err == nil
}
