(() => {
  'use strict';
  const content = window.CAMIE_STATIC_CONTENT || {articles: [], articlePaths: {}, columns: {}};
  const prefix = document.body.dataset.rootPrefix || './';
  const root = value => /^(?:[a-z]+:|\/\/|\/|#)/i.test(value) ? value : prefix + value.replace(/^\.\//, '');
  const params = new URLSearchParams(location.search);

  // Preserve previously shared SPA links and MIIC-style detail compatibility URLs.
  function legacyTarget() {
    if (location.hash.startsWith('#/')) {
      const [route, query = ''] = location.hash.slice(2).split('?');
      const legacy = new URLSearchParams(query);
      if (route.startsWith('news/')) return content.articlePaths[route.split('/')[1]] || 'news.html';
      if (route.startsWith('video/')) {
        const id = Number(route.split('/')[1]);
        return content.articles.find(a => a.id === id && a.video)?.href || content.defaultVideo;
      }
      if (legacy.has('q')) return 'search.html?q=' + encodeURIComponent(legacy.get('q'));
      if (legacy.has('category')) return content.columns[legacy.get('category')] || 'news.html';
      if (route === 'videos') return 'videos.html';
      if (route === 'news') return 'news.html';
      if (legacy.get('section') === 'services') return content.columns['协会服务'] || 'news.html';
      return 'index.html';
    }
    if (params.has('id') && /(?:^|\/)detail\.html$/.test(location.pathname)) {
      return content.articlePaths[params.get('id')] || 'news.html';
    }
    if (params.has('category')) return content.columns[params.get('category')];
    if (params.has('q') && document.body.dataset.page !== 'search') return 'search.html?q=' + encodeURIComponent(params.get('q'));
    return '';
  }
  const target = legacyTarget();
  if (target) {location.replace(root(target)); return;}
  window.addEventListener('hashchange', () => {const target = legacyTarget(); if (target) location.replace(root(target));});

  const query = params.get('q') || '';
  document.querySelectorAll('.search-form input').forEach(input => {input.value = query;});
  const escape = value => String(value).replace(/[&<>"']/g, char => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[char]));
  if (document.body.dataset.page === 'search') {
    const listing = document.querySelector('#listing');
    const term = query.trim().toLocaleLowerCase();
    const matches = term ? content.articles.filter(a => `${a.title} ${a.summary} ${a.category}`.toLocaleLowerCase().includes(term)) : [];
    let page = 1;
    const size = 14;
    const total = Math.max(1, Math.ceil(matches.length / size));
    const render = () => {
      listing.innerHTML = `<div class="list-panel"><h1 class="result-title">${term ? `“${escape(query)}” 的搜索结果（${matches.length}）` : '搜索资讯'}</h1>${matches.length ? matches.slice((page - 1) * size, page * size).map(a => `<a class="news-row" href="${escape(root(a.href))}"><span>${escape(a.title)}</span><time>${escape(a.date)}</time></a>`).join('') : '<p class="empty-state">未找到相关资讯，请尝试其他关键词。</p>'}</div><nav class="pagination" aria-label="搜索结果分页"><button data-search-page="${page - 1}" ${page === 1 ? 'disabled' : ''}>上一页</button><span>${page} / ${total}</span><button data-search-page="${page + 1}" ${page === total ? 'disabled' : ''}>下一页</button></nav>`;
    };
    listing.addEventListener('click', e => {const button = e.target.closest('[data-search-page]'); if (button) {page = Math.max(1, Math.min(total, Number(button.dataset.searchPage))); render();}});
    render();
  }

  const selectPanel = (group, value) => {
    document.querySelectorAll(`[data-${group}]`).forEach(tab => {
      const active = tab.dataset[group] === value;
      tab.classList.toggle('active', active);
      tab.setAttribute(group === 'slide' ? 'aria-pressed' : 'aria-selected', String(active));
    });
    document.querySelectorAll(`[data-${group}-panel]`).forEach(panel => {panel.hidden = panel.dataset[`${group}Panel`] !== value;});
  };
  for (const group of ['notice', 'tech']) {
    document.querySelectorAll(`[data-${group}]`).forEach(tab => {
      tab.addEventListener('pointerenter', () => selectPanel(group, tab.dataset[group]));
      tab.addEventListener('focus', () => selectPanel(group, tab.dataset[group]));
    });
  }
  document.addEventListener('click', e => {
    const button = e.target.closest('button');
    if (!button) return;
    for (const group of ['notice', 'tech', 'partner']) {
      if (button.hasAttribute(`data-${group}`)) selectPanel(group, button.dataset[group]);
    }
    if (button.hasAttribute('data-quick')) {
      const container = button.closest('.quick-links');
      const links = [...container.querySelectorAll('.quick-link')];
      if (button.dataset.quick === '1') container.insertBefore(links[0], container.querySelector('button'));
      else container.insertBefore(links.at(-1), links[0]);
    }
    const dialog = document.querySelector('.modal');
    if (button.matches('.modal-close,.modal-dismiss')) dialog?.close();
    if (button.hasAttribute('data-open')) {
      const name = button.dataset.open;
      if (name === '党建专栏') {location.href = root(content.columns[name] || 'news.html'); return;}
      if (name !== '视频播放') return;
      dialog.querySelector('h2').textContent = name;
      dialog.querySelector('.modal-body').innerHTML = '<p>本条资讯尚未配置视频文件。</p>';
      dialog.showModal();
    }
  });

  const heroCarousel = document.querySelector('[data-hero-carousel]');
  if (heroCarousel) {
    const slides = [...heroCarousel.querySelectorAll('[data-slide-panel]')];
    const dots = [...heroCarousel.querySelectorAll('[data-slide]')];
    const reducedMotion = matchMedia('(prefers-reduced-motion: reduce)').matches;
    let current = 0;
    let timer;
    const pauseReasons = new Set();
    const showHeroSlide = value => {
      if (!slides.length) return;
      current = (Number(value) + slides.length) % slides.length;
      slides.forEach((slide, index) => {slide.hidden = index !== current;});
      dots.forEach((dot, index) => {
        const active = index === current;
        dot.classList.toggle('active', active);
        dot.setAttribute('aria-pressed', String(active));
      });
    };
    const stopHeroCarousel = () => {clearInterval(timer); timer = undefined;};
    const startHeroCarousel = () => {
      stopHeroCarousel();
      if (!reducedMotion && slides.length > 1 && !document.hidden && pauseReasons.size === 0) {
        timer = setInterval(() => showHeroSlide(current + 1), 5000);
      }
    };
    dots.forEach(dot => dot.addEventListener('click', () => {
      showHeroSlide(dot.dataset.slide);
      startHeroCarousel();
    }));
    heroCarousel.querySelectorAll('[data-hero-step]').forEach(button => button.addEventListener('click', () => {
      showHeroSlide(current + Number(button.dataset.heroStep));
      startHeroCarousel();
    }));
    heroCarousel.addEventListener('focusin', () => {pauseReasons.add('focus'); stopHeroCarousel();});
    heroCarousel.addEventListener('focusout', e => {
      if (!heroCarousel.contains(e.relatedTarget)) {pauseReasons.delete('focus'); startHeroCarousel();}
    });
    document.addEventListener('visibilitychange', () => document.hidden ? stopHeroCarousel() : startHeroCarousel());
    showHeroSlide(0);
    startHeroCarousel();
  }

  const eventCarousel = document.querySelector('[data-event-carousel]');
  if (eventCarousel) {
    const slides = [...eventCarousel.querySelectorAll('[data-event-panel]')];
    const dots = [...eventCarousel.querySelectorAll('[data-event-slide]')];
    const reducedMotion = matchMedia('(prefers-reduced-motion: reduce)').matches;
    let current = 0;
    let timer;
    const pauseReasons = new Set();
    const showEventSlide = value => {
      current = (Number(value) + slides.length) % slides.length;
      eventCarousel.style.setProperty('--event-index', current);
      slides.forEach((slide, index) => {
        const active = index === current;
        slide.setAttribute('aria-hidden', String(!active));
        slide.tabIndex = active ? 0 : -1;
      });
      dots.forEach((dot, index) => {
        const active = index === current;
        dot.classList.toggle('active', active);
        dot.setAttribute('aria-pressed', String(active));
      });
    };
    const stopEventCarousel = () => {clearInterval(timer); timer = undefined;};
    const startEventCarousel = () => {
      stopEventCarousel();
      if (!reducedMotion && slides.length > 1 && !document.hidden && pauseReasons.size === 0) {
        timer = setInterval(() => showEventSlide(current + 1), 5000);
      }
    };
    dots.forEach(dot => dot.addEventListener('click', () => {
      showEventSlide(dot.dataset.eventSlide);
      startEventCarousel();
    }));
    eventCarousel.addEventListener('pointerenter', () => {pauseReasons.add('pointer'); stopEventCarousel();});
    eventCarousel.addEventListener('pointerleave', () => {pauseReasons.delete('pointer'); startEventCarousel();});
    eventCarousel.addEventListener('focusin', () => {pauseReasons.add('focus'); stopEventCarousel();});
    eventCarousel.addEventListener('focusout', e => {
      if (!eventCarousel.contains(e.relatedTarget)) {pauseReasons.delete('focus'); startEventCarousel();}
    });
    document.addEventListener('visibilitychange', () => document.hidden ? stopEventCarousel() : startEventCarousel());
    showEventSlide(0);
    startEventCarousel();
  }

  function jump(input) {
    const value = Math.min(Number(input.max), Math.max(1, Math.trunc(Number(input.value)) || 1));
    location.href = input.dataset.pagePrefix + value + '.html';
  }
  document.addEventListener('change', e => {if (e.target.matches('[data-page-prefix]')) jump(e.target);});
  document.addEventListener('keydown', e => {
    if (e.key === 'Enter' && e.target.matches('[data-page-prefix]')) {e.preventDefault(); jump(e.target);}
    if (e.target.matches('[role="tab"]') && ['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(e.key)) {
      e.preventDefault();
      const tabs = [...e.target.closest('[role="tablist"]').querySelectorAll('[role="tab"]')];
      const index = tabs.indexOf(e.target);
      const next = e.key === 'Home' ? 0 : e.key === 'End' ? tabs.length - 1 : (index + (e.key === 'ArrowRight' ? 1 : -1) + tabs.length) % tabs.length;
      const nextTab = tabs[next];
      nextTab.focus();
      for (const group of ['notice', 'tech', 'partner']) {
        if (nextTab.hasAttribute(`data-${group}`)) selectPanel(group, nextTab.dataset[group]);
      }
    }
  });
})();
