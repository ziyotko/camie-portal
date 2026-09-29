import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = dirname(fileURLToPath(import.meta.url));
const pagesDir = join(root, 'pages');
mkdirSync(pagesDir, { recursive: true });

const columnStructure = JSON.parse(readFileSync(join(root, 'column-structure.json'), 'utf8'));
const columnPaths = {};
const collectColumnPaths = (columns, parents = []) => columns.forEach((column) => {
  const path = [...parents, column.name];
  columnPaths[column.code] = path;
  if (column.children) collectColumnPaths(column.children, path);
});
collectColumnPaths(columnStructure.navigation);
const columnPathsJSON = JSON.stringify(columnPaths).replaceAll('<', '\\u003c');

const nav = [
  ['home', '首页', '../index.html'],
  ['party', '党建专栏', 'party.html'],
  ['ministry', '部委动态', 'ministry.html'],
  ['news', '新闻中心', 'news.html'],
  ['training', '交流培训', 'training.html'],
  ['standards', '科技标准', 'standards.html'],
  ['about', '关于协会', 'about.html'],
];

const navSubmenus = {
  party: [
    ['党建要闻', 'party.html#party-news'],
    ['学习教育', 'party.html#study'],
    ['支部活动', 'party.html#branch-activity'],
  ],
  ministry: [
    ['工作动态', 'ministry.html#work'],
    ['政策文件', 'ministry.html#policy-files'],
    ['政策解读', 'ministry.html#interpretation'],
  ],
  news: [
    ['热点关注', 'news.html#hot'],
    ['通知公告', 'news.html#notice'],
    ['协会动态', 'news.html#association'],
    ['会员动态', 'news.html#member-news'],
  ],
  training: [
    ['会议活动', 'training.html#meetings'],
    ['人才培训', 'training.html#talent'],
    ['国际交流与合作', 'training.html#international'],
  ],
  standards: [
    ['科技创新及成果转化', 'standards.html#innovation'],
    ['标准工作', 'standards.html#standard-work'],
    ['绿色技术推广', 'standards.html#green-promotion'],
  ],
  about: [
    ['协会简介', 'about.html#introduction'],
    ['组织架构', 'about.html#organization'],
    ['协会章程', 'about.html#charter'],
    ['分支机构', 'about.html#branches'],
    ['专家委员会', 'experts.html#experts'],
    ['公开公示', 'about.html#disclosure'],
    ['联系我们', 'about.html#contact'],
  ],
};

const friends = [
  '中华人民共和国国家发展和改革委员会', '中共中央社会工作部', '中华人民共和国民政部',
  '中华人民共和国工业和信息化部', '中华人民共和国生态环境部', '中华人民共和国科学技术部',
  '国家知识产权局', '中国机经网',
];

function header(active = 'home', isHome = false) {
  const pagePrefix = isHome ? 'pages/' : '';
  const homeHref = isHome ? 'index.html' : '../index.html';
  const searchHref = `${pagePrefix}search.html`;
  return `
  <header class="site-header">
    <div class="header-main">
      <div class="header-inner">
        <a class="brand" href="${homeHref}" aria-label="中国环保机械行业协会首页"><img src="${isHome ? '' : '../'}assets/images/brand-logo.png" alt="中国环保机械行业协会 CAMIE"></a>
        <img class="brand-slogan" src="${isHome ? '' : '../'}assets/images/brand-slogan.png" alt="企业与政府沟通的桥梁">
        <div class="header-tools">
          <form class="search-form" role="search" action="${searchHref}" method="get"><label class="sr-only" for="search-${active}">站内搜索</label><input id="search-${active}" name="q" type="search" placeholder="请输入关键词" autocomplete="off"><button type="submit">搜索 <span class="search-icon" aria-hidden="true"></span></button></form>
          <button class="language-switch" type="button">中 | EN</button>
        </div>
        <button class="mobile-menu-button" type="button" aria-label="打开导航" aria-expanded="false"><span></span></button>
      </div>
    </div>
    <nav class="nav-bar" aria-label="主导航"><div class="nav-inner">
      ${nav.map(([id, label, href]) => {
        const finalHref = id === 'home' ? homeHref : `${pagePrefix}${href}`;
        const submenuItems = navSubmenus[id] || [];
        const submenu = submenuItems.length
          ? `<div class="nav-submenu">${submenuItems.map(([itemLabel, itemHref]) => `<a href="${pagePrefix}${itemHref}">${itemLabel}</a>`).join('')}</div>`
          : '';
        return `<div class="nav-item ${active === id ? 'active' : ''}"><a class="nav-link" href="${finalHref}"${submenuItems.length ? ' aria-haspopup="true"' : ''}>${label}</a>${submenu}</div>`;
      }).join('')}
    </div></nav>
  </header>`;
}

function footer(isHome = false) {
  const p = isHome ? '' : '../';
  return `
  <footer class="site-footer"><div class="footer-inner">
    <div class="friend-title">友情链接：</div>
    <div class="friend-links">${friends.map((item) => `<a href="#" data-demo-link>${item}</a>`).join('')}</div>
    <div class="copyright">© 2026 中国环保机械行业协会 版权所有　 京 ICP 备 12345678 号</div>
  </div></footer>
  <div class="toast" role="status" aria-live="polite"></div>
  <button class="back-to-top" type="button" aria-label="返回顶部">↑</button>
  <script id="column-paths" type="application/json">${columnPathsJSON}</script>
  <script src="${p}js/main.js"></script>`;
}

function documentPage({ title, active, bodyClass = '', content, isHome = false }) {
  const p = isHome ? '' : '../';
  return `<!doctype html>
<html lang="zh-CN"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="description" content="中国环保机械行业协会静态门户演示"><title>${title} - 中国环保机械行业协会</title><link rel="stylesheet" href="${p}css/common.css">${isHome ? '<link rel="stylesheet" href="css/home.css">' : ''}</head>
<body class="${bodyClass}">${header(active, isHome)}${content}${footer(isHome)}</body></html>`;
}

