//go:build e2e
// +build e2e

package require

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/e2e/client"
	"google.golang.org/protobuf/proto"
)

type keyable interface {
	GetKey() string
}

func ContainsKeys[T keyable](t *testing.T, shouldContain bool, appClient *client.TestClient, keyables []T, keys ...string) {
	for _, key := range keys {
		key := appClient.GetKeyStore().GetKey(key)
		for _, msg := range keyables {
			if msg.GetKey() == key {
				if shouldContain {
					return
				}
				require.FailNow(t, fmt.Sprintf("Unexpected key found %q", key))
			}
		}

		if shouldContain {
			require.FailNow(t, fmt.Sprintf("Expected key not found %q", key))
		}
	}
}

func ProtoEqual(t *testing.T, expected proto.Message, actual proto.Message) {
	require.Truef(t, proto.Equal(expected, actual), "expected: %v, actual: %v", expected, actual)
}

func ProtoMapEqual[T proto.Message](t *testing.T, a, b map[string]T) {
	require.Equal(t, len(a), len(b), "Maps have mismatching length")

	for k, va := range a {
		if vb, ok := b[k]; !ok || !proto.Equal(va, vb) {
			require.FailNow(t, fmt.Sprintf("Maps have mismatching content for key %q: %v vs %v", k, va, vb))
		}
	}
}

func ProtoArrayEqual[T proto.Message](t *testing.T, a, b []T) {
	require.Equal(t, len(a), len(b), "Arrays have mismatching length")

	for idx, itemA := range a {
		itemB := b[idx]
		if !proto.Equal(itemA, itemB) {
			require.FailNow(t, fmt.Sprintf("Arrays have mismatching content at index %q: %v vs %v", idx, itemA, itemB))
		}
	}
}
