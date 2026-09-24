package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"camie-portal/backend/internal/config"
	"camie-portal/backend/internal/model"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

type Source interface {
	FetchPageColumns(context.Context, string, int64) ([]model.Column, error)
	FetchColumnArticles(context.Context, int64) (model.Column, []model.Article, error)
	FetchAttachments(context.Context, int64) ([]model.Attachment, error)
}
type Article struct {
	ID          int64              `json:"id"`
	Title       string             `json:"title"`
	Summary     string             `json:"summary"`
	Date        string             `json:"date"`
	Href        string             `json:"href"`
	Category    string             `json:"category"`
	Cover       string             `json:"cover"`
	Source      string             `json:"source"`
	Video       bool               `json:"video"`
	HasVideo    bool               `json:"-"`
	Content     template.HTML      `json:"-"`
	Attachments []model.Attachment `json:"-"`
}
type Link struct {
	Name, Href string
	Active     bool
	Number     int
}
type Group struct {
	Name, Href string
	Items      []*Article
	Feature    *Article
	Rest       []*Article
}
type Page struct {
	ActiveNav                             string
	Title, Kind, BodyClass, RootPrefix    string
	Items                                 []*Article
	Article, PreviousArticle, NextArticle *Article
	Parent                                *Link
	Breadcrumbs                           []Link
	Side, Pages                           []Link
	Hero                                  []*Article
	Notices, Topics, Cards                []Group
	DataCenterURL                         string
	Branches                              []string
	Total, PageSize, Page, TotalPages     int
	Previous, Next, BackHref, PagePrefix  string
}
type Result struct {
	Output      string    `json:"output"`
	Files       int       `json:"files"`
	Articles    int       `json:"articles"`
	Lists       int       `json:"lists"`
	GeneratedAt time.Time `json:"generated_at"`
}
type Generator struct {
	cfg    config.Config
	source Source
	mu     sync.Mutex
}
type catalog struct {
	columns []model.Column
	groups  map[int64][]*Article
	byID    map[int64]*Article
	byName  map[string]int64
	all     []*Article
	raw     map[int64]model.Article
}

func New(cfg config.Config, source Source) *Generator { return &Generator{cfg: cfg, source: source} }