function breadcrumb(items, { syncSection = false, detail = false } = {}) {
  const normalizedItems = items.flatMap((item) => String(item).split('›').map((part) => part.trim()).filter(Boolean));
  return `<nav class="breadcrumb" aria-label="面包屑"${syncSection ? ` data-sync-section="${syncSection === true ? 'path' : syncSection}"` : ''}${detail ? ' data-detail-breadcrumb' : ''}><ol><li><a href="../index.html">首页</a></li>${normalizedItems.map((item, index) => index === normalizedItems.length - 1 ? `<li><span aria-current="page">${item}</span></li>` : `<li><a href="#">${item}</a></li>`).join('')}</ol></nav>`;
}

const menus = {
  party: `<ul class="side-menu"><li class="is-active"><a href="#party-news" data-section="party-news">党建要闻</a></li><li><a href="#study" data-section="study">学习教育</a></li><li><a href="#branch-activity" data-section="branch-activity">支部活动</a></li></ul>`,
  ministry: `<ul class="side-menu"><li><a href="#work" data-section="work">工作动态</a></li><li class="is-active"><button type="button" aria-expanded="true" data-section="policy-files">政策文件 <span class="side-menu-toggle-icon" aria-hidden="true"></span></button><ul class="side-submenu"><li class="current"><a href="ministry.html">国家鼓励发展的重大环保技术装备目录</a></li><li><a href="ministry.html">环保装备制造业规范条件企业</a></li><li><a href="ministry.html">重大环保技术装备创新任务揭榜挂帅</a></li><li><a href="ministry.html">其他文件</a></li></ul></li><li><a href="ministry.html#interpretation" data-section="interpretation">政策解读</a></li></ul>`,
  about: `<ul class="side-menu"><li class="is-active"><a href="#introduction" data-section="introduction">协会简介</a></li><li><a href="#organization" data-section="organization">组织架构</a></li><li><a href="#charter" data-section="charter">协会章程</a></li><li><button type="button" aria-expanded="true" data-section="branches">分支机构 <span class="side-menu-toggle-icon" aria-hidden="true"></span></button><ul class="side-submenu"><li><a href="about.html">水分会</a></li><li><a href="about.html">大气分会</a></li><li><a href="about.html">固废分会</a></li><li><a href="about.html">环境监测分会</a></li><li><a href="about.html">噪声分会</a></li><li><a href="about.html">紫外线分会</a></li><li><a href="about.html">臭氧分会</a></li><li><a href="about.html">人工智能分会</a></li><li><a href="about.html">环境工程分会</a></li></ul></li><li><a href="experts.html#experts" data-section="experts">专家委员会</a></li><li><a href="#disclosure" data-section="disclosure">公开公示</a></li><li><a href="#contact" data-section="contact">联系我们</a></li></ul>`,
  member: `<ul class="side-menu"><li class="is-active"><a href="member.html">行业报告</a></li><li><a href="member.html">数据中心</a></li><li><a href="member.html">电子刊物</a></li><li><a href="member.html">我的空间</a></li></ul>`,
  training: `<ul class="side-menu"><li><a href="#meetings" data-section="meetings">会议活动</a></li><li class="is-active"><button type="button" aria-expanded="true" data-section="talent">人才培训 <span class="side-menu-toggle-icon" aria-hidden="true"></span></button><ul class="side-submenu"><li class="current"><a href="training.html">通知及动态</a></li><li><a href="training.html">人才交流平台</a></li></ul></li><li><button type="button" aria-expanded="true" data-section="international">国际交流与合作 <span class="side-menu-toggle-icon" aria-hidden="true"></span></button><ul class="side-submenu"><li><a href="training.html">出海培训</a></li><li><a href="training.html">海外考察</a></li><li><a href="training.html">交流对接</a></li></ul></li></ul>`,
  standards: `<ul class="side-menu"><li class="is-active"><button type="button" aria-expanded="true" data-section="innovation">科技创新及成果转化 <span class="side-menu-toggle-icon" aria-hidden="true"></span></button><ul class="side-submenu"><li><a href="standards.html">科技成果评价</a></li><li><a href="standards.html">科技成果转化平台</a></li></ul></li><li><button type="button" aria-expanded="true" data-section="standard-work">标准工作 <span class="side-menu-toggle-icon" aria-hidden="true"></span></button><ul class="side-submenu"><li><a href="standards.html">通知及动态</a></li><li><a href="standards.html">协会标准发布</a></li></ul></li><li><a href="#green-promotion" data-section="green-promotion">绿色技术推广</a></li></ul>`,
  policy: `<ul class="side-menu"><li class="is-active"><a href="#reports" data-section="reports">行业报告</a></li><li><a href="#data" data-section="data">数据中心</a></li><li><a href="#catalogue" data-section="catalogue">国家鼓励发展的重大环保技术装备目录</a></li><li><a href="#qualified" data-section="qualified">环保装备制造业规范条件企业</a></li><li><a href="#innovation-tasks" data-section="innovation-tasks">重大环保技术装备创新任务揭榜挂帅</a></li></ul>`,
  experts: `<ul class="side-menu"><li class="is-active"><a href="#experts" data-section="experts">专家视野</a></li></ul>`,
  news: `<ul class="side-menu"><li class="is-active"><a href="#hot" data-section="hot">热点关注</a></li><li><a href="#notice" data-section="notice">通知公告</a></li><li><a href="#association" data-section="association">协会动态</a></li><li><a href="#member-news" data-section="member-news">会员动态</a></li></ul>`,
  videos: `<ul class="side-menu"><li class="is-active"><a href="videos.html">视频资讯</a></li><li><a href="videos.html">会员专享</a></li></ul>`,
};

