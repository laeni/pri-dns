// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: query.sql

package mysql

import (
	"context"
)

const countForward = `-- name: CountForward :one
select count(*)
from forward
`

// CountForward 统计forward记录数
func (q *Queries) CountForward(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countForward)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const findDomainByHostAndNameLike1 = `-- name: FindDomainByHostAndNameLike1 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from domain
where host = ?
  and name = ?
`

type FindDomainByHostAndNameLike1Params struct {
	Host string
	Name string
}

// FindDomainByHostAndNameLike1 查询客户端专属的解析
func (q *Queries) FindDomainByHostAndNameLike1(ctx context.Context, arg FindDomainByHostAndNameLike1Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainByHostAndNameLike1, arg.Host, arg.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainByHostAndNameLike2 = `-- name: FindDomainByHostAndNameLike2 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from domain
where host = ?
  and (name = ? or name = ?)
`

type FindDomainByHostAndNameLike2Params struct {
	Host   string
	Name   string
	Name_2 string
}

// FindDomainByHostAndNameLike2 查询客户端专属的解析
func (q *Queries) FindDomainByHostAndNameLike2(ctx context.Context, arg FindDomainByHostAndNameLike2Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainByHostAndNameLike2, arg.Host, arg.Name, arg.Name_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainByHostAndNameLike3 = `-- name: FindDomainByHostAndNameLike3 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from domain
where host = ?
  and (name = ? or name = ? or name = ?)
`

type FindDomainByHostAndNameLike3Params struct {
	Host   string
	Name   string
	Name_2 string
	Name_3 string
}

// FindDomainByHostAndNameLike3 查询客户端专属的解析
func (q *Queries) FindDomainByHostAndNameLike3(ctx context.Context, arg FindDomainByHostAndNameLike3Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainByHostAndNameLike3,
		arg.Host,
		arg.Name,
		arg.Name_2,
		arg.Name_3,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainByHostAndNameLike4 = `-- name: FindDomainByHostAndNameLike4 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from domain
where host = ?
  and (name = ? or name = ? or name = ? or name = ?)
`

type FindDomainByHostAndNameLike4Params struct {
	Host   string
	Name   string
	Name_2 string
	Name_3 string
	Name_4 string
}

// FindDomainByHostAndNameLike4 查询客户端专属的解析
func (q *Queries) FindDomainByHostAndNameLike4(ctx context.Context, arg FindDomainByHostAndNameLike4Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainByHostAndNameLike4,
		arg.Host,
		arg.Name,
		arg.Name_2,
		arg.Name_3,
		arg.Name_4,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainByHostAndNameLike5 = `-- name: FindDomainByHostAndNameLike5 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from domain
where host = ?
  and (name = ? or name = ? or name = ? or name = ? or name = ?)
`

type FindDomainByHostAndNameLike5Params struct {
	Host   string
	Name   string
	Name_2 string
	Name_3 string
	Name_4 string
	Name_5 string
}

// FindDomainByHostAndNameLike5 查询客户端专属的解析
func (q *Queries) FindDomainByHostAndNameLike5(ctx context.Context, arg FindDomainByHostAndNameLike5Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainByHostAndNameLike5,
		arg.Host,
		arg.Name,
		arg.Name_2,
		arg.Name_3,
		arg.Name_4,
		arg.Name_5,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainByHostAndNameLike6 = `-- name: FindDomainByHostAndNameLike6 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from domain
where host = ?
  and (name = ? or name = ? or name = ? or name = ? or name = ? or name = ?)
`

type FindDomainByHostAndNameLike6Params struct {
	Host   string
	Name   string
	Name_2 string
	Name_3 string
	Name_4 string
	Name_5 string
	Name_6 string
}

// FindDomainByHostAndNameLike6 查询客户端专属的解析
func (q *Queries) FindDomainByHostAndNameLike6(ctx context.Context, arg FindDomainByHostAndNameLike6Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainByHostAndNameLike6,
		arg.Host,
		arg.Name,
		arg.Name_2,
		arg.Name_3,
		arg.Name_4,
		arg.Name_5,
		arg.Name_6,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainGlobalByName1 = `-- name: FindDomainGlobalByName1 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from ` + "`" + `domain` + "`" + `
where ` + "`" + `name` + "`" + ` = ?
    and ` + "`" + `host` + "`" + ` = ''
   or ` + "`" + `host` + "`" + ` is null
`

// FindDomainGlobalByName1 查询指定域名的全局解析
func (q *Queries) FindDomainGlobalByName1(ctx context.Context, name string) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainGlobalByName1, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainGlobalByName2 = `-- name: FindDomainGlobalByName2 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from ` + "`" + `domain` + "`" + `
where (` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ?)
    and ` + "`" + `host` + "`" + ` = ''
   or ` + "`" + `host` + "`" + ` is null
`

type FindDomainGlobalByName2Params struct {
	Name   string
	Name_2 string
}

// FindDomainGlobalByName2 查询指定域名的全局解析
func (q *Queries) FindDomainGlobalByName2(ctx context.Context, arg FindDomainGlobalByName2Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainGlobalByName2, arg.Name, arg.Name_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainGlobalByName3 = `-- name: FindDomainGlobalByName3 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from ` + "`" + `domain` + "`" + `
where (` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ?)
    and ` + "`" + `host` + "`" + ` = ''
   or ` + "`" + `host` + "`" + ` is null
`

type FindDomainGlobalByName3Params struct {
	Name   string
	Name_2 string
	Name_3 string
}

// FindDomainGlobalByName3 查询指定域名的全局解析
func (q *Queries) FindDomainGlobalByName3(ctx context.Context, arg FindDomainGlobalByName3Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainGlobalByName3, arg.Name, arg.Name_2, arg.Name_3)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainGlobalByName4 = `-- name: FindDomainGlobalByName4 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from ` + "`" + `domain` + "`" + `
where (` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ?)
    and ` + "`" + `host` + "`" + ` = ''
   or ` + "`" + `host` + "`" + ` is null
`

type FindDomainGlobalByName4Params struct {
	Name   string
	Name_2 string
	Name_3 string
	Name_4 string
}

// FindDomainGlobalByName4 查询指定域名的全局解析
func (q *Queries) FindDomainGlobalByName4(ctx context.Context, arg FindDomainGlobalByName4Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainGlobalByName4,
		arg.Name,
		arg.Name_2,
		arg.Name_3,
		arg.Name_4,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainGlobalByName5 = `-- name: FindDomainGlobalByName5 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from ` + "`" + `domain` + "`" + `
where (` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ?)
    and ` + "`" + `host` + "`" + ` = ''
   or ` + "`" + `host` + "`" + ` is null
`

type FindDomainGlobalByName5Params struct {
	Name   string
	Name_2 string
	Name_3 string
	Name_4 string
	Name_5 string
}

// FindDomainGlobalByName5 查询指定域名的全局解析
func (q *Queries) FindDomainGlobalByName5(ctx context.Context, arg FindDomainGlobalByName5Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainGlobalByName5,
		arg.Name,
		arg.Name_2,
		arg.Name_3,
		arg.Name_4,
		arg.Name_5,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findDomainGlobalByName6 = `-- name: FindDomainGlobalByName6 :many
select id, host, name, value, ttl, status, type, priority, create_time, update_time
from ` + "`" + `domain` + "`" + `
where (` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ? or ` + "`" + `name` + "`" + ` = ?)
    and ` + "`" + `host` + "`" + ` = ''
   or ` + "`" + `host` + "`" + ` is null
`

type FindDomainGlobalByName6Params struct {
	Name   string
	Name_2 string
	Name_3 string
	Name_4 string
	Name_5 string
	Name_6 string
}

// FindDomainGlobalByName6 查询指定域名的全局解析
func (q *Queries) FindDomainGlobalByName6(ctx context.Context, arg FindDomainGlobalByName6Params) ([]Domain, error) {
	rows, err := q.db.QueryContext(ctx, findDomainGlobalByName6,
		arg.Name,
		arg.Name_2,
		arg.Name_3,
		arg.Name_4,
		arg.Name_5,
		arg.Name_6,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Domain
	for rows.Next() {
		var i Domain
		if err := rows.Scan(
			&i.ID,
			&i.Host,
			&i.Name,
			&i.Value,
			&i.Ttl,
			&i.Status,
			&i.Type,
			&i.Priority,
			&i.CreateTime,
			&i.UpdateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