func (g *Generator) load(ctx context.Context) (*catalog, error) {
	c := &catalog{groups: map[int64][]*Article{}, byID: map[int64]*Article{}, byName: map[string]int64{}, raw: map[int64]model.Article{}}
	queue := []int64{0}
	seen := map[int64]bool{}
	zone, _ := time.LoadLocation(g.cfg.Site.Timezone)
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		columns, err := g.source.FetchPageColumns(ctx, g.cfg.Site.PageName, parent)
		if err != nil {
			return nil, err
		}
		for _, col := range columns {
			if col.ID <= 0 || seen[col.ID] {
				return nil, fmt.Errorf("invalid or cyclic column %d", col.ID)
			}
			if _, exists := c.byName[col.Name]; exists {
				return nil, fmt.Errorf("duplicate column name %q", col.Name)
			}
			seen[col.ID] = true
			c.byName[col.Name] = col.ID
			c.columns = append(c.columns, col)
			queue = append(queue, col.ID)
			_, items, err := g.source.FetchColumnArticles(ctx, col.ID)
			if err != nil {
				return nil, err
			}
			ids := map[int64]bool{}
			for _, a := range items {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				if a.ID <= 0 || ids[a.ID] || !model.IsSupportedArticleType(a.Type) || a.PublishTime.IsZero() || a.PublishTime.After(time.Now()) {
					continue
				}
				ids[a.ID] = true
				a.PublishTime = a.PublishTime.In(zone)
				href := fmt.Sprintf("article/%s/%d.html", a.PublishTime.Format("2006/01"), a.ID)
				if a.URL != "" {
					u, err := url.Parse(a.URL)
					if err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" {
						href = u.String()
					}
				}
				cover := g.media(a.Cover)
				if cover == "" {
					cover = g.cfg.Site.FallbackCover
				}
				v := &Article{ID: a.ID, Title: a.Title, Summary: a.Summary, Category: col.Name, Source: a.Source, Date: a.PublishTime.Format("2006-01-02"), Href: href, Cover: cover, Video: a.Type == 2}
				if existing, ok := c.byID[a.ID]; ok {
					v = existing
				} else {
					attachments, err := g.source.FetchAttachments(ctx, a.ID)
					if err != nil {
						return nil, err
					}
					for _, attachment := range attachments {
						attachment.URL = g.media(attachment.URL)
						if attachment.URL != "" {
							v.Attachments = append(v.Attachments, attachment)
						}
					}
					c.byID[a.ID] = v
					c.raw[a.ID] = a
					c.all = append(c.all, v)
				}
				c.groups[col.ID] = append(c.groups[col.ID], v)
			}
		}
	}
	sort.SliceStable(c.all, func(i, j int) bool {
		a, b := c.raw[c.all[i].ID], c.raw[c.all[j].ID]
		if a.IsTop != b.IsTop {
			return a.IsTop
		}
		if !a.PublishTime.Equal(b.PublishTime) {
			return a.PublishTime.After(b.PublishTime)
		}
		return a.ID > b.ID
	})
	return c, nil
}
func root(p Page, raw string) string {
	if raw == "" || strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "#") {
		return raw
	}
	u, err := url.Parse(raw)
	if err == nil && u.IsAbs() {
		return raw
	}
	return p.RootPrefix + strings.TrimPrefix(raw, "./")
}
func (g *Generator) media(raw string) string {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if u.IsAbs() {
		if (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
			return raw
		}
		return ""
	}
	if strings.HasPrefix(raw, "//") {
		return ""
	}
	clean := strings.TrimPrefix(raw, "./")
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return ""
	}
	// Project-owned graphics always stay local, even when an upload host is configured.
	if strings.HasPrefix(clean, "assets/") || strings.HasPrefix(clean, "slice/") {
		return clean
	}
	if g.cfg.Site.MediaBaseURL != "" {
		return strings.TrimRight(g.cfg.Site.MediaBaseURL, "/") + "/" + strings.TrimLeft(clean, "/")
	}
	return clean
}
func (g *Generator) content(raw, prefix string) template.HTML {
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("video", "source")
	policy.AllowAttrs("src", "poster", "controls", "preload", "width", "height").OnElements("video")
	policy.AllowAttrs("src", "type").OnElements("source")
	policy.AllowRelativeURLs(true)
	policy.AllowURLSchemes("http", "https")
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return ""
	}
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		for i, a := range n.Attr {
			if a.Key == "src" || a.Key == "poster" || a.Key == "href" {
				if strings.HasPrefix(a.Val, "#") {
					continue
				}
				media := g.media(a.Val)
				n.Attr[i].Val = root(Page{RootPrefix: prefix}, media)
			}
		}
		if n.Type == html.ElementNode && n.Data == "video" {
			n.Attr = append(n.Attr, html.Attribute{Key: "controls", Val: "controls"}, html.Attribute{Key: "preload", Val: "metadata"})
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(doc)
	var b bytes.Buffer
	_ = html.Render(&b, doc)
	return template.HTML(policy.Sanitize(b.String()))
}
func group(c *catalog, name string, limit int) Group {
	id := c.byName[name]
	items := c.groups[id]
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	href := "news.html"
	if id > 0 {
		href = fmt.Sprintf("list/%d/1.html", id)
	}
	out := Group{Name: name, Href: href, Items: items}
	if len(items) > 0 {
		out.Feature = items[0]
		out.Rest = items[1:]
	}
	return out
}
func side(c *catalog, current string) []Link {
	names := []string{"行业资讯", "产经新闻", "行业活动", "市场动态", "专题报告"}
	switch current {
	case "党建专栏", "党建栏目", "党建教育":
		party := group(c, "党建栏目", 0)
		if c.byName["党建栏目"] == 0 {
			party = group(c, "党建专栏", 0)
		}
		return []Link{
			{Name: "党建栏目", Href: party.Href, Active: current != "党建教育"},
			{Name: "党建教育", Href: group(c, "党建教育", 0).Href, Active: current == "党建教育"},
		}
	case "协会服务", "科技成果评价", "标准立项申报", "奖项申报服务", "年度服务计划":
		names = []string{"协会服务", "科技成果评价", "标准立项申报", "奖项申报服务", "年度服务计划"}
	case "部委动态", "政策解读", "政策文件", "通知公告", "名单公示":
		names = []string{"政策解读", "政策文件", "通知公告", "名单公示"}
	case "会员中心", "申请入会", "会员权益", "政策咨询", "标准制修订":
		names = []string{"会员中心", "申请入会", "会员权益", "政策咨询", "标准制修订"}
	case "关于协会", "协会简介", "组织架构", "协会章程", "联系我们":
		names = []string{"协会简介", "组织架构", "协会章程", "联系我们"}
		if current == "关于协会" {
			current = "协会简介"
		}
	}
	if current != "" {
		found := false
		for _, n := range names {
			if n == current {
				found = true
			}
		}
		if !found {
			names = append([]string{current}, names...)
		}
	}
	links := []Link{}
	for _, name := range names {
		grp := group(c, name, 0)
		links = append(links, Link{Name: name, Href: grp.Href, Active: name == current})
	}
	return links
}

