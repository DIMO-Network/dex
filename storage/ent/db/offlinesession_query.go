// Code generated by entc, DO NOT EDIT.

package db

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/dexidp/dex/storage/ent/db/offlinesession"
	"github.com/dexidp/dex/storage/ent/db/predicate"
	"github.com/facebook/ent/dialect/sql"
	"github.com/facebook/ent/dialect/sql/sqlgraph"
	"github.com/facebook/ent/schema/field"
)

// OfflineSessionQuery is the builder for querying OfflineSession entities.
type OfflineSessionQuery struct {
	config
	limit      *int
	offset     *int
	order      []OrderFunc
	fields     []string
	predicates []predicate.OfflineSession
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the OfflineSessionQuery builder.
func (osq *OfflineSessionQuery) Where(ps ...predicate.OfflineSession) *OfflineSessionQuery {
	osq.predicates = append(osq.predicates, ps...)
	return osq
}

// Limit adds a limit step to the query.
func (osq *OfflineSessionQuery) Limit(limit int) *OfflineSessionQuery {
	osq.limit = &limit
	return osq
}

// Offset adds an offset step to the query.
func (osq *OfflineSessionQuery) Offset(offset int) *OfflineSessionQuery {
	osq.offset = &offset
	return osq
}

// Order adds an order step to the query.
func (osq *OfflineSessionQuery) Order(o ...OrderFunc) *OfflineSessionQuery {
	osq.order = append(osq.order, o...)
	return osq
}

// First returns the first OfflineSession entity from the query.
// Returns a *NotFoundError when no OfflineSession was found.
func (osq *OfflineSessionQuery) First(ctx context.Context) (*OfflineSession, error) {
	nodes, err := osq.Limit(1).All(ctx)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{offlinesession.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (osq *OfflineSessionQuery) FirstX(ctx context.Context) *OfflineSession {
	node, err := osq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first OfflineSession ID from the query.
// Returns a *NotFoundError when no OfflineSession ID was found.
func (osq *OfflineSessionQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = osq.Limit(1).IDs(ctx); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{offlinesession.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (osq *OfflineSessionQuery) FirstIDX(ctx context.Context) string {
	id, err := osq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single OfflineSession entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when exactly one OfflineSession entity is not found.
// Returns a *NotFoundError when no OfflineSession entities are found.
func (osq *OfflineSessionQuery) Only(ctx context.Context) (*OfflineSession, error) {
	nodes, err := osq.Limit(2).All(ctx)
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{offlinesession.Label}
	default:
		return nil, &NotSingularError{offlinesession.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (osq *OfflineSessionQuery) OnlyX(ctx context.Context) *OfflineSession {
	node, err := osq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only OfflineSession ID in the query.
// Returns a *NotSingularError when exactly one OfflineSession ID is not found.
// Returns a *NotFoundError when no entities are found.
func (osq *OfflineSessionQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = osq.Limit(2).IDs(ctx); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{offlinesession.Label}
	default:
		err = &NotSingularError{offlinesession.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (osq *OfflineSessionQuery) OnlyIDX(ctx context.Context) string {
	id, err := osq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of OfflineSessions.
func (osq *OfflineSessionQuery) All(ctx context.Context) ([]*OfflineSession, error) {
	if err := osq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	return osq.sqlAll(ctx)
}

// AllX is like All, but panics if an error occurs.
func (osq *OfflineSessionQuery) AllX(ctx context.Context) []*OfflineSession {
	nodes, err := osq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of OfflineSession IDs.
func (osq *OfflineSessionQuery) IDs(ctx context.Context) ([]string, error) {
	var ids []string
	if err := osq.Select(offlinesession.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (osq *OfflineSessionQuery) IDsX(ctx context.Context) []string {
	ids, err := osq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (osq *OfflineSessionQuery) Count(ctx context.Context) (int, error) {
	if err := osq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return osq.sqlCount(ctx)
}

// CountX is like Count, but panics if an error occurs.
func (osq *OfflineSessionQuery) CountX(ctx context.Context) int {
	count, err := osq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (osq *OfflineSessionQuery) Exist(ctx context.Context) (bool, error) {
	if err := osq.prepareQuery(ctx); err != nil {
		return false, err
	}
	return osq.sqlExist(ctx)
}

// ExistX is like Exist, but panics if an error occurs.
func (osq *OfflineSessionQuery) ExistX(ctx context.Context) bool {
	exist, err := osq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the OfflineSessionQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (osq *OfflineSessionQuery) Clone() *OfflineSessionQuery {
	if osq == nil {
		return nil
	}
	return &OfflineSessionQuery{
		config:     osq.config,
		limit:      osq.limit,
		offset:     osq.offset,
		order:      append([]OrderFunc{}, osq.order...),
		predicates: append([]predicate.OfflineSession{}, osq.predicates...),
		// clone intermediate query.
		sql:  osq.sql.Clone(),
		path: osq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		UserID string `json:"user_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.OfflineSession.Query().
//		GroupBy(offlinesession.FieldUserID).
//		Aggregate(db.Count()).
//		Scan(ctx, &v)
//
func (osq *OfflineSessionQuery) GroupBy(field string, fields ...string) *OfflineSessionGroupBy {
	group := &OfflineSessionGroupBy{config: osq.config}
	group.fields = append([]string{field}, fields...)
	group.path = func(ctx context.Context) (prev *sql.Selector, err error) {
		if err := osq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		return osq.sqlQuery(), nil
	}
	return group
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		UserID string `json:"user_id,omitempty"`
//	}
//
//	client.OfflineSession.Query().
//		Select(offlinesession.FieldUserID).
//		Scan(ctx, &v)
//
func (osq *OfflineSessionQuery) Select(field string, fields ...string) *OfflineSessionSelect {
	osq.fields = append([]string{field}, fields...)
	return &OfflineSessionSelect{OfflineSessionQuery: osq}
}

func (osq *OfflineSessionQuery) prepareQuery(ctx context.Context) error {
	for _, f := range osq.fields {
		if !offlinesession.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("db: invalid field %q for query", f)}
		}
	}
	if osq.path != nil {
		prev, err := osq.path(ctx)
		if err != nil {
			return err
		}
		osq.sql = prev
	}
	return nil
}

func (osq *OfflineSessionQuery) sqlAll(ctx context.Context) ([]*OfflineSession, error) {
	var (
		nodes = []*OfflineSession{}
		_spec = osq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]interface{}, error) {
		node := &OfflineSession{config: osq.config}
		nodes = append(nodes, node)
		return node.scanValues(columns)
	}
	_spec.Assign = func(columns []string, values []interface{}) error {
		if len(nodes) == 0 {
			return fmt.Errorf("db: Assign called without calling ScanValues")
		}
		node := nodes[len(nodes)-1]
		return node.assignValues(columns, values)
	}
	if err := sqlgraph.QueryNodes(ctx, osq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (osq *OfflineSessionQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := osq.querySpec()
	return sqlgraph.CountNodes(ctx, osq.driver, _spec)
}

func (osq *OfflineSessionQuery) sqlExist(ctx context.Context) (bool, error) {
	n, err := osq.sqlCount(ctx)
	if err != nil {
		return false, fmt.Errorf("db: check existence: %v", err)
	}
	return n > 0, nil
}

func (osq *OfflineSessionQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table:   offlinesession.Table,
			Columns: offlinesession.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeString,
				Column: offlinesession.FieldID,
			},
		},
		From:   osq.sql,
		Unique: true,
	}
	if fields := osq.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, offlinesession.FieldID)
		for i := range fields {
			if fields[i] != offlinesession.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := osq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := osq.limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := osq.offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := osq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector, offlinesession.ValidColumn)
			}
		}
	}
	return _spec
}

func (osq *OfflineSessionQuery) sqlQuery() *sql.Selector {
	builder := sql.Dialect(osq.driver.Dialect())
	t1 := builder.Table(offlinesession.Table)
	selector := builder.Select(t1.Columns(offlinesession.Columns...)...).From(t1)
	if osq.sql != nil {
		selector = osq.sql
		selector.Select(selector.Columns(offlinesession.Columns...)...)
	}
	for _, p := range osq.predicates {
		p(selector)
	}
	for _, p := range osq.order {
		p(selector, offlinesession.ValidColumn)
	}
	if offset := osq.offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := osq.limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// OfflineSessionGroupBy is the group-by builder for OfflineSession entities.
type OfflineSessionGroupBy struct {
	config
	fields []string
	fns    []AggregateFunc
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Aggregate adds the given aggregation functions to the group-by query.
func (osgb *OfflineSessionGroupBy) Aggregate(fns ...AggregateFunc) *OfflineSessionGroupBy {
	osgb.fns = append(osgb.fns, fns...)
	return osgb
}

// Scan applies the group-by query and scans the result into the given value.
func (osgb *OfflineSessionGroupBy) Scan(ctx context.Context, v interface{}) error {
	query, err := osgb.path(ctx)
	if err != nil {
		return err
	}
	osgb.sql = query
	return osgb.sqlScan(ctx, v)
}

// ScanX is like Scan, but panics if an error occurs.
func (osgb *OfflineSessionGroupBy) ScanX(ctx context.Context, v interface{}) {
	if err := osgb.Scan(ctx, v); err != nil {
		panic(err)
	}
}

// Strings returns list of strings from group-by.
// It is only allowed when executing a group-by query with one field.
func (osgb *OfflineSessionGroupBy) Strings(ctx context.Context) ([]string, error) {
	if len(osgb.fields) > 1 {
		return nil, errors.New("db: OfflineSessionGroupBy.Strings is not achievable when grouping more than 1 field")
	}
	var v []string
	if err := osgb.Scan(ctx, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// StringsX is like Strings, but panics if an error occurs.
func (osgb *OfflineSessionGroupBy) StringsX(ctx context.Context) []string {
	v, err := osgb.Strings(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// String returns a single string from a group-by query.
// It is only allowed when executing a group-by query with one field.
func (osgb *OfflineSessionGroupBy) String(ctx context.Context) (_ string, err error) {
	var v []string
	if v, err = osgb.Strings(ctx); err != nil {
		return
	}
	switch len(v) {
	case 1:
		return v[0], nil
	case 0:
		err = &NotFoundError{offlinesession.Label}
	default:
		err = fmt.Errorf("db: OfflineSessionGroupBy.Strings returned %d results when one was expected", len(v))
	}
	return
}

// StringX is like String, but panics if an error occurs.
func (osgb *OfflineSessionGroupBy) StringX(ctx context.Context) string {
	v, err := osgb.String(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Ints returns list of ints from group-by.
// It is only allowed when executing a group-by query with one field.
func (osgb *OfflineSessionGroupBy) Ints(ctx context.Context) ([]int, error) {
	if len(osgb.fields) > 1 {
		return nil, errors.New("db: OfflineSessionGroupBy.Ints is not achievable when grouping more than 1 field")
	}
	var v []int
	if err := osgb.Scan(ctx, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// IntsX is like Ints, but panics if an error occurs.
func (osgb *OfflineSessionGroupBy) IntsX(ctx context.Context) []int {
	v, err := osgb.Ints(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Int returns a single int from a group-by query.
// It is only allowed when executing a group-by query with one field.
func (osgb *OfflineSessionGroupBy) Int(ctx context.Context) (_ int, err error) {
	var v []int
	if v, err = osgb.Ints(ctx); err != nil {
		return
	}
	switch len(v) {
	case 1:
		return v[0], nil
	case 0:
		err = &NotFoundError{offlinesession.Label}
	default:
		err = fmt.Errorf("db: OfflineSessionGroupBy.Ints returned %d results when one was expected", len(v))
	}
	return
}

// IntX is like Int, but panics if an error occurs.
func (osgb *OfflineSessionGroupBy) IntX(ctx context.Context) int {
	v, err := osgb.Int(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Float64s returns list of float64s from group-by.
// It is only allowed when executing a group-by query with one field.
func (osgb *OfflineSessionGroupBy) Float64s(ctx context.Context) ([]float64, error) {
	if len(osgb.fields) > 1 {
		return nil, errors.New("db: OfflineSessionGroupBy.Float64s is not achievable when grouping more than 1 field")
	}
	var v []float64
	if err := osgb.Scan(ctx, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Float64sX is like Float64s, but panics if an error occurs.
func (osgb *OfflineSessionGroupBy) Float64sX(ctx context.Context) []float64 {
	v, err := osgb.Float64s(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Float64 returns a single float64 from a group-by query.
// It is only allowed when executing a group-by query with one field.
func (osgb *OfflineSessionGroupBy) Float64(ctx context.Context) (_ float64, err error) {
	var v []float64
	if v, err = osgb.Float64s(ctx); err != nil {
		return
	}
	switch len(v) {
	case 1:
		return v[0], nil
	case 0:
		err = &NotFoundError{offlinesession.Label}
	default:
		err = fmt.Errorf("db: OfflineSessionGroupBy.Float64s returned %d results when one was expected", len(v))
	}
	return
}

// Float64X is like Float64, but panics if an error occurs.
func (osgb *OfflineSessionGroupBy) Float64X(ctx context.Context) float64 {
	v, err := osgb.Float64(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Bools returns list of bools from group-by.
// It is only allowed when executing a group-by query with one field.
func (osgb *OfflineSessionGroupBy) Bools(ctx context.Context) ([]bool, error) {
	if len(osgb.fields) > 1 {
		return nil, errors.New("db: OfflineSessionGroupBy.Bools is not achievable when grouping more than 1 field")
	}
	var v []bool
	if err := osgb.Scan(ctx, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// BoolsX is like Bools, but panics if an error occurs.
func (osgb *OfflineSessionGroupBy) BoolsX(ctx context.Context) []bool {
	v, err := osgb.Bools(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Bool returns a single bool from a group-by query.
// It is only allowed when executing a group-by query with one field.
func (osgb *OfflineSessionGroupBy) Bool(ctx context.Context) (_ bool, err error) {
	var v []bool
	if v, err = osgb.Bools(ctx); err != nil {
		return
	}
	switch len(v) {
	case 1:
		return v[0], nil
	case 0:
		err = &NotFoundError{offlinesession.Label}
	default:
		err = fmt.Errorf("db: OfflineSessionGroupBy.Bools returned %d results when one was expected", len(v))
	}
	return
}

// BoolX is like Bool, but panics if an error occurs.
func (osgb *OfflineSessionGroupBy) BoolX(ctx context.Context) bool {
	v, err := osgb.Bool(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

func (osgb *OfflineSessionGroupBy) sqlScan(ctx context.Context, v interface{}) error {
	for _, f := range osgb.fields {
		if !offlinesession.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("invalid field %q for group-by", f)}
		}
	}
	selector := osgb.sqlQuery()
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := osgb.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

func (osgb *OfflineSessionGroupBy) sqlQuery() *sql.Selector {
	selector := osgb.sql
	columns := make([]string, 0, len(osgb.fields)+len(osgb.fns))
	columns = append(columns, osgb.fields...)
	for _, fn := range osgb.fns {
		columns = append(columns, fn(selector, offlinesession.ValidColumn))
	}
	return selector.Select(columns...).GroupBy(osgb.fields...)
}

// OfflineSessionSelect is the builder for selecting fields of OfflineSession entities.
type OfflineSessionSelect struct {
	*OfflineSessionQuery
	// intermediate query (i.e. traversal path).
	sql *sql.Selector
}

// Scan applies the selector query and scans the result into the given value.
func (oss *OfflineSessionSelect) Scan(ctx context.Context, v interface{}) error {
	if err := oss.prepareQuery(ctx); err != nil {
		return err
	}
	oss.sql = oss.OfflineSessionQuery.sqlQuery()
	return oss.sqlScan(ctx, v)
}

// ScanX is like Scan, but panics if an error occurs.
func (oss *OfflineSessionSelect) ScanX(ctx context.Context, v interface{}) {
	if err := oss.Scan(ctx, v); err != nil {
		panic(err)
	}
}

// Strings returns list of strings from a selector. It is only allowed when selecting one field.
func (oss *OfflineSessionSelect) Strings(ctx context.Context) ([]string, error) {
	if len(oss.fields) > 1 {
		return nil, errors.New("db: OfflineSessionSelect.Strings is not achievable when selecting more than 1 field")
	}
	var v []string
	if err := oss.Scan(ctx, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// StringsX is like Strings, but panics if an error occurs.
func (oss *OfflineSessionSelect) StringsX(ctx context.Context) []string {
	v, err := oss.Strings(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// String returns a single string from a selector. It is only allowed when selecting one field.
func (oss *OfflineSessionSelect) String(ctx context.Context) (_ string, err error) {
	var v []string
	if v, err = oss.Strings(ctx); err != nil {
		return
	}
	switch len(v) {
	case 1:
		return v[0], nil
	case 0:
		err = &NotFoundError{offlinesession.Label}
	default:
		err = fmt.Errorf("db: OfflineSessionSelect.Strings returned %d results when one was expected", len(v))
	}
	return
}

// StringX is like String, but panics if an error occurs.
func (oss *OfflineSessionSelect) StringX(ctx context.Context) string {
	v, err := oss.String(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Ints returns list of ints from a selector. It is only allowed when selecting one field.
func (oss *OfflineSessionSelect) Ints(ctx context.Context) ([]int, error) {
	if len(oss.fields) > 1 {
		return nil, errors.New("db: OfflineSessionSelect.Ints is not achievable when selecting more than 1 field")
	}
	var v []int
	if err := oss.Scan(ctx, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// IntsX is like Ints, but panics if an error occurs.
func (oss *OfflineSessionSelect) IntsX(ctx context.Context) []int {
	v, err := oss.Ints(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Int returns a single int from a selector. It is only allowed when selecting one field.
func (oss *OfflineSessionSelect) Int(ctx context.Context) (_ int, err error) {
	var v []int
	if v, err = oss.Ints(ctx); err != nil {
		return
	}
	switch len(v) {
	case 1:
		return v[0], nil
	case 0:
		err = &NotFoundError{offlinesession.Label}
	default:
		err = fmt.Errorf("db: OfflineSessionSelect.Ints returned %d results when one was expected", len(v))
	}
	return
}

// IntX is like Int, but panics if an error occurs.
func (oss *OfflineSessionSelect) IntX(ctx context.Context) int {
	v, err := oss.Int(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Float64s returns list of float64s from a selector. It is only allowed when selecting one field.
func (oss *OfflineSessionSelect) Float64s(ctx context.Context) ([]float64, error) {
	if len(oss.fields) > 1 {
		return nil, errors.New("db: OfflineSessionSelect.Float64s is not achievable when selecting more than 1 field")
	}
	var v []float64
	if err := oss.Scan(ctx, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Float64sX is like Float64s, but panics if an error occurs.
func (oss *OfflineSessionSelect) Float64sX(ctx context.Context) []float64 {
	v, err := oss.Float64s(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Float64 returns a single float64 from a selector. It is only allowed when selecting one field.
func (oss *OfflineSessionSelect) Float64(ctx context.Context) (_ float64, err error) {
	var v []float64
	if v, err = oss.Float64s(ctx); err != nil {
		return
	}
	switch len(v) {
	case 1:
		return v[0], nil
	case 0:
		err = &NotFoundError{offlinesession.Label}
	default:
		err = fmt.Errorf("db: OfflineSessionSelect.Float64s returned %d results when one was expected", len(v))
	}
	return
}

// Float64X is like Float64, but panics if an error occurs.
func (oss *OfflineSessionSelect) Float64X(ctx context.Context) float64 {
	v, err := oss.Float64(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Bools returns list of bools from a selector. It is only allowed when selecting one field.
func (oss *OfflineSessionSelect) Bools(ctx context.Context) ([]bool, error) {
	if len(oss.fields) > 1 {
		return nil, errors.New("db: OfflineSessionSelect.Bools is not achievable when selecting more than 1 field")
	}
	var v []bool
	if err := oss.Scan(ctx, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// BoolsX is like Bools, but panics if an error occurs.
func (oss *OfflineSessionSelect) BoolsX(ctx context.Context) []bool {
	v, err := oss.Bools(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Bool returns a single bool from a selector. It is only allowed when selecting one field.
func (oss *OfflineSessionSelect) Bool(ctx context.Context) (_ bool, err error) {
	var v []bool
	if v, err = oss.Bools(ctx); err != nil {
		return
	}
	switch len(v) {
	case 1:
		return v[0], nil
	case 0:
		err = &NotFoundError{offlinesession.Label}
	default:
		err = fmt.Errorf("db: OfflineSessionSelect.Bools returned %d results when one was expected", len(v))
	}
	return
}

// BoolX is like Bool, but panics if an error occurs.
func (oss *OfflineSessionSelect) BoolX(ctx context.Context) bool {
	v, err := oss.Bool(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

func (oss *OfflineSessionSelect) sqlScan(ctx context.Context, v interface{}) error {
	rows := &sql.Rows{}
	query, args := oss.sqlQuery().Query()
	if err := oss.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

func (oss *OfflineSessionSelect) sqlQuery() sql.Querier {
	selector := oss.sql
	selector.Select(selector.Columns(oss.fields...)...)
	return selector
}
