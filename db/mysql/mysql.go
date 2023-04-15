package mysql

import (
	"context"
	"database/sql"
	"fmt"
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
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()
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
			ClientHost: temp.ClientHost,
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
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()
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
			ClientHost: temp.ClientHost,
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
	if err != nil {
		return err
	}
	defer tx.Rollback()

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

func (s *StoreMysql) FindHistoryByHost(ctx context.Context, host string) []string {
	tx, err := s.db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	// 查询全局和客户端对应的转发域名
	forwards, err := s.queries.WithTx(tx).FindForwardByHost(ctx, host)
	if err != nil {
		panic(err)
	}

	// 去除否定用途的以及被否定数据否定的全局域名
	denied := make(map[string]struct{}, 0)
	for _, row := range forwards {
		if row.ClientHost != "" && row.DenyGlobal == "Y" {
			denied[row.Name] = struct{}{}
		}
	}
	j := 0
	for i, row := range forwards {
		if row.ClientHost == "" {
			if _, ok := denied[row.Name]; ok {
				continue
			}
		} else {
			if row.DenyGlobal == "Y" {
				continue
			}
		}
		forwards[j] = forwards[i]
		j++
	}
	forwards = forwards[:j]

	// 查询转发域名对应的解析历史
	his, err := findHistoryHostsByNames(ctx, tx, forwards)
	if err != nil {
		panic(err)
	}

	// 简单去重
	his = util.SliceDeduplication(his)

	// 根据IP范围语义进行合并
	his = cidr_merger.MergeIp(his, false)

	tx.Commit()

	return his
}

// 查询域名对应的解析IP历史
func findHistoryHostsByNames(ctx context.Context, tx *sql.Tx, forwards []FindForwardByHostRow) ([]string, error) {
	l := len(forwards)
	sqlStr := fmt.Sprintf("select history from history where %s", func() string {
		if l == 0 {
			return "0=1"
		}
		args := make([]string, l)
		for i := 0; i < l; i++ {
			args[i] = "?"
		}
		return "name in (" + strings.Join(args, ", ") + ")"
	}())

	args := make([]interface{}, l)
	for i, s := range forwards {
		args[i] = s.Name
	}

	rows, err := tx.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var history sql.NullString
		if err := rows.Scan(&history); err != nil {
			return nil, err
		}
		if history.String != "" {
			items = append(items, strings.Split(history.String, ",")...)
		}
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
