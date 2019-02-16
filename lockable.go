package lockable

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/mediocregopher/radix"
)

var (
	lockScript = radix.NewEvalScript(1, `return redis.call('SET', KEYS[1], ARGV[1], 'NX', 'PX', ARGV[2])`)

	unlockScript = radix.NewEvalScript(1, `
		if redis.call("get",KEYS[1]) == ARGV[1] then
		    return redis.call("del",KEYS[1])
		else
		    return 0
		end
	`)
)

type lockable struct {
	client  *radix.Pool
	lockVal uuid.UUID
}

// SetCache sets the radix client to be used for locking
func (l *lockable) SetCache(c *radix.Pool) {
	l.client = c
}

// Lock attempts to put a lock on the key for a specified duration (in milliseconds).
// If the lock was successfully acquired, true will be returned.
func (l *lockable) Lock(key string, timeout int) (bool, error) {
	var res string
	if l.lockVal == uuid.Nil {
		l.lockVal = uuid.New()
	}
	t := fmt.Sprintf("%d", timeout)
	if err := l.client.Do(lockScript.Cmd(&res, key, l.lockVal.String(), t)); err != nil {
		return false, nil
	} else {
		return res == "OK", nil
	}
}

// Unlock attempts to remove the lock on a key .
// If the lock cannot be removed, because the key has already expired an error will be returned.
func (l *lockable) Unlock(key string) error {
	var res int
	if err := l.client.Do(unlockScript.Cmd(&res, key, l.lockVal.String())); err != nil {
		return err
	} else if res != 1 {
		return errors.New("Unlock failed, key incorrect or lock timedout")
	}

	// Success
	return nil
}