const defaultTitles = [
  '《环保装备制造业高质量发展行动计划》专题研讨会在京召开',
  '关于征集环保科技创新成果和示范案例的通知',
  '中国环保机械行业协会会员单位绿色制造实践交流',
  '重点环保技术装备推广目录申报工作正式启动',
  '生态环境装备行业标准化工作座谈会成功举办',
  '推进减污降碳协同创新，培育行业发展新动能',
  '环保装备数字化转型与智能制造专题培训通知',
  '协会赴会员企业开展高质量发展专题调研',
  '工业固废资源化利用技术交流会报名开启',
  '环保装备企业国际合作服务对接活动举行',
];

const searchResults = [
  { title: '关于征集《晶体硅光伏组件回收再利用设备》等三项行业标准项目', summary: '面向全行业公开征集三项行业标准项目参编单位及相关技术资料。', category: '科技标准', date: '2026-07-28', href: 'standards-detail.html?from=standards-work' },
  { title: '加快构建环保装备制造业发展新格局 助力生态文明建设', summary: '协会将围绕产业基础提升、关键装备攻关和绿色制造汇聚行业力量。', category: '新闻中心', date: '2026-04-03', href: 'focus-detail-1.html?from=news-hot' },
  { title: '环保装备制造业高质量发展政策研究报告发布', summary: '围绕行业发展现状、重点任务和政策建议形成专题研究成果。', category: '政策研究', date: '2026-05-19', href: 'policy-detail.html?from=policy-reports' },
  { title: '国家鼓励发展的重大环保技术装备目录申报工作启动', summary: '有关部门启动年度技术装备目录申报及材料审核工作。', category: '部委动态', date: '2026-05-18', href: 'ministry-detail.html?from=ministry-major-equipment-catalogue' },
  { title: '三项环保装备团体标准正式发布实施', summary: '进一步完善环保装备标准体系，促进行业规范化发展。', category: '科技标准', date: '2026-05-19', href: 'standards-detail.html?from=standards-publications' },
  { title: '环保装备数字化转型与智能制造专题培训通知', summary: '专题培训面向会员企业技术人员和管理人员开放报名。', category: '交流培训', date: '2026-05-20', href: 'training-detail.html?from=training-talent-notices' },
  { title: '中国环保机械行业协会会员单位绿色制造实践交流', summary: '会员企业分享绿色工厂建设、节能改造和智能运维实践。', category: '新闻中心', date: '2026-05-16', href: 'news-detail.html?from=news-member' },
  { title: '绿色低碳产业政策专题解读会召开', summary: '专家围绕绿色低碳产业政策、申报要求和实施路径进行解读。', category: '部委动态', date: '2026-05-15', href: 'ministry-detail.html?from=ministry-policy-interpretation' },
  { title: '水污染治理装备技术应用研讨会会议通知', summary: '研讨会聚焦水污染治理装备创新、典型案例和供需对接。', category: '交流培训', date: '2026-05-16', href: 'training-detail.html?from=training-meetings' },
  { title: '专家视野：环保装备产业如何补短板、锻长板、聚优势', summary: '行业专家就科技创新和环保产业提质增效提出建议。', category: '专家委员会', date: '2026-04-18', href: 'experts-detail.html?from=about-expert-insights' },
  { title: '协会组织会员单位开展绿色工厂现场交流', summary: '活动组织会员单位考察绿色制造现场并开展经验交流。', category: '协会动态', date: '2026-05-12', href: 'news-detail.html?from=news-association' },
  { title: '环保装备行业标准建设与绿色低碳技术成果转化专题视频', summary: '视频介绍行业标准项目和绿色低碳技术成果转化进展。', category: '视频专区', date: '2026-07-28', href: 'video-detail.html?from=videos-news' },
];

const homeNoticeTitles = [
  '看见中国汽车的“同济”力量！2026中国汽车论坛同济大学专场主题论坛在上海嘉定召开',
  '中国汽车工业协会外资汽车企业委员会第八次主任委员会会议在北京召开',
  '凝聚行业共识 共绘发展蓝图—汽车行业“十五五”发展趋势研讨会在京成功召开',
  '构建智驾事故分析新范式—基于汽车行业可信数据空间的智驾保险服务落地实践',
  '构建智驾事故分析新范式—基于汽车行业可信数据空间的智驾保险服务落地实践',
  '构建智驾事故分析新范式—基于汽车行业可信数据空间的智驾保险服务落地实践',
];

const homeHeroArticles = [
  {
    id: 'demo-focus-01',
    title: '加快构建环保装备制造业发展新格局 助力生态文明建设',
    date: '2026-04-03',
    source: '众链科技',
    image: 'hero-building.png',
    alt: '会员企业厂区',
    href: 'pages/focus-detail-1.html',
    paragraphs: [
      '环保装备制造业是绿色低碳产业的重要组成部分。协会将围绕产业基础提升、关键装备攻关、绿色制造与服务能力建设，进一步汇聚企业、科研机构和应用单位力量，推动产业链协同创新。',
      '下一步将持续开展供需对接、技术交流和示范案例推广，促进先进环保装备在重点行业和重点场景落地应用，为生态文明建设提供更坚实的装备支撑。',
    ],
  },
  {
    id: 'demo-focus-02',
    title: '以科技创新推动环保装备产业绿色低碳转型',
    date: '2026-04-02',
    source: '中国环保机械行业协会',
    image: 'hero-water-treatment.jpg',
    alt: '现代化水处理设施',
    href: 'pages/focus-detail-2.html',
    paragraphs: [
      '科技创新正加速推动环保装备向高效、低碳、智能方向升级。行业企业通过数字化设计、智能监测和全过程能效管理，不断提升装备运行质量与资源利用效率。',
      '协会将加强创新成果评价、转化与应用服务，搭建产学研用协同平台，支持更多成熟技术形成可复制、可推广的绿色解决方案。',
    ],
  },
  {
    id: 'demo-focus-03',
    title: '凝聚行业共识 共绘环保装备高质量发展新蓝图',
    date: '2026-04-01',
    source: '中国环保机械行业协会',
    image: 'hero-conference.jpg',
    alt: '环保装备行业交流会议',
    href: 'pages/focus-detail-3.html',
    paragraphs: [
      '环保装备行业交流会议围绕高质量发展、标准体系建设、科技成果转化和国际合作等重点议题展开深入研讨，与会代表分享实践经验并提出协同发展的建议。',
      '协会将充分吸收会议成果，持续完善行业服务体系，凝聚会员单位共识，携手构建开放合作、创新驱动、规范有序的产业发展新格局。',
    ],
  },
];

