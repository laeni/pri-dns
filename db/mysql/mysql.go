package mysql

import (
	"context"
	"database/sql"
	cidr_merger "github.com/laeni/pri-dns/cidr-merger"
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

		var snsSvr []string
		if temp.DnsSvr.String != "" {
			snsSvr = strings.Split(temp.DnsSvr.String, ",")
		}

		forwards[i] = db.Forward{
			ID:         temp.ID,
			Host:       temp.Host,
			Name:       temp.Name,
			DnsSvr:     snsSvr,
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

func (s *StoreMysql) SavaHistory(ctx context.Context, name string, newHis []string) error {
	tx, err := s.db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}

	var oldHis []string
	historyTmp, err := s.queries.WithTx(tx).FindHistoryByName(ctx, name)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else {
		oldHis = strings.Split(historyTmp.History.String, ",")
	}

	// 合并新老历史
	ipHis := append(newHis, oldHis...)
	// 从语义上再次进行合并
	ipHis = cidr_merger.MergeIp(ipHis, false)

	sava := len(oldHis) != len(ipHis)
	if !sava {
		for i := 0; i < len(ipHis); i++ {
			if oldHis[i] != ipHis[i] {
				sava = true
				break
			}
		}
	}
	if sava {
		if len(oldHis) == 0 {
			err = s.queries.WithTx(tx).InsertHistory(ctx, InsertHistoryParams{
				Name:    name,
				History: sql.NullString{Valid: true, String: strings.Join(ipHis, ",")},
			})
		} else {
			err = s.queries.WithTx(tx).UpdateHistory(ctx, UpdateHistoryParams{
				Name:    name,
				History: sql.NullString{Valid: true, String: strings.Join(ipHis, ",")},
			})
		}
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
