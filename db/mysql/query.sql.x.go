package mysql

import "context"

// FindDomainByHostAndNameLike 查询客户端专属的解析
func (q *Queries) FindDomainByHostAndNameLike(ctx context.Context, host string, names []string) ([]Domain, error) {
	switch len(names) {
	case 0:
		return nil, nil
	case 1:
		return q.FindDomainByHostAndNameLike1(ctx, FindDomainByHostAndNameLike1Params{
			Host: host,
			Name: names[0],
		})
	case 2:
		return q.FindDomainByHostAndNameLike2(ctx, FindDomainByHostAndNameLike2Params{
			Host:   host,
			Name:   names[0],
			Name_2: names[1],
		})
	case 3:
		return q.FindDomainByHostAndNameLike3(ctx, FindDomainByHostAndNameLike3Params{
			Host:   host,
			Name:   names[0],
			Name_2: names[1],
			Name_3: names[2],
		})
	case 4:
		return q.FindDomainByHostAndNameLike4(ctx, FindDomainByHostAndNameLike4Params{
			Host:   host,
			Name:   names[0],
			Name_2: names[1],
			Name_3: names[2],
			Name_4: names[3],
		})
	case 5:
		return q.FindDomainByHostAndNameLike5(ctx, FindDomainByHostAndNameLike5Params{
			Host:   host,
			Name:   names[0],
			Name_2: names[1],
			Name_3: names[2],
			Name_4: names[3],
			Name_5: names[4],
		})
	default: // 最多支持到 6 级域名
		return q.FindDomainByHostAndNameLike6(ctx, FindDomainByHostAndNameLike6Params{
			Host:   host,
			Name:   names[0],
			Name_2: names[len(names)-1],
			Name_3: names[len(names)-2],
			Name_4: names[len(names)-3],
			Name_5: names[len(names)-4],
			Name_6: names[len(names)-5],
		})
	}
}

// FindDomainGlobalByName 查询指定域名的全局解析
func (q *Queries) FindDomainGlobalByName(ctx context.Context, names []string) ([]Domain, error) {
	switch len(names) {
	case 0:
		return nil, nil
	case 1:
		return q.FindDomainGlobalByName1(ctx, names[0])
	case 2:
		return q.FindDomainGlobalByName2(ctx, FindDomainGlobalByName2Params{
			Name:   names[0],
			Name_2: names[1],
		})
	case 3:
		return q.FindDomainGlobalByName3(ctx, FindDomainGlobalByName3Params{
			Name:   names[0],
			Name_2: names[1],
			Name_3: names[2],
		})
	case 4:
		return q.FindDomainGlobalByName4(ctx, FindDomainGlobalByName4Params{
			Name:   names[0],
			Name_2: names[1],
			Name_3: names[2],
			Name_4: names[3],
		})
	case 5:
		return q.FindDomainGlobalByName5(ctx, FindDomainGlobalByName5Params{
			Name:   names[0],
			Name_2: names[1],
			Name_3: names[2],
			Name_4: names[3],
			Name_5: names[4],
		})
	default: // 最多支持到 6 级域名
		return q.FindDomainGlobalByName6(ctx, FindDomainGlobalByName6Params{
			Name:   names[0],
			Name_2: names[len(names)-1],
			Name_3: names[len(names)-2],
			Name_4: names[len(names)-3],
			Name_5: names[len(names)-4],
			Name_6: names[len(names)-5],
		})
	}
}