const homeNoticeColumns = [
  { panel: 'notice-1', code: 'news-notice', href: 'pages/news.html#notice', label: '通知公告' },
  { panel: 'notice-2', code: 'news-association', href: 'pages/news.html#association', label: '协会动态' },
  { panel: 'notice-3', code: 'news-member', href: 'pages/news.html#member-news', label: '会员动态' },
];

const homeUpdates = [
  {
    code: 'training-meetings',
    detail: 'pages/training-detail.html',
    image: 'meeting-photo.png',
    alt: '环保装备行业会议现场',
    items: [
      ['第十一届环保装备高质量发展大会报名工作启动', '2026年5月24日'],
      ['环保装备产业链协同发展交流会在京召开', '2026年5月20日'],
      ['水污染治理装备技术应用研讨会会议通知', '2026年5月16日'],
      ['协会组织会员单位开展绿色工厂现场交流', '2026年5月12日'],
      ['固废资源化利用专题会议完成成果发布', '2026年5月8日'],
      ['环保装备供需对接活动征集参会单位', '2026年5月5日'],
    ],
  },
  {
    code: 'standards-work',
    detail: 'pages/standards-detail.html',
    image: 'hero-building.png',
    alt: '环保装备制造企业厂区',
    items: [
      ['《工业锅炉烟气多污染物协同治理技术规范》征求意见', '2026年5月23日'],
      ['三项环保装备团体标准正式发布实施', '2026年5月19日'],
      ['关于征集年度标准制修订项目建议的通知', '2026年5月15日'],
      ['环保装备标准化技术委员会年度会议召开', '2026年5月11日'],
      ['水处理装备能效评价标准研讨工作启动', '2026年5月7日'],
      ['协会标准参编单位及技术资料公开征集', '2026年5月3日'],
    ],
  },
  {
    code: 'standards-innovation',
    detail: 'pages/standards-detail.html',
    image: 'hero-water-treatment.jpg',
    alt: '现代化水处理技术设施',
    items: [
      ['关于征集环保科技创新成果和示范案例的通知', '2026年5月22日'],
      ['高效节能水处理装备科技成果完成评价', '2026年5月18日'],
      ['智能监测技术赋能环保装备数字化升级', '2026年5月14日'],
      ['绿色低碳技术成果转化对接平台上线', '2026年5月10日'],
      ['环保装备首台套产品应用案例开始征集', '2026年5月6日'],
      ['产学研协同创新项目取得阶段性成果', '2026年5月2日'],
    ],
  },
  {
    code: 'training-international',
    detail: 'pages/training-detail.html',
    image: 'hero-conference.jpg',
    alt: '环保装备行业国际交流会议',
    items: [
      ['关于赴丹麦开展水与环境技术考察交流活动的通知', '2026年5月21日'],
      ['中欧环保装备产业合作交流会成功举办', '2026年5月17日'],
      ['协会参加国际环境技术与设备展览会', '2026年5月13日'],
      ['海外环保项目合作需求信息集中发布', '2026年5月9日'],
      ['环保装备企业国际化经营专题座谈会召开', '2026年5月5日'],
      ['国际绿色技术合作伙伴招募工作启动', '2026年5月1日'],
    ],
  },
  {
    code: 'training-talent',
    detail: 'pages/training-detail.html',
    image: 'expert-photo.png',
    alt: '环保装备行业专家授课现场',
    items: [
      ['环保装备数字化转型与智能制造专题培训通知', '2026年5月20日'],
      ['环境工程项目管理高级研修班开始报名', '2026年5月16日'],
      ['协会举办标准编写与质量管理实务培训', '2026年5月12日'],
      ['环保装备专业技术人才评价工作启动', '2026年5月8日'],
      ['青年工程师创新能力提升计划正式发布', '2026年5月4日'],
      ['会员企业职业技能培训需求开始征集', '2026年4月30日'],
    ],
  },
  {
    code: 'policy-reports',
    detail: 'pages/policy-detail.html',
    image: 'conference-banner.png',
    alt: '环保装备行业政策研究交流活动',
    items: [
      ['环保装备制造业高质量发展政策研究报告发布', '2026年5月19日'],
      ['绿色低碳产业政策专题解读会召开', '2026年5月15日'],
      ['环保装备企业经营情况问卷调研启动', '2026年5月11日'],
      ['行业“十五五”发展规划前期研究取得进展', '2026年5月7日'],
      ['重点环保技术装备推广政策研究座谈会举行', '2026年5月3日'],
      ['环保产业投融资政策信息汇编正式发布', '2026年4月29日'],
    ],
  },
];

