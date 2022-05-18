package scenarios

import (
	"github.com/go-playground/log/v7"
	"github.com/golang/protobuf/proto"
	"github.com/owkin/orchestrator/e2e/client"
)

type keyable interface {
	GetKey() string
}

func assertContainsKeys[T keyable](shouldContain bool, appClient *client.TestClient, keyables []T, keys ...string) {
	for _, key := range keys {
		key := appClient.GetKeyStore().GetKey(key)
		for _, msg := range keyables {
			if msg.GetKey() == key {
				if shouldContain {
					return
				}
				log.Fatalf("Unexpected key found %q", key)
			}
		}
		if shouldContain {
			log.Fatalf("Expected key not found %q", key)
		}
	}
}

func assertProtoMapEqual[T proto.Message](a, b map[string]T) {
	if len(a) != len(b) {
		log.Fatalf("Maps have mismatching length: %d vs %d", len(a), len(b))
	}
	for k, va := range a {
		if vb, ok := b[k]; !ok || !proto.Equal(va, vb) {
			log.Fatalf("Maps have mismatching content for key %q: %s vs %s", k, va, vb)
		}
	}
}

func assertProtoArrayEqual[T proto.Message](a, b []T) {
	if len(a) != len(b) {
		log.Fatalf("Arrays have mismatching length: %d vs %d", len(a), len(b))
	}
	for idx, itemA := range a {
		itemB := b[idx]
		if !proto.Equal(itemA, itemB) {
			log.Fatalf("Arrays have mismatching content at index %q: %s vs %s", idx, itemA, itemB)
		}
	}
}
