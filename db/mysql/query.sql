-- CountForward 统计forward记录数
-- name: CountForward :one
select count(*)
from forward;

-- FindDomainByHostAndNameLike1 查询客户端专属的解析
-- name: FindDomainByHostAndNameLike1 :many
select *
from domain
where host = ?
  and name = ?;

-- FindDomainByHostAndNameLike2 查询客户端专属的解析
-- name: FindDomainByHostAndNameLike2 :many
select *
from domain
where host = ?
  and (name = ? or name = ?);

-- FindDomainByHostAndNameLike3 查询客户端专属的解析
-- name: FindDomainByHostAndNameLike3 :many
select *
from domain
where host = ?
  and (name = ? or name = ? or name = ?);

-- FindDomainByHostAndNameLike4 查询客户端专属的解析
-- name: FindDomainByHostAndNameLike4 :many
select *
from domain
where host = ?
  and (name = ? or name = ? or name = ? or name = ?);

-- FindDomainByHostAndNameLike5 查询客户端专属的解析
-- name: FindDomainByHostAndNameLike5 :many
select *
from domain
where host = ?
  and (name = ? or name = ? or name = ? or name = ? or name = ?);

-- FindDomainByHostAndNameLike6 查询客户端专属的解析
-- name: FindDomainByHostAndNameLike6 :many
select *
from domain
where host = ?
  and (name = ? or name = ? or name = ? or name = ? or name = ? or name = ?);

-- FindDomainGlobalByName1 查询指定域名的全局解析
-- name: FindDomainGlobalByName1 :many
select *
from `domain`
where `name` = ?
    and `host` = ''
   or `host` is null;

-- FindDomainGlobalByName2 查询指定域名的全局解析
-- name: FindDomainGlobalByName2 :many
select *
from `domain`
where (`name` = ? or `name` = ?)
    and `host` = ''
   or `host` is null;

-- FindDomainGlobalByName3 查询指定域名的全局解析
-- name: FindDomainGlobalByName3 :many
select *
from `domain`
where (`name` = ? or `name` = ? or `name` = ?)
    and `host` = ''
   or `host` is null;

-- FindDomainGlobalByName4 查询指定域名的全局解析
-- name: FindDomainGlobalByName4 :many
select *
from `domain`
where (`name` = ? or `name` = ? or `name` = ? or `name` = ?)
    and `host` = ''
   or `host` is null;

-- FindDomainGlobalByName5 查询指定域名的全局解析
-- name: FindDomainGlobalByName5 :many
select *
from `domain`
where (`name` = ? or `name` = ? or `name` = ? or `name` = ? or `name` = ?)
    and `host` = ''
   or `host` is null;

-- FindDomainGlobalByName6 查询指定域名的全局解析
-- name: FindDomainGlobalByName6 :many
select *
from `domain`
where (`name` = ? or `name` = ? or `name` = ? or `name` = ? or `name` = ? or `name` = ?)
    and `host` = ''
   or `host` is null;
