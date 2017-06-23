package collections

import (
	"bytes"
	"encoding/gob"

	"github.com/boltdb/bolt"
	"github.com/splitio/go-agent/log"
	"github.com/splitio/go-agent/splitio/storage/boltdb"
)

const segmentChangesCollectionName = "SEGMENT_CHANGES_COLLECTION"

// NewSegmentChangesCollection returns an instance of SegmentChangesCollection
func NewSegmentChangesCollection(dbb *bolt.DB) SegmentChangesCollection {
	baseCollection := boltdb.Collection{DB: dbb, Name: segmentChangesCollectionName}
	var sCollection = SegmentChangesCollection{Collection: baseCollection}
	return sCollection
}

// SegmentChangesItem represents an SplitChanges service response
type SegmentChangesItem struct {
	Name string
	Keys map[string]struct{}
}

// SegmentChangesCollection represents a collection of SplitChangesItem
type SegmentChangesCollection struct {
	boltdb.Collection
}

// Add an item
func (c SegmentChangesCollection) Add(item *SegmentChangesItem) error {
	key := []byte(item.Name)
	err := c.Collection.SaveAs(key, item)
	return err
}

// Fetch return a SegmentChangesItem
func (c SegmentChangesCollection) Fetch(name string) (*SegmentChangesItem, error) {
	key := []byte(name)
	item, err := c.Collection.FetchBy(key)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, nil
	}

	var decodeBuffer bytes.Buffer
	decodeBuffer.Write(item)
	dec := gob.NewDecoder(&decodeBuffer)

	var q SegmentChangesItem
	errq := dec.Decode(&q)
	if errq != nil {
		log.Error.Println("decode error:", errq)
	}
	return &q, nil
}
