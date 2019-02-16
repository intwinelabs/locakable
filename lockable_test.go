package lockable

import (
	"testing"

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

	if isLocked, _ := l.Lock(l.Id.String(), 3000); !isLocked {
		assert.FailNow("failure locking")
	}
	defer l.Unlock(l.Id.String())

	l.Name = "Fred"

	go func() {
		if isLocked, _ := l.Lock(l.Id.String(), 3000); !isLocked {
			return
		}
		l.Name = "Bob"
	}()

	assert.Equal("Fred", l.Name)
}
