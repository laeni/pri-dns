package mysql

import (
	"context"
	"database/sql"
	"github.com/laeni/pri-dns/db"
	"github.com/laeni/pri-dns/util"
	"strings"
)

type StoreMysql struct {
	db      *sql.DB
	queries *Queries
}

func NewStore(db *sql.DB) StoreMysql {
	return StoreMysql{db: db, queries: New(db)}
}

func (s *StoreMysql) FindForwardByHostAndName(ctx context.Context, host, name string) []db.Forward {
	names := util.GenAllMatchDomain(name)

	var forwardTemps []Forward
	var err error
	tx, err := s.db.Begin()
	defer tx.Rollback()
	if err != nil {
		panic(err)
	}
	if host != "" {
		forwardTemps, err = s.queries.WithTx(tx).FindForwardByHostAndNameLike(ctx, host, names)
	} else {
		forwardTemps, err = s.queries.WithTx(tx).FindForwardGlobalByName(ctx, names)
	}
	if err != nil {
		panic(err)
	}
	if err = tx.Commit(); err != nil {
		panic(err)
	}

	forwards := make([]db.Forward, len(forwardTemps))
	for i := 0; i < len(forwardTemps); i++ {
		temp := forwardTemps[i]

		var history []string
		if temp.History.String != "" {
			history = strings.Split(temp.History.String, ",")
		}

		var snsSvr []string
		if temp.DnsSvr.String != "" {
			snsSvr = strings.Split(temp.DnsSvr.String, ",")
		}

		forwards[i] = db.Forward{
			ID:         temp.ID,
			Host:       temp.Host,
			Name:       temp.Name,
			DnsSvr:     snsSvr,
			History:    history,
			DenyGlobal: strings.ToUpper(temp.DenyGlobal) == "Y",
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
	tx, err := s.db.Begin()
	defer tx.Rollback()
	if err != nil {
		panic(err)
	}
	if host != "" {
		domainTemps, err = s.queries.WithTx(tx).FindDomainByHostAndNameLike(ctx, host, names)
	} else {
		domainTemps, err = s.queries.WithTx(tx).FindDomainGlobalByName(ctx, names)
	}
	if err != nil {
		panic(err)
	}
	if err = tx.Commit(); err != nil {
		panic(err)
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
			DnsType:    temp.DnsType.String,
			DenyGlobal: strings.ToUpper(temp.DenyGlobal) == "Y",
			Status:     temp.Status,
			CreateTime: temp.CreateTime,
			UpdateTime: temp.UpdateTime,
		}
	}
	return domains
}