const partnerRows = [
  [
    { name: '南京大学环境学院', image: 'nanjing-university.png', url: 'https://www.nju.edu.cn/', size: '600 535', viewBox: '40 14 526 521' },
    { name: '紫金龙净环保新能源股份有限公司', image: 'longking.png', url: 'https://www.longking.com.cn/', size: '356 349', viewBox: '35 35 280 282' },
    { name: '北京城市排水集团有限责任公司', image: 'beijing-drainage.png', url: 'https://www.bdc.cn/', size: '438 258', viewBox: '0 78 438 125' },
    { name: '苏州帝瀚环保科技股份有限公司', image: 'dihill.png', url: 'https://www.dihillgreen.com/', size: '945 944', viewBox: '85 385 809 162' },
    { name: '江苏一环集团有限公司', image: 'jiangsu-yihuan.png', url: 'http://www.yihuan.com/', size: '1178 784', viewBox: '0 40 1178 716' },
    { name: '河南康宁特环保科技股份有限公司', image: 'kangningte.png', url: 'https://www.knthb.com/', size: '167 63', viewBox: '0 0 167 63' },
  ],
  [
    { name: '中国天楹股份有限公司', image: 'cnty.png', url: 'https://www.cnty.cn/', size: '192 192', viewBox: '0 8 192 176' },
    { name: '杰瑞新能源再生循环科技有限公司', image: 'jereh-recycling.png', url: 'https://www.jereh.com/cn/', size: '536 536', viewBox: '0 175 536 190' },
    { name: '合肥通用机械研究院有限公司', image: 'hefei-general-machinery.png', url: 'http://www.hgmri.com/', size: '1068 446', viewBox: '10 45 1045 376' },
    { name: '长江生态环保集团有限公司', image: 'yangtze-ecology.png', url: 'https://www.yeec.com.cn/', size: '418 55', viewBox: '0 0 418 55' },
    { name: '科林环保技术有限责任公司', image: 'kelin.png', url: 'https://www.kelin-china.com/', size: '500 500', viewBox: '25 105 441 303' },
    { name: '中车产业投资有限公司', image: 'crrc-investment.png', url: 'https://www.crrcgc.cc/cytz/277_19585/index.html', size: '600 434', viewBox: '20 112 560 208' },
  ],
];

const partnerMarquee = partnerRows.map((row) => `<div class="partner-row"><div class="partner-track">${[false, true].map((duplicate) => `<div class="partner-set" ${duplicate ? 'aria-hidden="true"' : ''}>${row.map((partner) => {
  const [width, height] = partner.size.split(' ');
  return `<a class="partner-logo" href="${partner.url}" target="_blank" rel="noopener noreferrer" aria-label="访问${partner.name}官网" ${duplicate ? 'tabindex="-1"' : ''}><svg class="partner-logo-art" viewBox="${partner.viewBox}" aria-hidden="true" focusable="false"><image href="assets/images/partner-logos/${partner.image}" width="${width}" height="${height}"></image></svg></a>`;
}).join('')}</div>`).join('')}</div></div>`).join('');

function detailHref(path, columnCode) {
  return `${path}?from=${encodeURIComponent(columnCode)}`;
}

function listRows(titles = defaultTitles, detail = 'ministry-detail.html', detailColumnCode = 'ministry-policy-files') {
  return `<ul class="news-list">${titles.map((title, index) => `<li class="news-row searchable"><a href="${detailHref(detail, detailColumnCode)}" data-detail-link data-detail-base="${detail}" data-default-column-code="${detailColumnCode}" data-base-title="${title}">${title}</a><time datetime="2026-${index < 4 ? '07' : '06'}-${String(28 - index).padStart(2, '0')}">${index < 4 ? '2026-07' : '2026-06'}-${String(28 - index).padStart(2, '0')}</time></li>`).join('')}</ul>`;
}

function pagination() {
  return `<div class="pagination-wrap"><span>共 101 项数据</span><div class="pagination" aria-label="分页"><button class="plain" type="button" aria-label="上一页">‹</button>${[1,2,3,4,5].map((n) => `<button class="${n === 1 ? 'active' : ''}" type="button" data-page="${n}">${n}</button>`).join('')}<span>…</span><button type="button" data-page="11">11</button><button class="plain" type="button" aria-label="下一页">›</button></div><label class="page-jump">跳至 <input type="number" min="1" max="20" value="11"> /20 页</label></div>`;
}

function listPage({ title, active, crumb, menu, titles, detail, detailColumnCode, theme = '', syncSectionBreadcrumb = false }) {
  return documentPage({ title, active, content: `<main class="page-shell ${theme}"><div class="page-inner">${breadcrumb([crumb], { syncSection: syncSectionBreadcrumb })}<div class="content-grid"><aside>${menu}</aside><section><div class="list-card">${listRows(titles, detail, detailColumnCode)}</div>${pagination()}</section></div></div></main>` });
}

