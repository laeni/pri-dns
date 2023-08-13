package mysql

import (
	"database/sql"
	"errors"
	cidr_merger "github.com/laeni/pri-dns/cidr-merger"
	"github.com/laeni/pri-dns/db"
	"github.com/laeni/pri-dns/util"
	"gorm.io/gorm"
	"strings"
)

type StoreMysql struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) StoreMysql {
	return StoreMysql{db: db}
}

func (s *StoreMysql) FindForwardByHostAndName(host, name string) []db.Forward {
	names := util.GenAllMatchDomain(name)

	var forwardTemps []Forward
	s.db.Where("name IN ? AND (client_host IS NULL OR client_host = '' OR client_host = ?)", names, host).Find(&forwardTemps)

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

func (s *StoreMysql) FindDomainByHostAndName(host, name string) []db.Domain {
	names := util.GenAllMatchDomain(name)

	var domainTemps []Domain
	s.db.Where("name IN ? AND (client_host IS NULL OR client_host = '' OR client_host = ?)", names, host).Find(&domainTemps)

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

func (s *StoreMysql) SavaHistory(name string, newHis []string) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	var oldHis []string
	var historyTmp History
	err := s.db.Where("name = ?", name).Take(&historyTmp).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	} else {
		oldHis = strings.Split(historyTmp.History.String, ",")
	}

	// 合并新老历史
	ipHis := append(newHis, oldHis...)
	// 从语义上再次进行合并
	ipHis = mergeIp(ipHis)

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
			his := History{Name: name, History: sql.NullString{Valid: true, String: strings.Join(ipHis, ",")}}
			err = s.db.Create(&his).Error
		} else {
			err = s.db.Model(&History{}).Where("name = ?", name).
				Update("history", sql.NullString{Valid: true, String: strings.Join(ipHis, ",")}).Error
		}
		if err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func (s *StoreMysql) FindHistoryByHost(host string) ([]string, []string) {
	tx := s.db.Begin()
	if tx.Error != nil {
		panic(tx.Error)
	}
	defer tx.Rollback()

	// 查询全局和客户端对应的转发域名
	var forwards []Forward
	err := tx.Where("status = 'ENABLE' AND (client_host IS NULL OR client_host = '' OR client_host = ?)").Find(&forwards).Error
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
	names := make([]string, len(forwards))
	for i, s := range forwards {
		names[i] = s.Name
	}
	his, err := findHistoryHostsByNames(tx, names)
	if err != nil {
		panic(err)
	}
	// 简单去重
	his = util.SliceDeduplication(his)

	// 查询需要排除的网段，比如内网网段
	var historyExes []HistoryEx
	err = tx.Where("client_host IS NULL OR client_host = '' OR client_host = ?", host).Find(&historyExes).Error
	if err != nil {
		panic(err)
	}
	husExStr := make([]string, 0, len(historyExes))
	{
		// 去除否定用途的记录
		denied := make(map[string]struct{}, 0)
		for _, ex := range historyExes {
			if ex.ClientHost != "" && ex.DenyGlobal == "Y" {
				denied[ex.IpNet] = struct{}{}
			}
		}
		for _, ex := range historyExes {
			if ex.ClientHost == "" {
				if _, ok := denied[ex.IpNet]; ok {
					continue
				}
			}
			if ex.DenyGlobal == "Y" {
				continue
			}
			husExStr = append(husExStr, ex.IpNet)
		}
	}

	tx.Commit()

	// 根据IP范围语义进行合并
	return mergeIp(his), husExStr
}

// 查询域名对应的解析IP历史
func findHistoryHostsByNames(tx *gorm.DB, names []string) ([]string, error) {
	var his []History
	err := tx.Where("name IN ?", names).Find(&his).Error
	if err != nil {
		return nil, err
	}

	var items []string
	for _, item := range his {
		if item.History.String != "" {
			items = append(items, strings.Split(item.History.String, ",")...)
		}
	}
	return items, nil
}

// mergeIp 对ip进行合并
func mergeIp(in []string) []string {
	result := cidr_merger.IpNetToRange(cidr_merger.StrToIpNet(in))
	result = cidr_merger.SortAndMerge(result)
	return cidr_merger.IpNetToString(cidr_merger.IpRangeToIpNet(result))
}
