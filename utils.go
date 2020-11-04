package pgrest

import (
	"fmt"
	"reflect"

	"github.com/go-pg/pg/v10/orm"
	"github.com/go-pg/pg/v10/types"
)

func setPk(resourceType reflect.Type, elem reflect.Value, key string) error {
	table := orm.GetTable(resourceType)
	if len(table.PKs) == 1 {
		pk := table.PKs[0]
		return pk.ScanValue(elem, NewBytesReader([]byte(key)), len(key))
	}
	return NewErrorBadRequest(fmt.Sprintf("only single pk is permitted for resource '%v'", resourceType))
}

func addQueryLimit(query *orm.Query, limit int) *orm.Query {
	if limit == 0 {
		return query
	}
	return query.Limit(int(limit))
}

func addQueryOffset(query *orm.Query, offset int) *orm.Query {
	if offset == 0 {
		return query
	}
	return query.Offset(int(offset))
}

func addQueryFields(query *orm.Query, fields []*Field) *orm.Query {
	if fields == nil {
		return query
	}
	q := query
	if len(fields) > 0 {
		for _, field := range fields {
			q = q.Column(field.Name)
		}
	}
	return q
}

func addQueryRelations(query *orm.Query, relations []*Relation) *orm.Query {
	if relations == nil {
		return query
	}
	q := query
	if len(relations) > 0 {
		for _, relation := range relations {
			q = q.Relation(relation.Name)
		}
	}
	return q
}

func addQuerySorts(query *orm.Query, sorts []*Sort) *orm.Query {
	if sorts == nil {
		return query
	}
	q := query
	if len(sorts) > 0 {
		for _, sort := range sorts {
			var orderExpr string
			if sort.Asc {
				orderExpr = sort.Name + " ASC"
			} else {
				orderExpr = sort.Name + " DESC"
			}
			q = q.Order(orderExpr)
		}
	}
	return q
}

func addQueryFilter(query *orm.Query, filter *Filter, parentGroupOp Op) *orm.Query {
	if filter == nil {
		return query
	}

	if filter.Op == And || filter.Op == Or {
		return addWhereGroup(query,
			func(query *orm.Query) (*orm.Query, error) {
				q := query
				for _, subfilter := range filter.Filters {
					q = addQueryFilter(query, subfilter, filter.Op)
				}
				return q, nil
			},
			parentGroupOp)
	}

	switch filter.Op {
	case Eq:
		return addWhere(query, "? = ?", filter.Attr, filter.Value, parentGroupOp)
	case Neq:
		return addWhere(query, "? != ?", filter.Attr, filter.Value, parentGroupOp)
	case In:
		return addWhere(query, "? IN (?)", filter.Attr, types.In(filter.Value), parentGroupOp)
	case Nin:
		return addWhere(query, "? NOT IN (?)", filter.Attr, types.In(filter.Value), parentGroupOp)
	case Gt:
		return addWhere(query, "? > ?", filter.Attr, filter.Value, parentGroupOp)
	case Gte:
		return addWhere(query, "? >= ?", filter.Attr, filter.Value, parentGroupOp)
	case Lt:
		return addWhere(query, "? < ?", filter.Attr, filter.Value, parentGroupOp)
	case Lte:
		return addWhere(query, "? <= ?", filter.Attr, filter.Value, parentGroupOp)
	case Lk:
		return addWhere(query, "? LIKE ?", filter.Attr, filter.Value, parentGroupOp)
	case Nlk:
		return addWhere(query, "? NOT LIKE ?", filter.Attr, filter.Value, parentGroupOp)
	case Ilk:
		return addWhere(query, "? ILIKE ?", filter.Attr, filter.Value, parentGroupOp)
	case Nilk:
		return addWhere(query, "? NOT ILIKE ?", filter.Attr, filter.Value, parentGroupOp)
	case Sim:
		return addWhere(query, "? SIMILAR TO ?", filter.Attr, filter.Value, parentGroupOp)
	case Nsim:
		return addWhere(query, "? NOT SIMILAR TO ?", filter.Attr, filter.Value, parentGroupOp)
	case Ilkua:
		return addWhere(query, "unaccent(?) ILIKE unaccent(?)", filter.Attr, filter.Value, parentGroupOp)
	case Nilkua:
		return addWhere(query, "unaccent(?) NOT ILIKE unaccent(?)", filter.Attr, filter.Value, parentGroupOp)
	case Null:
		return addWhere(query, "? IS NULL", filter.Attr, "", parentGroupOp)
	case Nnull:
		return addWhere(query, "? IS NOT NULL", filter.Attr, "", parentGroupOp)
	default:
		return query
	}
}

func addWhere(query *orm.Query, condition string, attribute string, value interface{}, parentGroupOp Op) *orm.Query {
	if parentGroupOp == Or {
		return query.WhereOr(condition, types.Ident(attribute), value)
	}
	return query.Where(condition, types.Ident(attribute), value)
}

func addWhereGroup(query *orm.Query, fnGroup func(query *orm.Query) (*orm.Query, error), parentGroupOp Op) *orm.Query {
	if parentGroupOp == Or {
		return query.WhereOrGroup(fnGroup)
	}
	return query.WhereGroup(fnGroup)
}