func activeNav(c *catalog, p Page) string {
	if p.Kind == "home" {
		return "首页"
	}
	if p.Kind == "about" {
		return "关于协会"
	}
	name := p.Title
	if p.Article != nil {
		name = p.Article.Category
	}
	// Real content columns inherit the top-level navigation from their ancestry.
	columns := map[int64]model.Column{}
	for _, col := range c.columns {
		columns[col.ID] = col
	}
	branch := []string{name}
	for id := c.byName[name]; id != 0; {
		col, ok := columns[id]
		if !ok {
			break
		}
		branch = append(branch, col.Name)
		id = col.ParentID
	}
	for i := len(branch) - 1; i >= 0; i-- {
		switch branch[i] {
		case "党建专栏", "关于协会", "部委动态", "会员中心", "协会服务", "行业资讯":
			return branch[i]
		}
	}
	// Flat preview/legacy columns use the same grouping as the site's entry links.
	switch name {
	case "党建栏目", "党建教育":
		return "党建专栏"
	case "协会简介", "组织架构", "协会章程", "联系我们":
		return "关于协会"
	case "政策解读", "政策文件", "通知公告", "名单公示":
		return "部委动态"
	case "申请入会", "会员权益", "政策咨询", "标准制修订":
		return "会员中心"
	case "科技成果评价", "标准立项申报", "奖项申报服务", "年度服务计划":
		return "协会服务"
	default:
		return "行业资讯"
	}
}

