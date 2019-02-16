package lockable

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mediocregopher/radix"
	"github.com/stretchr/testify/assert"
)

type lockableTest struct {
	Id   uuid.UUID
	Name string
	lockable
}

func TestLockable(t *testing.T) {
	assert := assert.New(t)

	l := lockableTest{
		Id: uuid.New(),
	}

	client, _ := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	l.SetCache(client)

	if isLocked, _ := l.Lock(l.Id.String(), 30); !isLocked {
		// return with error in a function
		assert.FailNow("failure locking")
	}
	defer l.Unlock(l.Id.String())

	l.Name = "Fred"

	go func() {
		if isLocked, _ := l.Lock(l.Id.String(), 30); !isLocked {
			return
		}
		l.Name = "Bob"
	}()

	assert.Equal("Fred", l.Name)

	time.Sleep(30 * time.Millisecond)

	err := l.Unlock(l.Id.String())
	_err := errors.New("Unlock failed, key incorrect or lock timedout")
	assert.Equal(_err, err)
}
