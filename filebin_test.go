package todolite

import (
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/couchbaselabs/logg"
)

func TestCopyUrlToFileBin(t *testing.T) {

	sourceUrl := "https://raw.githubusercontent.com/tleyden/todolite-appserver/master/README.md"
	fileBinUrl, err := copyUrlToFileBin(sourceUrl)
	assert.True(t, err == nil)
	assert.False(t, fileBinUrl == sourceUrl)

	assert.True(t, true == true)
	logg.LogTo("TEST", "test finished")

}
