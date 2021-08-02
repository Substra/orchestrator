package event

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventIndex(t *testing.T) {
	dir, err := ioutil.TempDir("", "indexer-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir) // clean up created directory

	tmpfn := filepath.Join(dir, "tmpfile")

	index, err := NewIndex(tmpfn)
	assert.NoError(t, err)

	lastSeen := index.GetLastEvent("test")
	assert.Zero(t, lastSeen)

	event := &fab.CCEvent{
		BlockNumber: 12,
		TxID:        "a64ed3e8-7336-45d6-a4c8-a92e4c625231",
	}

	err = index.SetLastEvent("chanA", event)
	assert.NoError(t, err)

	assert.Equal(t, uint64(12), index.GetLastEvent("chanA").BlockNum)
	assert.Equal(t, "a64ed3e8-7336-45d6-a4c8-a92e4c625231", index.GetLastEvent("chanA").TxID)

	otherIdx, err := NewIndex(tmpfn)
	assert.NoError(t, err)

	assert.Equal(t, uint64(12), otherIdx.GetLastEvent("chanA").BlockNum)
	assert.Equal(t, "a64ed3e8-7336-45d6-a4c8-a92e4c625231", otherIdx.GetLastEvent("chanA").TxID)
}
