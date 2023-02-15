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
from domain
where name = ?
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindDomainGlobalByName2 查询指定域名的全局解析
-- name: FindDomainGlobalByName2 :many
select *
from domain
where (name = ? or name = ?)
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindDomainGlobalByName3 查询指定域名的全局解析
-- name: FindDomainGlobalByName3 :many
select *
from domain
where (name = ? or name = ? or name = ?)
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindDomainGlobalByName4 查询指定域名的全局解析
-- name: FindDomainGlobalByName4 :many
select *
from domain
where (name = ? or name = ? or name = ? or name = ?)
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindDomainGlobalByName5 查询指定域名的全局解析
-- name: FindDomainGlobalByName5 :many
select *
from domain
where (name = ? or name = ? or name = ? or name = ? or name = ?)
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindDomainGlobalByName6 查询指定域名的全局解析
-- name: FindDomainGlobalByName6 :many
select *
from domain
where (name = ? or name = ? or name = ? or name = ? or name = ? or name = ?)
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindForwardByHostAndNameLike1 查询客户端专属的转发
-- name: FindForwardByHostAndNameLike1 :many
select *
from forward
where host = ?
  and name = ?;

-- FindForwardByHostAndNameLike2 查询客户端专属的转发
-- name: FindForwardByHostAndNameLike2 :many
select *
from forward
where host = ?
  and (name = ? or name = ?);

-- FindForwardByHostAndNameLike3 查询客户端专属的转发
-- name: FindForwardByHostAndNameLike3 :many
select *
from forward
where host = ?
  and (name = ? or name = ? or name = ?);

-- FindForwardByHostAndNameLike4 查询客户端专属的转发
-- name: FindForwardByHostAndNameLike4 :many
select *
from forward
where host = ?
  and (name = ? or name = ? or name = ? or name = ?);

-- FindForwardByHostAndNameLike5 查询客户端专属的转发
-- name: FindForwardByHostAndNameLike5 :many
select *
from forward
where host = ?
  and (name = ? or name = ? or name = ? or name = ? or name = ?);

-- FindForwardByHostAndNameLike6 查询客户端专属的转发
-- name: FindForwardByHostAndNameLike6 :many
select *
from forward
where host = ?
  and (name = ? or name = ? or name = ? or name = ? or name = ? or name = ?);

-- FindForwardGlobalByName1 查询指定域名的全局转发
-- name: FindForwardGlobalByName1 :many
select *
from forward
where name = ?
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindForwardGlobalByName2 查询指定域名的全局转发
-- name: FindForwardGlobalByName2 :many
select *
from forward
where (name = ? or name = ?)
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindForwardGlobalByName3 查询指定域名的全局转发
-- name: FindForwardGlobalByName3 :many
select *
from forward
where (name = ? or name = ? or name = ?)
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindForwardGlobalByName4 查询指定域名的全局转发
-- name: FindForwardGlobalByName4 :many
select *
from forward
where (name = ? or name = ? or name = ? or name = ?)
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindForwardGlobalByName5 查询指定域名的全局转发
-- name: FindForwardGlobalByName5 :many
select *
from forward
where (name = ? or name = ? or name = ? or name = ? or name = ?)
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindForwardGlobalByName6 查询指定域名的全局转发
-- name: FindForwardGlobalByName6 :many
select *
from forward
where (name = ? or name = ? or name = ? or name = ? or name = ? or name = ?)
  and (host = ''
    or host is null)
  and status = 'ENABLE';

-- FindHistoryByName 查询历史
-- name: FindHistoryByName :one
select *
from history
where name = ?;

-- InsertHistory 插入历史
-- name: InsertHistory :exec
insert into history(name, history)
VALUES (?, ?);

-- UpdateHistory 更新历史
-- name: UpdateHistory :exec
update history
set history = ?
where name = ?;
