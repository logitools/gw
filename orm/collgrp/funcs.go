package collgrp

import (
	"github.com/logitools/gw/model"
	"github.com/logitools/gw/orm/coll"
)

// GroupBy creates an Unordered CollectionGroup from a Collection
// You can give it a group order by Sort()
func GroupBy[
	MP model.Identifiable[ID],
	ID comparable,
	K comparable,
](
	srcColl *coll.Collection[MP, ID],
	keyFn func(MP) K,
) *CollectionGroup[MP, ID, K] {

	if srcColl == nil {
		return nil
	}

	g := NewEmptyCollectionGroup[MP, ID, K]()

	var subCollGen func() *coll.Collection[MP, ID]
	// respect the item-order of the source collection
	if srcColl.IsOrdered() {
		// source collection is ordered
		subCollGen = coll.NewEmptyOrderedCollection
	} else {
		subCollGen = coll.NewEmptyUnorderedCollection
	}

	// get or create subgroup with ordered behavior
	getOrCreateSubCollection := func(k K) *coll.Collection[MP, ID] {
		if c, ok := g.FindCollection(k); ok {
			return c
		}
		c := subCollGen()
		g.SetCollection(k, c)
		return c
	}

	// Respect the source's natural iteration order (ordered or unordered)
	srcColl.ForEach(func(mp MP) {
		k := keyFn(mp)
		getOrCreateSubCollection(k).AddIfNew(mp)
	})

	return g
}
