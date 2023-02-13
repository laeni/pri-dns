package mysql

import (
	"context"
	"database/sql"
	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/laeni/pri-dns/db"
	"github.com/laeni/pri-dns/util"
	"strings"
)

type StoreMysql struct {
	queries *Queries
}

func NewStore(db *sql.DB) StoreMysql {
	return StoreMysql{queries: New(db)}
}

func (s *StoreMysql) FindForwardByHostAndName(ctx context.Context, host, name string) []db.Forward {
	names := util.GenAllMatchDomain(name)

	var forwardTemps []Forward
	var err error
	if host != "" {
		forwardTemps, err = s.queries.FindForwardByHostAndNameLike(ctx, host, names)
	} else {
		forwardTemps, err = s.queries.FindForwardGlobalByName(ctx, names)
	}
	if err != nil {
		log.Error("从数据库查询转发记录失败", err)
	}

	forwards := make([]db.Forward, len(forwardTemps))
	for i := 0; i < len(forwardTemps); i++ {
		temp := forwardTemps[i]

		var history []string
		if temp.History.String != "" {
			history = strings.Split(temp.History.String, ",")
		}

		var snsSvr []string
		if temp.DnsSvr != "" {
			snsSvr = strings.Split(temp.DnsSvr, ",")
		}

		forwards[i] = db.Forward{
			ID:         temp.ID,
			Host:       temp.Host,
			Name:       temp.Name,
			DnsSvr:     snsSvr,
			History:    history,
			Type:       temp.Type.String,
			Status:     temp.Status,
			CreateTime: temp.CreateTime,
			UpdateTime: temp.UpdateTime,
		}
	}
	return forwards
}

func (s *StoreMysql) FindDomainByHostAndName(ctx context.Context, host, name string) []db.Domain {
	names := util.GenAllMatchDomain(name)

	var domainTemps []Domain
	var err error
	if host != "" {
		domainTemps, err = s.queries.FindDomainByHostAndNameLike(ctx, host, names)
	} else {
		domainTemps, err = s.queries.FindDomainGlobalByName(ctx, names)
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
