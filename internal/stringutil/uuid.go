package stringutil

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var uuidRand = rand.New(rand.NewSource(time.Now().UnixNano()))
var uuidMutex = &sync.Mutex{}

// UUID generates a random UUID according to RFC 4122.
func UUID() string {
	uuid := make([]byte, 16)
	uuidMutex.Lock()
	_, _ = uuidRand.Read(uuid)
	uuidMutex.Unlock()
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
