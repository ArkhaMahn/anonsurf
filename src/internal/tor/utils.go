package tor

import (
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
)

func IsAnonsurfRunning() bool {
	controller := systemd.NewAnonsurfdController()
	active, err := controller.IsActive()
	return err == nil && active
}
