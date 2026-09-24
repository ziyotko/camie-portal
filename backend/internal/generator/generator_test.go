package generator

import (
	"camie-portal/backend/internal/config"
	"camie-portal/backend/internal/demo"
	"camie-portal/backend/internal/model"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testConfig(t *testing.T) config.Config {
	t.Helper()
	cfg, err := config.Load("../../config.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	source := t.TempDir()
	if err = copyAssets(cfg.Site.SourceRoot, source); err != nil {
		t.Fatal(err)
	}
	cfg.Site.SourceRoot = source
	cfg.Site.DistRoot = filepath.Join(source, "dist", "test-site")
	return cfg
}
func read(t *testing.T, root, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
func TestCompleteSiteContainsStaticContentAndWorkingLinks(t *testing.T) {
	cfg := testConfig(t)
	s := demo.NewSource()
	g := New(cfg, s)
	result, err := g.GenerateSite(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Articles < 100 || result.Lists < 10 {
		t.Fatalf("incomplete site: %+v", result)
	}
	home := read(t, result.Output, "index.html")
	for _, text := range []string{"加快构建环保装备制造业发展新格局", "通知公告", "科技标准", "电子刊物", "数据中心", "协会服务", "会员中心", "我的账号", "环保装备高质量发展大会", "视频专栏"} {
		if !strings.Contains(home, text) {
			t.Fatalf("missing server rendered content %s", text)
		}
	}
	if strings.Count(home, `data-event-panel=`) != 3 || !strings.Contains(home, "assets/event-banner-tech.png") || !strings.Contains(home, "assets/event-banner-talent.png") {
		t.Fatal("event banner carousel is incomplete")
	}
	if !strings.Contains(home, `data-hero-carousel`) || strings.Count(home, `data-hero-step=`) != 2 {
		t.Fatal("hero carousel controls are incomplete")
	}
	for _, tab := range []string{
		`<a role="tab" aria-selected="true" data-notice="0" class="active" href="./list/10/1.html">通知公告</a>`,
		`data-notice="1" class="" href="./list/11/1.html">协会动态</a>`,
		`<a role="tab" aria-selected="true" data-tech="0" class="active" href="./list/12/1.html">科技标准</a>`,
		`data-tech="4" class="" href="./list/45/1.html">电子刊物</a>`,
	} {
		if !strings.Contains(home, tab) {
			t.Fatalf("hover tab must remain a column link: %s", tab)
		}
	}
	dataCenter := `class="section-tab-link" href="./article/2026/07/16001.html">数据中心</a>`
	if !strings.Contains(home, dataCenter) || strings.Contains(home, `data-tech="5"`) {
		t.Fatal("data center must link directly to its detail instead of acting as a tab")
	}
	for _, service := range []string{"科技成果评价", "标准立项申报", "奖项申报服务", "年度服务计划"} {
		if !strings.Contains(home, ">"+service+"</a>") {
			t.Fatalf("home missing service %s", service)
		}
	}
	for _, accountLink := range []string{
		`href="https://demo.miic.com.cn/business_member/register" target="_blank" rel="noopener noreferrer">`,
		`href="https://demo.miic.com.cn/business_member/login" target="_blank" rel="noopener noreferrer">`,
	} {
		if !strings.Contains(home, accountLink) {
			t.Fatalf("home missing member account link %s", accountLink)
		}
	}
	if !strings.Contains(home, `class="info-card video-column"`) || !strings.Contains(home, `<h3>视频专栏</h3>`) {
		t.Fatal("video column title and single-line styling hook are missing")
	}
	detail := read(t, result.Output, "article/2026/07/102.html")
	if !strings.Contains(detail, "为响应国家新能源绿色低碳发展战略") || !strings.Contains(detail, "../../../styles.css") {
		t.Fatal("detail is not statically rendered or nested assets are incorrect")
	}
	if strings.Contains(home, `href="#/`) || strings.Contains(detail, `href="#/`) {
		t.Fatal("SPA navigation survived")
	}
	news := read(t, result.Output, "news.html")
	if strings.Count(news, `class="news-row"`) != 14 {
		t.Fatal("wrong page size")
	}
	services := read(t, result.Output, "list/40/1.html")
	if strings.Count(services, `class="news-row"`) != 8 || !strings.Contains(services, `aria-label="分页"`) || !strings.Contains(services, "科技成果评价") {
		t.Fatal("services must render a static list with pagination and service categories")
	}
	for file, want := range map[string]string{
		"index.html": "首页", "news.html": "行业资讯", "list/news/2.html": "行业资讯",
		"list/6/1.html": "党建专栏", "article/2026/07/6001.html": "党建专栏",
		"list/41/1.html": "党建专栏", "article/2026/09/41001.html": "党建专栏",
		"list/7/1.html": "关于协会", "list/28/1.html": "关于协会", "article/2026/07/7001.html": "关于协会",
		"list/8/1.html": "部委动态", "article/2026/07/8001.html": "部委动态",
		"list/9/1.html": "会员中心", "article/2026/07/9001.html": "会员中心",
		"list/24/1.html": "协会服务", "article/2026/07/24001.html": "协会服务",
		"list/40/1.html": "协会服务", "article/2026/07/40001.html": "协会服务",
		"list/32/1.html": "会员中心", "search.html": "行业资讯", "detail.html": "行业资讯",
	} {
		data := read(t, result.Output, file)
		if strings.Contains(data, " party-page\"") != (want == "党建专栏") {
			t.Fatalf("%s party theme must follow the active navigation", file)
		}
		if want == "党建专栏" {
			sidebar := data[strings.Index(data, `<aside class="side-nav"`):]
			sidebar = sidebar[:strings.Index(sidebar, "</aside>")]
			if strings.Count(sidebar, "<a ") != 2 || !strings.Contains(sidebar, `list/6/1.html"`) || !strings.Contains(sidebar, `list/41/1.html"`) || strings.Index(sidebar, ">党建栏目</a>") > strings.Index(sidebar, ">党建教育</a>") {
				t.Fatalf("%s must have only the two party sidebar columns", file)
			}
			label := "党建栏目"
			if strings.Contains(file, "/41/") || strings.Contains(file, "/41001.html") {
				label = "党建教育"
			}
			if strings.Count(sidebar, `class="active"`) != 1 || !strings.Contains(sidebar, `aria-current="page">`+label+`</a>`) {
				t.Fatalf("%s must select the current party sidebar column", file)
			}
		}
		nav := data[strings.Index(data, `aria-label="主导航"`):]
		nav = nav[:strings.Index(nav, "</nav>")]
		if strings.Count(nav, `class="active"`) != 1 || strings.Count(nav, `aria-current="page"`) != 1 || !strings.Contains(nav, `data-nav="`+want+`" class="active" aria-current="page"`) {
			t.Fatalf("%s must highlight only %s: %s", file, want, nav)
		}
		if !strings.Contains(nav, `list/40/1.html" data-nav="协会服务"`) || strings.Contains(nav, "index.html#services") {
			t.Fatalf("%s services navigation must point to its column list", file)
		}
	}
	for _, file := range []string{"list/7/1.html", "list/28/1.html"} {
		about := read(t, result.Output, file)
		for _, want := range []string{"1994年成立", "现有会员单位480家", "协会是联系政府与企业的桥梁和纽带", "../../assets/association-honors.jpg", "北京市西城区月坛南街26号院", "100825", "camie1994@163.com", "010-68780110"} {
			if !strings.Contains(about, want) {
				t.Fatalf("%s missing introduction content: %s", file, want)
			}
		}
		if strings.Contains(about, `class="news-row"`) || strings.Contains(about, `aria-label="分页"`) {
			t.Fatalf("%s still renders a news list", file)
		}
	}
	organization := read(t, result.Output, "list/29/1.html")
	if !strings.Contains(organization, `>首页</a><span class="breadcrumb-separator" aria-hidden="true">›</span><a href="../../list/7/1.html">关于协会</a><span class="breadcrumb-separator" aria-hidden="true">›</span><span aria-current="page">组织架构</span>`) {
		t.Fatal("about child breadcrumb must retain its top-level parent")
	}
	ministry := read(t, result.Output, "list/8/1.html")
	for _, label := range []string{"政策解读", "政策文件", "通知公告", "名单公示"} {
		if !strings.Contains(ministry, ">"+label+"</a>") {
			t.Fatalf("ministry sidebar missing %s", label)
		}
	}
	policy := read(t, result.Output, "list/42/1.html")
	if !strings.Contains(policy, `>首页</a><span class="breadcrumb-separator" aria-hidden="true">›</span><a href="../../list/8/1.html">部委动态</a><span class="breadcrumb-separator" aria-hidden="true">›</span><span aria-current="page">政策解读</span>`) {
		t.Fatal("ministry child breadcrumb must retain its top-level parent")
	}
	members := read(t, result.Output, "list/9/1.html")
	memberSidebar := members[strings.Index(members, `<aside class="side-nav"`):]
	memberSidebar = memberSidebar[:strings.Index(memberSidebar, "</aside>")]
	for _, label := range []string{"会员中心", "申请入会", "会员权益", "政策咨询", "标准制修订"} {
		if !strings.Contains(memberSidebar, ">"+label+"</a>") {
			t.Fatalf("member sidebar missing %s", label)
		}
	}
	for _, label := range []string{"行业资讯", "产经新闻", "行业活动", "市场动态", "专题报告"} {
		if strings.Contains(memberSidebar, ">"+label+"</a>") {
			t.Fatalf("member sidebar must not link to industry column %s", label)
		}
	}
	join := read(t, result.Output, "list/32/1.html")
	if !strings.Contains(join, `>首页</a><span class="breadcrumb-separator" aria-hidden="true">›</span><a href="../../list/9/1.html">会员中心</a><span class="breadcrumb-separator" aria-hidden="true">›</span><span aria-current="page">申请入会</span>`) {
		t.Fatal("member child breadcrumb must retain the member center parent")
	}
	joinSidebar := join[strings.Index(join, `<aside class="side-nav"`):]
	joinSidebar = joinSidebar[:strings.Index(joinSidebar, "</aside>")]
	if !strings.Contains(joinSidebar, `>会员中心</a>`) || !strings.Contains(joinSidebar, `class="active" aria-current="page">申请入会</a>`) {
		t.Fatal("member child sidebar must retain the member center and highlight the current child")
	}
	for _, file := range []string{"index.html", "news.html", "videos.html", "detail.html", "video-detail.html", "list/7/1.html", "list/28/1.html"} {
		data := read(t, result.Output, file)
		nav := data[strings.Index(data, `aria-label="主导航"`):]
		nav = nav[:strings.Index(nav, "</nav>")]
		labels := []string{"首页", "党建专栏", "关于协会", "部委动态", "会员中心", "协会服务", "行业资讯"}
		previous := -1
		for _, label := range labels {
			index := strings.Index(nav, ">"+label+"</a>")
			if label == "党建专栏" {
				index = strings.Index(nav, "党建专栏</a>")
			}
			if index <= previous {
				t.Fatalf("navigation changed in %s: %s", file, label)
			}
			previous = index
		}
	}
	if err = validateSite(result.Output); err != nil {
		t.Fatal(err)
	}
}
func TestRebuildRemovesUnpublishedArticleAndKeepsPreviousRelease(t *testing.T) {
	cfg := testConfig(t)
	s := demo.NewSource()
	g := New(cfg, s)
	if _, err := g.GenerateSite(context.Background()); err != nil {
		t.Fatal(err)
	}
	s.Items[12] = s.Items[12][1:]
	if _, err := g.GenerateSite(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cfg.Site.DistRoot, "article/2026/07/102.html")); !os.IsNotExist(err) {
		t.Fatal("withdrawn article remains in release")
	}
	if strings.Contains(read(t, cfg.Site.DistRoot, "generated-content.js"), `"102":`) {
		t.Fatal("stale article mapping")
	}
	if _, err := os.Stat(filepath.Join(cfg.Site.DistRoot+".previous", "article/2026/07/102.html")); err != nil {
		t.Fatal("previous release missing")
	}
}

type failSource struct{ *demo.Source }

func (f failSource) FetchColumnArticles(context.Context, int64) (model.Column, []model.Article, error) {
	return model.Column{}, nil, errors.New("database unavailable")
}
func TestFailureAndCancellationPreserveLiveFiles(t *testing.T) {
	cfg := testConfig(t)
	s := demo.NewSource()
	if _, err := New(cfg, s).GenerateSite(context.Background()); err != nil {
		t.Fatal(err)
	}
	before := read(t, cfg.Site.DistRoot, "index.html")
	if _, err := New(cfg, failSource{s}).GenerateSite(context.Background()); err == nil {
		t.Fatal("expected load failure")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := New(cfg, s).GenerateSite(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if read(t, cfg.Site.DistRoot, "index.html") != before {
		t.Fatal("failed publication modified live site")
	}
}
func TestRichTextSanitizationAndMediaPrefix(t *testing.T) {
	cfg := testConfig(t)
	cfg.Site.MediaBaseURL = "https://media.example.test"
	g := New(cfg, demo.NewSource())
	raw := `<script>alert(1)</script><p onclick="bad()">正文</p><img src="uploads/a.png"><a href="javascript:alert(1)">bad</a><video src="uploads/a.mp4"></video><img src="assets/hero.png">`
	clean := string(g.content(raw, "../../../"))
	for _, bad := range []string{"<script", "onclick", "javascript:"} {
		if strings.Contains(clean, bad) {
			t.Fatal("unsafe content survived", clean)
		}
	}
	for _, want := range []string{"https://media.example.test/uploads/a.png", "https://media.example.test/uploads/a.mp4", "../../../assets/hero.png", "controls"} {
		if !strings.Contains(clean, want) {
			t.Fatal("missing media rewrite", clean)
		}
	}
}
func TestFutureArticlesAndDuplicateColumnNames(t *testing.T) {
	cfg := testConfig(t)
	s := demo.NewSource()
	item := s.Items[12][0]
	item.ID = 999999
	item.PublishTime = time.Now().Add(time.Hour)
	s.Items[12] = append(s.Items[12], item)
	catalog, err := New(cfg, s).load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if catalog.byID[item.ID] != nil {
		t.Fatal("future article published")
	}
	s.Columns[1].Name = s.Columns[0].Name
	if _, err = New(cfg, s).load(context.Background()); err == nil {
		t.Fatal("duplicate names accepted")
	}
}
