package mysql

import (
	"context"
	"database/sql"
	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/laeni/pri-dns/db"
	"github.com/laeni/pri-dns/util"
)

type StoreMysql struct {
	db *sql.DB
}

func NewStore(db *sql.DB) StoreMysql {
	return StoreMysql{db: db}
}

func (s *StoreMysql) FindForwardByClient(host string) []db.Forward {
	return nil
}

func (s *StoreMysql) FindDomainByHostAndName(ctx context.Context, host, name string) []db.Domain {
	names := util.GenAllMatchDomain(name)

	queries := New(s.db)
	var domainTemps []Domain
	var err error
	if host != "" {
		domainTemps, err = queries.FindDomainByHostAndNameLike(ctx, host, names)
	} else {
		domainTemps, err = queries.FindDomainGlobalByName(ctx, names)
	}
	if err != nil {
		log.Error("从数据库查询解析记录失败", err)
	}

	domains := make([]db.Domain, len(domainTemps))
	for i := 0; i < len(domainTemps); i++ {
		temp := domainTemps[i]
		domains[i] = db.Domain{
			ID:         temp.ID,
			Host:       temp.Host,
			Name:       temp.Name,
			Value:      temp.Value.String,
			Ttl:        temp.Ttl.Int32,
			Status:     temp.Status,
			Type:       temp.Type.String,
			Priority:   temp.Priority,
			CreateTime: temp.CreateTime,
			UpdateTime: temp.UpdateTime,
		}
	}
	return domains
}
