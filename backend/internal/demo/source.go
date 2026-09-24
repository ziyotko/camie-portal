package demo

import (
	"camie-portal/backend/internal/model"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"
	"time"
)

//go:embed content.json
var content []byte

// associationContent contains the body extracted from the supplied legacy introduction page.
//
//go:embed association.html
var associationContent string

//go:embed party-articles.json
var partyContent []byte

type Source struct {
	Columns []model.Column
	Items   map[int64][]model.Article
}

func NewSource() *Source {
	var fixture struct{ Titles, Dates, Paragraphs []string }
	if err := json.Unmarshal(content, &fixture); err != nil {
		panic(err)
	}
	s := &Source{Items: map[int64][]model.Article{}}
	names := []string{"行业资讯", "产经新闻", "行业活动", "市场动态", "专题报告", "党建专栏", "关于协会", "部委动态", "会员中心", "通知公告", "协会动态", "科技标准", "国际交流", "会展培训", "技术产品展示", "数据中心", "会议活动", "视频专栏", "焦点新闻", "环保装备高质量发展大会", "鼓励目录", "规范条件企业", "揭榜挂帅", "科技成果评价", "标准立项申报", "奖项申报服务", "年度服务计划", "协会简介", "组织架构", "协会章程", "联系我们", "申请入会", "会员权益", "政策咨询", "标准制修订", "大气污染防治装备分会", "水污染防治装备分会", "固体废物处理装备分会", "环境监测仪器专业委员会"}
	// Append new columns so existing preview URLs remain stable.
	names = append(names, "协会服务", "党建教育", "政策解读", "政策文件", "名单公示")
	names = append(names, "电子刊物")
	for i, name := range names {
		parentID := int64(0)
		switch name {
		case "政策解读", "政策文件", "名单公示":
			parentID = 8
		case "科技成果评价", "标准立项申报", "奖项申报服务", "年度服务计划":
			parentID = 40
		case "协会简介", "组织架构", "协会章程", "联系我们":
			parentID = 7
		case "申请入会", "会员权益", "政策咨询", "标准制修订":
			parentID = 9
		case "党建教育":
			parentID = 6
		}
		s.Columns = append(s.Columns, model.Column{ID: int64(i + 1), Name: name, PageID: 1, PageName: "环保机械协会", ParentID: parentID, Sort: i})
	}
	date := time.Date(2026, 7, 28, 9, 0, 0, 0, time.FixedZone("CST", 8*3600))
	for i := 1; i <= 101; i++ {
		c := int64(2 + (i-1)%4)
		a := model.Article{ID: int64(i), ColumnID: c, Type: 1, Title: "机械通用零部件行业 \"十五五\" 规划调研活动", Content: "<p>本条为静态预览演示内容，正式发布时由内容数据库提供正文。</p>", PublishTime: date.AddDate(0, 0, -i), Cover: "assets/hero.png", Source: "中国环保机械行业协会"}
		s.Items[c] = append(s.Items[c], a)
		s.Items[1] = append(s.Items[1], a)
	}
	for i, title := range fixture.Titles {
		d, _ := time.ParseInLocation("2006-01-02", fixture.Dates[i], date.Location())
		body := "<p>" + html.EscapeString(title) + "</p><p>本条为静态预览内容。</p>"
		if i == 0 {
			body = "<p>" + strings.Join(fixture.Paragraphs, "</p><p>") + "</p>"
		}
		a := model.Article{ID: int64(102 + i), ColumnID: 12, Type: 1, Title: title, Content: body, PublishTime: d, Cover: "assets/hero.png", Source: "中国环保机械行业协会"}
		s.Items[12] = append(s.Items[12], a)
	}
	var partyArticles []struct {
		ID, ColumnID                             int64
		Title, PublishedAt, Source, URL, Summary string
		Paragraphs                               []string
	}
	if err := json.Unmarshal(partyContent, &partyArticles); err != nil {
		panic(err)
	}
	for _, item := range partyArticles {
		published, err := time.Parse(time.RFC3339, item.PublishedAt)
		if err != nil {
			panic(err)
		}
		body := "<p>来源：" + html.EscapeString(item.Source) + "</p><h2>内容摘要</h2>"
		for _, paragraph := range item.Paragraphs {
			body += "<p>" + html.EscapeString(paragraph) + "</p>"
		}
		body += `<p><a href="` + html.EscapeString(item.URL) + `" target="_blank" rel="noopener noreferrer">阅读公众号原文（完整图文）</a></p>`
		s.Items[item.ColumnID] = append(s.Items[item.ColumnID], model.Article{
			ID: item.ID, ColumnID: item.ColumnID, Type: 1, Title: item.Title,
			Summary: item.Summary, Content: body, Source: item.Source, PublishTime: published,
		})
	}
	for _, id := range []int64{6, 41} {
		sort.SliceStable(s.Items[id], func(i, j int) bool {
			return s.Items[id][i].PublishTime.After(s.Items[id][j].PublishTime)
		})
	}
	for _, col := range s.Columns {
		if len(s.Items[col.ID]) > 0 {
			continue
		}
		if col.Name == "关于协会" || col.Name == "协会简介" {
			s.Items[col.ID] = []model.Article{{ID: 7001, ColumnID: 7, Type: 1, Title: "协会简介", Content: associationContent, PublishTime: date, Source: "中国环保机械行业协会"}}
			continue
		}
		for i := 0; i < 8; i++ {
			a := model.Article{ID: col.ID*1000 + int64(i+1), ColumnID: col.ID, Type: 1, Title: fmt.Sprintf("%s：环保装备行业交流与发展动态", col.Name), Summary: "关注环保装备产业发展，推动行业交流与技术创新。", Content: "<p>本条为静态预览演示内容，正式发布时由内容数据库提供正文。</p>", Cover: "assets/hero.png", PublishTime: date.AddDate(0, 0, -i), Source: "中国环保机械行业协会"}
			switch col.Name {
			case "通知公告":
				a.Title = []string{"关于赴丹麦开展水与环境技术考察交流活动的通知", "关于征集环保科技创新成果的通知", "《工业锅炉烟气多污染物协同治理技术规范》"}[i%3]
			case "会议活动":
				a.Title = "高质量环保装备助力生态文明建设"
				a.Cover = "assets/meeting.png"
			case "会展培训":
				a.Title = "高质量环保装备助力生态文明建设"
				a.Cover = "assets/training.png"
			case "视频专栏":
				a.Type = 2
				a.Title = fixture.Titles[0]
				a.Cover = "assets/video-poster.png"
				a.Content = ""
			case "焦点新闻":
				a.Title = []string{"高质量环保装备助力生态文明建设", "加快构建环保装备制造业发展新格局 助力生态文明建设", "推动环保装备行业交流合作 助力产业高质量发展"}[i%3]
				a.Cover = []string{"assets/hero-standard-release.png", "assets/hero-water-building.jpg", "assets/hero-industry-meeting.png"}[i%3]
			}
			s.Items[col.ID] = append(s.Items[col.ID], a)
		}
	}
	return s
}

func (s *Source) FetchPageColumns(ctx context.Context, _ string, parent int64) ([]model.Column, error) {
	var result []model.Column
	for _, c := range s.Columns {
		if c.ParentID == parent {
			result = append(result, c)
		}
	}
	return result, ctx.Err()
}
func (s *Source) FetchColumnArticles(ctx context.Context, id int64) (model.Column, []model.Article, error) {
	for _, c := range s.Columns {
		if c.ID == id {
			return c, s.Items[id], ctx.Err()
		}
	}
	return model.Column{}, nil, fmt.Errorf("unknown column %d", id)
}
func (s *Source) FetchAttachments(ctx context.Context, id int64) ([]model.Attachment, error) {
	return nil, ctx.Err()
}
