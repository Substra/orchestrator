package event

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// Indexer keeps track of last processed events per channel.
type Indexer interface {
	GetLastEvent(channel string) IndexedEvent
	SetLastEvent(channel string, event *fab.CCEvent) error
}

type IndexedEvent struct {
	BlockNum   uint64
	TxID       string
	IsIncluded bool // whether the bound must be included when listing events from this event
}

// Index implements Indexer by writing events in a file.
type Index struct {
	lock     *sync.RWMutex
	filepath string
	events   map[string]IndexedEvent
}

func NewIndex(filepath string) (*Index, error) {
	eventIndex := &Index{
		lock:     new(sync.RWMutex),
		events:   make(map[string]IndexedEvent),
		filepath: filepath,
	}

	data, err := ioutil.ReadFile(filepath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err == nil {
		// file exists
		err = json.Unmarshal(data, &eventIndex.events)
		if err != nil {
			return nil, err
		}
	}

	return eventIndex, nil
}

func (ei *Index) GetLastEvent(channel string) IndexedEvent {
	ei.lock.RLock()
	defer ei.lock.RUnlock()
	return ei.events[channel]
}

func (ei *Index) SetLastEvent(channel string, e *fab.CCEvent) error {
	ei.lock.Lock()
	defer ei.lock.Unlock()
	ei.events[channel] = IndexedEvent{
		BlockNum: e.BlockNumber,
		TxID:     e.TxID,
	}

	data, err := json.Marshal(ei.events)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(ei.filepath, data, 0600)
}