func (g *Generator) GenerateSite(ctx context.Context) (result Result, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err = ctx.Err(); err != nil {
		return result, err
	}
	target := g.cfg.Site.DistRoot
	if err = config.ValidateOutput(g.cfg.Site.SourceRoot, target); err != nil {
		return result, err
	}
	if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return result, err
	}
	// No stale timeout: a long-running live publisher must never lose its lock.
	lock, err := os.OpenFile(target+".lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return result, fmt.Errorf("publisher is locked (check %s.lock): %w", target, err)
	}
	_, _ = fmt.Fprintf(lock, "pid=%d\nstarted=%s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	_ = lock.Close()
	defer os.Remove(target + ".lock")
	c, err := g.load(ctx)
	if err != nil {
		return result, err
	}
	stage, err := os.MkdirTemp(filepath.Dir(target), ".camie-stage-")
	if err != nil {
		return result, err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	if err = copyAssets(g.cfg.Site.SourceRoot, stage); err != nil {
		return result, err
	}
	funcs := template.FuncMap{
		"root": root, "add": func(a, b int) int { return a + b },
		"columnURL": func(p Page, name string) string { return root(p, group(c, name, 0).Href) },
		"articleURL": func(p Page, id int) string {
			if a := c.byID[int64(id)]; a != nil {
				return root(p, a.Href)
			}
			return root(p, "news.html")
		},
	}
	templates, err := template.New("portal").Funcs(funcs).ParseGlob(filepath.Join(g.cfg.Site.TemplateRoot, "*.tmpl"))
	if err != nil {
		return result, err
	}
	render := func(file, kind string, p Page) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		p.ActiveNav = activeNav(c, p)
		currentColumn := p.Title
		if p.Article != nil {
			currentColumn = p.Article.Category
		}
		if p.ActiveNav != "" && p.ActiveNav != "首页" && p.ActiveNav != currentColumn {
			parent := Link{Name: p.ActiveNav, Href: group(c, p.ActiveNav, 0).Href}
			p.Parent = &parent
		}
		// Build the complete root-to-leaf column chain so second-, third-, and
		// deeper-level pages keep every breadcrumb segment. Flat legacy/demo
		// columns fall back to the navigation group used elsewhere on the site.
		columns := make(map[int64]model.Column, len(c.columns))
		for _, col := range c.columns {
			columns[col.ID] = col
		}
		var ancestors []Link
		if id := c.byName[currentColumn]; id != 0 {
			for parentID := columns[id].ParentID; parentID != 0; {
				col, ok := columns[parentID]
				if !ok {
					break
				}
				ancestors = append(ancestors, Link{Name: col.Name, Href: group(c, col.Name, 0).Href})
				parentID = col.ParentID
			}
			for left, right := 0, len(ancestors)-1; left < right; left, right = left+1, right-1 {
				ancestors[left], ancestors[right] = ancestors[right], ancestors[left]
			}
		}
		if len(ancestors) == 0 && p.ActiveNav != "" && p.ActiveNav != "首页" && p.ActiveNav != currentColumn {
			ancestors = append(ancestors, Link{Name: p.ActiveNav, Href: group(c, p.ActiveNav, 0).Href})
		}
		p.Breadcrumbs = ancestors
		if p.ActiveNav == "党建专栏" {
			p.BodyClass += " party-page"
		}
		p.RootPrefix = strings.Repeat("../", strings.Count(file, "/"))
		if p.RootPrefix == "" {
			p.RootPrefix = "./"
		}
		if p.Article != nil {
			v := *p.Article
			v.Content = g.content(c.raw[v.ID].Content, p.RootPrefix)
			v.HasVideo = strings.Contains(string(v.Content), "<video")
			p.Article = &v
		}
		var b bytes.Buffer
		if err := templates.ExecuteTemplate(&b, kind, p); err != nil {
			return fmt.Errorf("render %s: %w", file, err)
		}
		result.Files++
		return writeFile(stage, file, b.Bytes())
	}
	home := Page{Title: "首页", Kind: "home", BodyClass: "home", Hero: group(c, g.cfg.Site.HeroColumn, 3).Items, Branches: []string{"大气污染防治装备分会", "水污染防治装备分会", "固体废物处理装备分会", "环境监测仪器专业委员会"}}
	for _, n := range []string{"通知公告", "协会动态"} {
		home.Notices = append(home.Notices, group(c, n, 6))
	}
	for _, n := range []string{"科技标准", "国际交流", "会展培训", "技术产品展示", "电子刊物"} {
		home.Topics = append(home.Topics, group(c, n, 8))
	}
	dataCenter := group(c, "数据中心", 1)
	home.DataCenterURL = dataCenter.Href
	if dataCenter.Feature != nil {
		home.DataCenterURL = dataCenter.Feature.Href
	}
	for _, n := range []string{"会议活动", "会展培训", "视频专栏"} {
		home.Cards = append(home.Cards, group(c, n, 3))
	}
	if err = render("index.html", "home", home); err != nil {
		return result, err
	}
	makeLists := func(name, dir, alias string, items []*Article) error {
		totalPages := (len(items) + g.cfg.Site.PageSize - 1) / g.cfg.Site.PageSize
		if totalPages < 1 {
			totalPages = 1
		}
		for pageNum := 1; pageNum <= totalPages; pageNum++ {
			start := (pageNum - 1) * g.cfg.Site.PageSize
			end := start + g.cfg.Site.PageSize
			if end > len(items) {
				end = len(items)
			}
			p := Page{Title: name, Kind: "list", BodyClass: "list-page", Items: items[start:end], Side: side(c, name), Total: len(items), PageSize: g.cfg.Site.PageSize, Page: pageNum, TotalPages: totalPages, PagePrefix: dir + "/"}
			if pageNum > 1 {
				p.Previous = fmt.Sprintf("%s/%d.html", dir, pageNum-1)
			}
			if pageNum < totalPages {
				p.Next = fmt.Sprintf("%s/%d.html", dir, pageNum+1)
			}
			// Bounded page controls; all pages remain reachable through previous/next links.
			for n := 1; n <= totalPages; n++ {
				if n == 1 || n == totalPages || (n >= pageNum-2 && n <= pageNum+2) {
					p.Pages = append(p.Pages, Link{Number: n, Href: fmt.Sprintf("%s/%d.html", dir, n), Active: n == pageNum})
				}
			}
			if err := render(fmt.Sprintf("%s/%d.html", dir, pageNum), "list", p); err != nil {
				return err
			}
			result.Lists++
			if pageNum == 1 && alias != "" {
				if err := render(alias, "list", p); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, col := range c.columns {
		if col.Name == "关于协会" || col.Name == "协会简介" {
			p := Page{Title: "协会简介", Kind: "about", BodyClass: "about-page", Side: side(c, col.Name)}
			items := c.groups[col.ID]
			if col.Name == "关于协会" && len(c.groups[c.byName["协会简介"]]) > 0 {
				items = c.groups[c.byName["协会简介"]]
			}
			for _, a := range items {
				if !a.Video && !strings.HasPrefix(a.Href, "http") {
					p.Article = a
					break
				}
			}
			if err = render(fmt.Sprintf("list/%d/1.html", col.ID), "about", p); err != nil {
				return result, err
			}
			result.Lists++
			continue
		}
		if err = makeLists(col.Name, fmt.Sprintf("list/%d", col.ID), "", c.groups[col.ID]); err != nil {
			return result, err
		}
	}
	news := group(c, "行业资讯", 0).Items
	if len(news) == 0 {
		for _, a := range c.all {
			if !a.Video {
				news = append(news, a)
			}
		}
	}
	var videos []*Article
	for _, a := range c.all {
		if a.Video {
			videos = append(videos, a)
		}
	}
	if err = makeLists("行业资讯", "list/news", "news.html", news); err != nil {
		return result, err
	}
	if err = makeLists("视频专栏", "list/videos", "videos.html", videos); err != nil {
		return result, err
	}
	firstNews := (*Article)(nil)
	firstVideo := (*Article)(nil)
	for _, a := range c.all {
		if strings.HasPrefix(a.Href, "http") {
			continue
		}
		p := Page{Title: a.Title, Kind: "article", BodyClass: "article-page", Article: a, Side: side(c, a.Category), BackHref: group(c, a.Category, 0).Href}
		if a.Video {
			p.Kind = "video"
			p.BodyClass = "video-page"
			if firstVideo == nil {
				firstVideo = a
			}
		} else if firstNews == nil {
			firstNews = a
		}
		siblings := c.groups[c.byName[a.Category]]
		for i, v := range siblings {
			if v.ID == a.ID {
				if i > 0 {
					p.PreviousArticle = siblings[i-1]
				}
				if i+1 < len(siblings) {
					p.NextArticle = siblings[i+1]
				}
				break
			}
		}
		if err = render(a.Href, "article", p); err != nil {
			return result, err
		}
		result.Articles++
	}
	if a := c.byID[102]; a != nil && !a.Video && !strings.HasPrefix(a.Href, "http") {
		firstNews = a
	}
	for _, alias := range []struct {
		file, kind string
		a          *Article
	}{{"detail.html", "article", firstNews}, {"video-detail.html", "video", firstVideo}} {
		p := Page{Title: "内容详情", Kind: alias.kind, BodyClass: "article-page", Article: alias.a, Side: side(c, "行业资讯")}
		kind := "article"
		if alias.a == nil {
			kind = "empty-detail"
		} else {
			p.Title = alias.a.Title
			p.BackHref = group(c, alias.a.Category, 0).Href
			if alias.a.Video {
				p.BodyClass = "video-page"
			}
		}
		if err = render(alias.file, kind, p); err != nil {
			return result, err
		}
	}
	if err = render("search.html", "search", Page{Title: "搜索资讯", Kind: "search", BodyClass: "list-page", Side: side(c, "行业资讯")}); err != nil {
		return result, err
	}
	paths := map[string]string{}
	columns := map[string]string{}
	for _, a := range c.all {
		paths[strconv.FormatInt(a.ID, 10)] = a.Href
	}
	for name := range c.byName {
		columns[name] = group(c, name, 0).Href
	}
	content := struct {
		Generated    bool              `json:"generated"`
		Articles     []*Article        `json:"articles"`
		ArticlePaths map[string]string `json:"articlePaths"`
		Columns      map[string]string `json:"columns"`
		DefaultVideo string            `json:"defaultVideo"`
	}{true, c.all, paths, columns, "video-detail.html"}
	if firstVideo != nil {
		content.DefaultVideo = firstVideo.Href
	}
	payload, err := json.Marshal(content)
	if err != nil {
		return result, err
	}
	if err = writeFile(stage, "generated-content.js", append([]byte("window.CAMIE_STATIC_CONTENT = "), append(payload, []byte(";\n")...)...)); err != nil {
		return result, err
	}
	result.Files++
	if err = validateSite(stage); err != nil {
		return result, err
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if err = publishDirectory(g.cfg.Site.SourceRoot, stage, target); err != nil {
		return result, err
	}
	result.Output = target
	result.GeneratedAt = time.Now()
	return result, nil
}