function searchPage() {
  const rows = searchResults.map((item) => `<li class="search-result-item" data-search-result><a href="${item.href}"><h2>${item.title}</h2><p>${item.summary}</p><div class="search-result-meta"><span>${item.category}</span><time datetime="${item.date}">${item.date}</time></div></a></li>`).join('');
  return documentPage({ title: '站内搜索', active: 'search', content: `<main class="page-shell search-page" data-search-page><div class="page-inner">${breadcrumb(['站内搜索'])}<section class="search-card"><h1>站内搜索</h1><form class="search-page-form" role="search" action="search.html" method="get"><label class="sr-only" for="search-page-input">搜索关键词</label><input id="search-page-input" name="q" type="search" placeholder="请输入关键词" autocomplete="off"><button type="submit">搜索 <span class="search-icon" aria-hidden="true"></span></button></form><p class="search-summary" data-search-summary aria-live="polite">请输入关键词进行搜索</p><ul class="search-result-list">${rows}</ul><div class="search-empty" data-search-empty hidden><strong>未找到相关内容</strong><p>请尝试缩短关键词，或更换其他关键词重新搜索。</p></div><nav class="search-pagination pagination" data-search-pagination aria-label="搜索结果分页" hidden><button class="plain" type="button" data-search-prev aria-label="上一页">‹</button><span data-search-pages></span><button class="plain" type="button" data-search-next aria-label="下一页">›</button></nav></section></div></main>` });
}

  const articleParagraphs = [
  '为响应国家新能源绿色低碳发展战略，健全退役光伏设备循环利用产业标准体系，规范光伏回收设备研发、生产与应用流程，协会面向全行业公开征集《晶体硅光伏组件回收再利用设备》等三项行业标准项目参编单位及相关技术资料，助力新能源固废资源化利用产业规范化、高质量发展。',
  '本次征集的三项行业标准项目均纳入2026年度机械行业标准制定计划，项目编号分别为20-26-0656T-JB、2026-0657T-JB、2026-0658T-JB，聚焦新能源固废、生活垃圾、冶金污泥三大固废资源化核心领域。标准将重点规范晶体硅光伏组件拆解、分层、材料提取、提纯再生等全流程回收设备的术语定义、技术参数、结构设计、性能指标、安全规范、检测方法及使用运维要求，适配主流单晶、多晶、双玻等各类光伏组件的回收场景。另外两项配套标准同步完善固废资源化设备标准体系，提升固废精细化分选利用率，规范冶金污泥贵金属、有色金属回收设备的工艺指标与环保要求。',
  '据标准化工作负责人介绍，本次标准编制将立足国内产业实际，结合行业先进技术、成熟应用案例及国家绿色低碳、固废处置相关政策要求，兼顾设备安全性、环保性、经济性与通用性。标准落地后，将有效统一行业设备技术门槛，淘汰落后低效回收装备，引导光伏回收设备企业规范化研发生产，提升退役光伏组件硅片、银浆、玻璃、背板等核心材料的回收纯度与利用率，降低回收过程能耗及二次污染风险。',
];

function articlePage({ title, active, crumbItems, menu, detailColumnCode, party = false, articleTitle = '关于征集《晶体硅光伏组件回收再利用设备》等三项行业标准项目', articleDate = '2026-07-28', articleSource = '科技标准', paragraphs = articleParagraphs }) {
  return documentPage({ title, active, content: `<main class="page-shell ${party ? 'party-theme' : ''}" data-detail-page data-default-column-code="${detailColumnCode}"><div class="page-inner">${breadcrumb([...crumbItems, articleTitle], { detail: true })}<div class="content-grid"><aside>${menu}</aside><article class="article-card"><header class="article-header"><h1>${articleTitle}</h1><div class="article-meta"><span>${articleDate}</span><span>${articleSource}</span></div></header><div class="article-body">${paragraphs.map((p) => `<p>${p}</p>`).join('')}</div><nav class="article-pager"><a href="#"><strong>上一篇：</strong>凝聚产教校友合力，为汽车强国建设贡献同济力量</a><a href="#"><strong>下一篇：</strong>环保装备行业绿色低碳发展专题交流会召开</a></nav></article></div></div></main>` });
}

