(() => {
  const body = document.body;
  const menuButton = document.querySelector('.mobile-menu-button');
  menuButton?.addEventListener('click', () => {
    const open = body.classList.toggle('menu-open');
    menuButton.setAttribute('aria-expanded', String(open));
  });

  document.querySelectorAll('.side-menu button[aria-expanded]').forEach((button) => {
    button.addEventListener('click', () => {
      button.setAttribute('aria-expanded', String(button.getAttribute('aria-expanded') !== 'true'));
    });
  });

  const activateSectionFromHash = () => {
    const section = window.location.hash.slice(1);
    if (!section) return;
    const sectionItem = [...document.querySelectorAll('.side-menu [data-section]')]
      .find((item) => item.dataset.section === section);
    if (!sectionItem) return;
    document.querySelectorAll('.side-menu > li').forEach((item) => item.classList.remove('is-active'));
    sectionItem.closest('.side-menu > li')?.classList.add('is-active');
    if (sectionItem.matches('button[aria-expanded]')) sectionItem.setAttribute('aria-expanded', 'true');
  };
  activateSectionFromHash();
  window.addEventListener('hashchange', activateSectionFromHash);

  const toast = document.querySelector('.toast');
  let toastTimer;
  const showToast = (message) => {
    if (!toast) return;
    toast.textContent = message;
    toast.classList.add('show');
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => toast.classList.remove('show'), 2200);
  };

  document.querySelectorAll('a[href="#"], a[href="#mobile"]').forEach((link) => {
    link.addEventListener('click', (event) => {
      event.preventDefault();
      showToast(link.hasAttribute('data-demo-link') ? '外部友情链接为静态演示，正式地址待配置' : '该内容为静态交互演示');
    });
  });

  document.querySelectorAll('.search-form').forEach((form) => {
    form.addEventListener('submit', (event) => {
      event.preventDefault();
      const value = form.querySelector('input')?.value.trim() || '';
      if (!value) return showToast('请输入搜索关键词');
      const candidates = [...document.querySelectorAll('.searchable')];
      let matches = 0;
      candidates.forEach((node) => {
        const hit = node.textContent.includes(value);
        node.hidden = !hit;
        if (hit) matches += 1;
      });
      showToast(candidates.length ? `找到 ${matches} 条相关演示内容` : `已记录搜索关键词：${value}`);
    });
  });

  const wireTabs = (selector) => {
    document.querySelectorAll(selector).forEach((tabList) => {
      const buttons = [...tabList.querySelectorAll('[data-tab]')];
      const activate = (button) => {
        const panelId = button.dataset.tab;
        buttons.forEach((item) => item.classList.toggle('active', item === button));
        document.querySelectorAll(`[data-tab-group="${tabList.dataset.group}"]`).forEach((panel) => {
          panel.hidden = panel.id !== panelId;
        });
      };
      buttons.forEach((button) => {
        button.addEventListener('mouseenter', () => activate(button));
        button.addEventListener('focus', () => activate(button));
        button.addEventListener('click', () => {
          activate(button);
          if (button.dataset.href) window.location.href = button.dataset.href;
        });
      });
    });
  };
  wireTabs('.home-tabs');
  wireTabs('.updates-tabs');

  document.querySelectorAll('.hero-slider').forEach((slider) => {
    const slides = [...slider.querySelectorAll('.hero-slide')];
    const dots = [...slider.querySelectorAll('[data-slide-to]')];
    if (slides.length < 2) return;

    let activeIndex = Math.max(0, slides.findIndex((slide) => slide.classList.contains('active')));
    let timer;
    const showSlide = (index) => {
      activeIndex = (index + slides.length) % slides.length;
      slides.forEach((slide, slideIndex) => {
        const isActive = slideIndex === activeIndex;
        slide.classList.toggle('active', isActive);
        slide.setAttribute('aria-hidden', String(!isActive));
      });
      dots.forEach((dot, dotIndex) => {
        const isActive = dotIndex === activeIndex;
        dot.classList.toggle('active', isActive);
        if (isActive) dot.setAttribute('aria-current', 'true');
        else dot.removeAttribute('aria-current');
      });
    };
    const stop = () => clearInterval(timer);
    const start = () => {
      stop();
      if (!document.hidden) timer = setInterval(() => showSlide(activeIndex + 1), 5000);
    };

    dots.forEach((dot) => dot.addEventListener('click', () => {
      showSlide(Number(dot.dataset.slideTo));
      start();
    }));
    slider.addEventListener('mouseenter', stop);
    slider.addEventListener('mouseleave', start);
    slider.addEventListener('focusin', stop);
    slider.addEventListener('focusout', (event) => {
      if (!slider.contains(event.relatedTarget)) start();
    });
    document.addEventListener('visibilitychange', () => document.hidden ? stop() : start());
    showSlide(activeIndex);
    start();
  });

  document.querySelectorAll('.pagination').forEach((pagination) => {
    pagination.addEventListener('click', (event) => {
      const button = event.target.closest('button[data-page]');
      if (!button) return;
      pagination.querySelectorAll('button').forEach((item) => item.classList.remove('active'));
      button.classList.add('active');
      const page = Number(button.dataset.page);
      document.querySelectorAll('.news-row a').forEach((link, index) => {
        const base = link.dataset.baseTitle || link.textContent.trim();
        link.dataset.baseTitle = base;
        if (page > 1) link.textContent = `${base}（第 ${page} 页示例 ${index + 1}）`;
        else link.textContent = base;
      });
      showToast(`已切换到第 ${page} 页（静态演示）`);
    });
  });

  document.querySelectorAll('.language-switch').forEach((button) => {
    button.addEventListener('click', () => showToast('英文版内容将在后续阶段接入'));
  });

  const modal = document.querySelector('.video-modal');
  const openModal = () => {
    if (!modal) return;
    modal.classList.add('open');
    modal.setAttribute('aria-hidden', 'false');
    body.style.overflow = 'hidden';
  };
  const closeModal = () => {
    if (!modal) return;
    modal.classList.remove('open');
    modal.setAttribute('aria-hidden', 'true');
    body.style.overflow = '';
  };
  document.querySelectorAll('[data-video-play]').forEach((button) => button.addEventListener('click', openModal));
  modal?.querySelector('.video-modal-close')?.addEventListener('click', closeModal);
  modal?.addEventListener('click', (event) => { if (event.target === modal) closeModal(); });
  document.addEventListener('keydown', (event) => { if (event.key === 'Escape') closeModal(); });

  const backToTop = document.querySelector('.back-to-top');
  window.addEventListener('scroll', () => backToTop?.classList.toggle('visible', window.scrollY > 500), { passive: true });
  backToTop?.addEventListener('click', () => window.scrollTo({ top: 0, behavior: 'smooth' }));
})();
