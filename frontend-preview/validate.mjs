import { existsSync, readFileSync, readdirSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = dirname(fileURLToPath(import.meta.url));
const htmlFiles = [join(root, 'index.html'), ...readdirSync(join(root, 'pages')).filter((name) => name.endsWith('.html')).map((name) => join(root, 'pages', name))];
const errors = [];

for (const file of htmlFiles) {
  const html = readFileSync(file, 'utf8');
  if (!html.includes('<meta name="viewport"')) errors.push(`${file}: missing viewport meta`);
  if (/[A-Z]:\\|file:\/\//i.test(html)) errors.push(`${file}: contains a local absolute path`);
  for (const match of html.matchAll(/(?:href|src)="([^"]+)"/g)) {
    const value = match[1];
    if (!value || value.startsWith('#') || /^(?:https?:|mailto:|tel:|data:)/i.test(value)) continue;
    const target = resolve(dirname(file), value.split('#')[0].split('?')[0]);
    if (!existsSync(target)) errors.push(`${file}: missing ${value}`);
  }
}

if (errors.length) {
  console.error(errors.join('\n'));
  process.exit(1);
}

console.log(`Validated ${htmlFiles.length} HTML pages: all local links and assets resolve.`);
