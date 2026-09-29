import { existsSync, readFileSync, readdirSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = dirname(fileURLToPath(import.meta.url));
const htmlFiles = [join(root, 'index.html'), ...readdirSync(join(root, 'pages')).filter((name) => name.endsWith('.html')).map((name) => join(root, 'pages', name))];
const errors = [];
const structure = JSON.parse(readFileSync(join(root, 'column-structure.json'), 'utf8'));
const allColumns = [];
const collectColumns = (columns) => columns.forEach((column) => {
  allColumns.push(column);
  if (column.children) collectColumns(column.children);
});
collectColumns(structure.navigation);
const knownCodes = new Set([...allColumns.map((column) => column.code), ...structure.homepagePlacements.map((column) => column.code)]);
if (knownCodes.size !== allColumns.length + structure.homepagePlacements.length) errors.push('column-structure.json: duplicate column code');

for (const file of htmlFiles) {
  const html = readFileSync(file, 'utf8');
  if (!html.includes('<meta name="viewport"')) errors.push(`${file}: missing viewport meta`);
  if (/[A-Z]:\\|file:\/\//i.test(html)) errors.push(`${file}: contains a local absolute path`);
  if (html.includes('policy.html#interpretation')) {
    errors.push(`${file}: 政策解读应跳转到部委动态页 ministry.html#interpretation`);
  }
  for (const form of html.matchAll(/<form class="search-form"([^>]*)>[\s\S]*?<\/form>/g)) {
    if (!/\baction="[^"]*search\.html"/.test(form[1]) || !/\bname="q"/.test(form[0])) {
      errors.push(`${file}: header search form must submit q to search.html`);
    }
  }
  for (const match of html.matchAll(/(?:href|src)="([^"]+)"/g)) {
    const value = match[1];
    if (!value || value.startsWith('#') || /^(?:https?:|mailto:|tel:|data:)/i.test(value)) continue;
    const target = resolve(dirname(file), value.split('#')[0].split('?')[0]);
    if (!existsSync(target)) errors.push(`${file}: missing ${value}`);
    if (/detail(?:-\d+)?\.html(?:[?#]|$)/.test(value) && value.startsWith('#') === false && !/[?&]from=/.test(value)) {
      errors.push(`${file}: detail link is missing its source column: ${value}`);
    }
  }
  if (file === join(root, 'index.html')) {
    const heroLinks = [...html.matchAll(/<a class="hero-slide[^>]*href="([^"]+)"/g)];
    if (heroLinks.length < 1) errors.push(`${file}: hero slider items must be clickable detail links`);
    if (!html.includes('data-column-code="home-focus"')) errors.push(`${file}: hero slider is missing its column mapping`);
    if ([...html.matchAll(/class="mobile-qr-item"/g)].length !== 3) errors.push(`${file}: mobile media must contain exactly three QR codes`);
  }
  if (/detail(?:-\d+)?\.html$/.test(file)) {
    const currentCrumb = html.match(/<span aria-current="page">([^<]+)<\/span>/)?.[1]?.trim();
    const detailTitle = html.match(/<article class="article-card[^>]*>[\s\S]*?<h1>([^<]+)<\/h1>/)?.[1]?.trim();
    if (!currentCrumb || !detailTitle || currentCrumb !== detailTitle) {
      errors.push(`${file}: detail breadcrumb must end with the current detail title`);
    }
    const defaultColumnCode = html.match(/data-detail-page[^>]*data-default-column-code="([^"]+)"/)?.[1];
    if (!defaultColumnCode || !knownCodes.has(defaultColumnCode)) errors.push(`${file}: missing or unknown default detail column`);
  }
}

const renderedHtml = htmlFiles.map((file) => readFileSync(file, 'utf8')).join('\n');
const searchFile = join(root, 'pages', 'search.html');
if (!existsSync(searchFile) || !readFileSync(searchFile, 'utf8').includes('data-search-page')) {
  errors.push('pages/search.html: missing search results page');
}
for (const column of allColumns) {
  if (!renderedHtml.includes(column.name)) errors.push(`column-structure.json: column is not represented in generated pages: ${column.name} (${column.code})`);
}
for (const match of renderedHtml.matchAll(/data-(?:source-)?column-code="([^"]+)"/g)) {
  if (!knownCodes.has(match[1])) errors.push(`generated HTML: unknown column code ${match[1]}`);
}

if (errors.length) {
  console.error(errors.join('\n'));
  process.exit(1);
}

console.log(`Validated ${htmlFiles.length} HTML pages: all local links and assets resolve.`);
