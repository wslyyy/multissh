package auth

import (
	"errors"
	"multissh/auth/driver"
	"sync"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]driver.GetPassworder)
)

func GetPassword(driverName, ip, user string) (string, error) {
	driversMu.Lock()
	defer driversMu.Unlock()
	d, ok := drivers[driverName]

	if !ok {
		return "", errors.New("unknown password driver: " + driverName)
	}
	return d.GetPassword(ip, user)
}