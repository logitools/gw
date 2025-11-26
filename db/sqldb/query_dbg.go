//go:build debug && verbose

package sqldb

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/logitools/gw/model"
	"github.com/logitools/gw/orm/coll"
)

// LoadBelongsTo - Load Parents on Children from SQL DB and Link Child-BelongsTo-Parent Relation
// Returns the Parents
func LoadBelongsTo[
	CP model.Identifiable[CID],
	CID comparable,
	P any, // Model struct
	PP ScannableIdentifiable[P, PID],
	PID comparable,
](
	ctx context.Context,
	dbClient Client,
	children *coll.Collection[CP, CID],
	sqlSelectBase string,
	foreignKey func(c CP) PID,
	relationFieldPtr func(c CP) *PP,
) (*coll.Collection[PP, PID], error) {
	fKeysAsAny := coll.CollectUniqueToSlice(children, func(c CP) any { return foreignKey(c) })
	parts := make([]string, len(fKeysAsAny))
	for i, v := range fKeysAsAny {
		parts[i] = fmt.Sprint(v) // fmt.Sprint converts any value to string e.g. 3->"3", true->"true", nil->"<nil>"
	}
	log.Printf("[DEBUG] LoadBelongsTo() FKs: %s", strings.Join(parts, ","))
	sqlStmt := sqlSelectBase + fmt.Sprintf(" WHERE id IN (%s)", dbClient.Placeholders(len(fKeysAsAny)))
	log.Printf("[DEBUG] LoadBelongsTo() sqlStmt %s", sqlStmt)
	parents, err := RawQueryCollection[P, PP, PID](ctx, dbClient, sqlStmt, fKeysAsAny...)
	if err != nil {
		return nil, err
	}
	err = coll.LinkBelongsTo[CP, CID, PP, PID](children, parents, foreignKey, relationFieldPtr)
	if err != nil {
		return nil, err
	}
	return parents, nil
}

func LoadHasMany[
	PP model.Identifiable[PID],
	PID comparable,
	C any, // Model struct
	CP ScannableIdentifiable[C, CID],
	CID comparable,
](
	ctx context.Context,
	dbClient Client,
	parents *coll.Collection[PP, PID],
	sqlSelectBase string,
	foreignKeyColumn Column, // on the child
	foreignKey func(CP) PID, // on the child
	relationFieldPtr func(PP) **coll.Collection[CP, CID], // on the parent
) (*coll.Collection[CP, CID], error) {
	sqlStmt := sqlSelectBase + fmt.Sprintf(" WHERE %s IN (%s)", foreignKeyColumn.Name(),
		dbClient.Placeholders(parents.Len(), 2))
	parentIDsAsAny := parents.IDsAsAny()
	children, err := RawQueryCollection[C, CP, CID](ctx, dbClient, sqlStmt, parentIDsAsAny...)
	if err != nil {
		return nil, err
	}
	coll.LinkHasMany[PP, PID, CP, CID](
		parents,
		children,
		foreignKey,
		relationFieldPtr,
	)
	return children, nil
}
