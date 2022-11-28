package mybench

import (
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

var randomSeed int64

func init() {
	randomSeed = time.Now().UnixMicro()
	logrus.WithField("seed", randomSeed).Info("test generated new random seed")

	// Seed the global random, but really each random source should seed with the above.
	rand.Seed(randomSeed)
}

func newRandForTest() *rand.Rand {
	return rand.New(rand.NewSource(randomSeed))
}