const homeContent = `
<main class="page-shell home-page"><div class="page-inner home-main">
  <section class="home-lead" aria-label="重点信息">
    <article class="hero-slider" data-column-code="home-focus" data-source-column-code="news-hot" data-display-type="carousel" aria-roledescription="轮播图" aria-label="重点新闻"><div class="hero-slides">${homeHeroArticles.map((article, index) => `<a class="hero-slide${index === 0 ? ' active' : ''}" href="${detailHref(article.href, 'news-hot')}" data-slide="${index}" data-article-id="${article.id}"${index ? ' aria-hidden="true" tabindex="-1"' : ''}><div class="hero-media"><img src="assets/images/${article.image}" alt="${article.alt}"></div><div class="hero-caption"><h1>${article.title}</h1><p>发布时间：${article.date}　　来源：${article.source}</p></div></a>`).join('')}</div><div class="hero-dots" aria-label="选择轮播内容">${homeHeroArticles.map((_, index) => `<button class="${index === 0 ? 'active' : ''}" type="button" data-slide-to="${index}" aria-label="显示第${index + 1}张"${index === 0 ? ' aria-current="true"' : ''}></button>`).join('')}</div></article>
    <section class="notice-panel" data-column-group="home-news-tabs"><div class="home-tabs" data-group="notice">${homeNoticeColumns.map((column, index) => `<button class="${index === 0 ? 'active' : ''}" data-tab="${column.panel}" data-column-code="${column.code}" data-href="${column.href}" type="button">${column.label}</button>`).join('')}</div>${homeNoticeColumns.map((column, tab) => `<ul id="${column.panel}" class="notice-list" data-tab-group="notice" data-column-code="${column.code}" ${tab ? 'hidden' : ''}>${(tab ? defaultTitles.slice(tab, tab + 6) : homeNoticeTitles).map((item) => `<li class="searchable"><a href="${detailHref('pages/news-detail.html', column.code)}">${item}</a></li>`).join('')}</ul>`).join('')}</section>
    <aside class="quick-rail" data-column-code="home-service-links" data-display-type="link"><a href="pages/member.html"><img class="quick-icon" src="assets/images/icon-join.png" alt="">申请入会</a><a href="pages/member.html"><img class="quick-icon" src="assets/images/icon-member.png" alt="">会员中心</a><div class="quick-mobile"><button class="quick-mobile-trigger" type="button" aria-expanded="false" aria-controls="mobile-qr-card" aria-haspopup="dialog"><img class="quick-icon" src="assets/images/icon-mobile.png" alt="">移动媒体</button><div class="mobile-qr-card" id="mobile-qr-card" role="dialog" aria-label="协会移动媒体二维码" hidden><div class="mobile-qr-item"><img src="assets/images/qr-wechat-service.jpg" alt="CAMIE 微信服务号二维码"><span>CAMIE</span><strong>微信服务号</strong></div><div class="mobile-qr-item"><img src="assets/images/qr-wechat-video.png" alt="CAMIE 微信视频号二维码"><span>CAMIE</span><strong>微信视频号</strong></div><div class="mobile-qr-item"><img src="assets/images/qr-wechat-subscription.jpg" alt="CAMIE 微信订阅号二维码"><span>CAMIE</span><strong>微信订阅号</strong></div></div></div><a href="pages/videos.html"><img class="quick-icon" src="assets/images/icon-video.png" alt="">视频专区</a></aside>
  </section>
  <section class="updates" data-column-group="home-update-tabs"><div class="updates-tabs" data-group="updates"><button class="active" data-tab="updates-1" data-column-code="training-meetings" data-href="pages/training.html#meetings" type="button">会议活动</button><button data-tab="updates-2" data-column-code="standards-work" data-href="pages/standards.html#standard-work" type="button">标准工作</button><button data-tab="updates-3" data-column-code="standards-innovation" data-href="pages/standards.html#innovation" type="button">科技创新及成果转化</button><button data-tab="updates-4" data-column-code="training-international" data-href="pages/training.html#international" type="button">国际合作</button><button data-tab="updates-5" data-column-code="training-talent" data-href="pages/training.html#talent" type="button">人才培训</button><button data-tab="updates-6" data-column-code="policy-reports" data-href="pages/policy.html#reports" type="button">政策研究</button></div>${homeUpdates.map((update, index) => `<div id="updates-${index + 1}" class="updates-panel" data-tab-group="updates" data-column-code="${update.code}" ${index ? 'hidden' : ''}><ul class="updates-list">${update.items.map(([title, date]) => `<li class="searchable"><a href="${detailHref(update.detail, update.code)}">${title}</a><time>${date}</time></li>`).join('')}</ul><a class="updates-image" href="${detailHref(update.detail, update.code)}"><img src="assets/images/${update.image}" alt="${update.alt}"></a></div>`).join('')}</section>
  <a class="conference-banner" data-column-code="home-conference-banner" data-display-type="ad" href="pages/training.html"><img src="assets/images/conference-banner.png" alt="第十一届环保装备高质量发展大会"></a>
  <nav class="feature-links" data-column-code="home-special-topics" data-display-type="link"><a href="pages/standards.html">绿色技术推广</a><a href="pages/ministry.html">国家鼓励发展的重大环保技术装备目录</a><a href="pages/policy.html">环保装备制造业规范条件企业</a><a href="pages/ministry.html">重大环保技术装备创新任务揭榜挂帅</a></nav>
  <div class="home-lower"><div class="columns-two"><section data-column-code="about-branches"><h2 class="section-title"><a href="pages/about.html#branches"><img src="assets/images/leaf-icon.png" alt="">分支机构</a></h2><div class="branch-grid">${[['水分会','water'],['大气分会','atmosphere'],['固废分会','solid-waste'],['环境监测分会','monitoring'],['噪声分会','noise'],['紫外线分会','uv'],['臭氧分会','odor'],['人工智能分会','ai'],['环境工程分会','engineering']].map(([name,img]) => `<a class="branch-card" href="pages/about.html"><img src="assets/images/branch-${img}.png" alt="">${name}</a>`).join('')}</div></section><section data-column-code="about-experts"><h2 class="section-title"><a href="pages/experts.html"><img src="assets/images/leaf-icon.png" alt="">专家委员会</a></h2><a class="expert-feature" href="${detailHref('pages/experts-detail.html', 'about-expert-insights')}" aria-label="查看王亦宁委员专家视野详情"><img src="assets/images/expert-photo.png" alt="王亦宁委员发言"><p>王亦宁<br><small>中国环保机械行业协会<br>专家委员会主任委员</small></p></a><ul class="expert-links"><li><a href="pages/experts.html">如何补短板锻长板聚优势蓄后势？</a></li><li><a href="pages/experts.html">科技创新引领环保产业提质增效</a></li></ul></section></div>
  <section class="partners" data-column-code="home-vice-president-members" data-display-type="link"><h2 class="section-title"><img src="assets/images/leaf-icon.png" alt="">副会长单位 <small>（排名不分前后）</small></h2><div class="partner-grid">${partnerMarquee}</div></section></div>
</div></main>`;

writeFileSync(join(root, 'index.html'), documentPage({ title: '首页', active: 'home', isHome: true, bodyClass: 'home-body', content: homeContent }));

