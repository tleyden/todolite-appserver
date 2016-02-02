package libtodolite

import (
	"testing"

	"github.com/couchbaselabs/go.assert"
)

func TestLastProcessedSeq(t *testing.T) {
	todolite := TodoLiteApp{}
	lastSeq := "1"
	err := todolite.saveLastProcessedSeq(lastSeq)
	assert.True(t, err == nil)
	fetchedLastSeq, err := todolite.lastProcessedSeq()
	assert.True(t, err == nil)
	assert.Equals(t, fetchedLastSeq, lastSeq)
}
