package mybench

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateUniqueStringFromInteger(t *testing.T) {
	output := make(map[string]struct{})
	for i := int64(0); i < 1000000; i++ {
		value := generateUniqueStringFromInt(i, 20)
		_, found := output[value]
		require.Equal(t, found, false, fmt.Sprintf("found duplicate values for integer %d with value %s", i, value))
		output[value] = struct{}{}
	}
}
