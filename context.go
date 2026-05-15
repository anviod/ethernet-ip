package ethernet_ip

import (
	"math/rand"
	"sync"
	"time"

	"github.com/anviod/ethernet-ip/types"
)

var (
	randMu  sync.Mutex
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func contextGenerator() types.ULInt {
	randMu.Lock()
	defer randMu.Unlock()
	return types.ULInt(randGen.Int63())
}