const pageSpecs = [
  ['party.html', listPage({ title: '党建专栏', active: 'party', crumb: '党建专栏 › 党建要闻', menu: menus.party, titles: defaultTitles.map((_, i) => `党建引领环保装备行业高质量发展工作动态 ${i + 1}`), detail: 'party-detail.html', detailColumnCode: 'party-news', theme: 'party-theme', syncSectionBreadcrumb: true })],
  ['ministry.html', listPage({ title: '部委动态', active: 'ministry', crumb: '部委动态 › 政策文件 › 国家鼓励发展的重大环保技术装备目录', menu: menus.ministry, detail: 'ministry-detail.html', detailColumnCode: 'ministry-major-equipment-catalogue', syncSectionBreadcrumb: true })],
  ['about.html', listPage({ title: '关于协会', active: 'about', crumb: '关于协会', menu: menus.about, detail: 'about-detail.html', detailColumnCode: 'about-introduction', syncSectionBreadcrumb: true })],
  ['member.html', listPage({ title: '会员中心', active: 'home', crumb: '会员中心', menu: menus.member, detail: 'member-detail.html', detailColumnCode: 'member-reports' })],
  ['training.html', listPage({ title: '交流培训', active: 'training', crumb: '交流培训', menu: menus.training, titles: defaultTitles.map((_, i) => `关于环保装备专业人才培训第 ${i + 1} 期的通知`), detail: 'training-detail.html', detailColumnCode: 'training-talent', syncSectionBreadcrumb: true })],
  ['standards.html', listPage({ title: '科技标准', active: 'standards', crumb: '科技标准', menu: menus.standards, detail: 'standards-detail.html', detailColumnCode: 'standards-innovation', syncSectionBreadcrumb: true })],
  ['policy.html', listPage({ title: '政策研究', active: 'ministry', crumb: '政策研究', menu: menus.policy, detail: 'policy-detail.html', detailColumnCode: 'policy-reports', syncSectionBreadcrumb: true })],
  ['experts.html', listPage({ title: '专家委员会', active: 'about', crumb: '专家委员会', menu: menus.experts, titles: defaultTitles.map((_, i) => `专家视野：环保装备产业如何补短板、锻长板、聚优势（${i + 1}）`), detail: 'experts-detail.html', detailColumnCode: 'about-expert-insights' })],
  ['news.html', listPage({ title: '新闻中心', active: 'news', crumb: '新闻中心', menu: menus.news, detail: 'news-detail.html', detailColumnCode: 'news-hot', syncSectionBreadcrumb: true })],
  ['search.html', searchPage()],
  ['party-detail.html', articlePage({ title: '党建内容', active: 'party', crumbItems: ['党建专栏', '党建要闻'], menu: menus.party, detailColumnCode: 'party-news', party: true })],
  ['ministry-detail.html', articlePage({ title: '部委动态详情', active: 'ministry', crumbItems: ['部委动态', '政策文件', '国家鼓励发展的重大环保技术装备目录'], menu: menus.ministry, detailColumnCode: 'ministry-major-equipment-catalogue' })],
  ['about-detail.html', articlePage({ title: '关于协会详情', active: 'about', crumbItems: ['关于协会', '协会简介'], menu: menus.about, detailColumnCode: 'about-introduction' })],
  ['member-detail.html', articlePage({ title: '会员中心详情', active: 'home', crumbItems: ['会员中心', '行业报告'], menu: menus.member, detailColumnCode: 'member-reports' })],
  ['training-detail.html', articlePage({ title: '交流培训详情', active: 'training', crumbItems: ['交流培训', '人才培训'], menu: menus.training, detailColumnCode: 'training-talent' })],
  ['standards-detail.html', articlePage({ title: '科技标准详情', active: 'standards', crumbItems: ['科技标准', '科技创新及成果转化'], menu: menus.standards, detailColumnCode: 'standards-innovation' })],
  ['policy-detail.html', articlePage({ title: '政策研究详情', active: 'ministry', crumbItems: ['政策研究', '行业报告'], menu: menus.policy, detailColumnCode: 'policy-reports' })],
  ['experts-detail.html', articlePage({ title: '专家视野详情', active: 'about', crumbItems: ['关于协会', '专家委员会', '专家视野'], menu: menus.experts, detailColumnCode: 'about-expert-insights' })],
  ['news-detail.html', articlePage({ title: '新闻详情', active: 'news', crumbItems: ['新闻中心', '热点关注'], menu: menus.news, detailColumnCode: 'news-hot' })],
  ...homeHeroArticles.map((article, index) => [`focus-detail-${index + 1}.html`, articlePage({ title: article.title, active: 'news', crumbItems: ['新闻中心', '热点关注'], menu: menus.news, detailColumnCode: 'news-hot', articleTitle: article.title, articleDate: article.date, articleSource: article.source, paragraphs: article.paragraphs })]),
];

const videoDetailHref = detailHref('video-detail.html', 'videos-news');
const videoCards = Array.from({ length: 6 }, (_, i) => `<article class="video-card searchable"><div class="video-thumb"><img src="../assets/images/video-thumb.png" alt="会议视频缩略图"><a class="play-button" href="${videoDetailHref}" aria-label="进入视频详情页播放"></a></div><h2><a href="${videoDetailHref}">关于征集《晶体硅光伏组件回收再利用设备》等三项行业标准项目</a></h2><p>聚焦环保装备行业标准建设与绿色低碳技术成果转化的专题视频内容。</p><footer><time>2026-07-${28 - i}</time><a href="${videoDetailHref}">查看详情 →</a></footer></article>`).join('');
pageSpecs.push(['videos.html', documentPage({ title: '视频专区', active: 'news', content: `<main class="page-shell"><div class="page-inner">${breadcrumb(['视频专区'])}<div class="content-grid"><aside>${menus.videos}</aside><section><div class="video-grid-card"><div class="video-grid">${videoCards}</div></div>${pagination()}</section></div></div></main>` })]);
pageSpecs.push(['video-detail.html', documentPage({ title: '视频详情', active: 'news', content: `<main class="page-shell" data-detail-page data-default-column-code="videos-news"><div class="page-inner">${breadcrumb(['视频专区', '视频资讯', '关于征集《晶体硅光伏组件回收再利用设备》等三项行业标准项目'], { detail: true })}<div class="content-grid"><aside>${menus.videos}</aside><article class="article-card video-detail-card"><header class="article-header"><h1>关于征集《晶体硅光伏组件回收再利用设备》等三项行业标准项目</h1><div class="article-meta"><span>2026-07-28</span><span>环保装备行业标准专题视频</span></div></header><div class="video-stage"><video controls preload="metadata" poster="../assets/images/video-poster.png"><source src="../assets/videos/big-buck-bunny.mp4" type="video/mp4">您的浏览器不支持 HTML5 视频播放。</video><button class="play-button" type="button" data-inline-video-play aria-label="播放视频"></button></div></article></div></div></main>` })]);

for (const [filename, html] of pageSpecs) writeFileSync(join(pagesDir, filename), html);

console.log(`Generated ${pageSpecs.length + 1} static preview pages.`);
