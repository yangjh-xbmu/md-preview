// INPUT: Local Vite server and Playwright CLI page.
// OUTPUT: Browser assertions for syntax token backgrounds across preview themes.
// POS: Run with playwright-cli run-code --filename=frontend/tests/code-theme.browser.js.
async page => {
  const base = 'http://127.0.0.1:5173';
  await page.goto(`${base}/node_modules/prismjs/themes/prism.css`);
  await page.setContent('<main class="markdown-body"></main>');
  for (const path of [
    '/node_modules/github-markdown-css/github-markdown.css',
    '/node_modules/prismjs/themes/prism.css',
    '/src/App.css',
  ]) {
    const response = await page.request.get(`${base}${path}?raw`);
    if (!response.ok()) throw new Error(`Cannot load ${path}`);
    await page.addStyleTag({ content: await response.text() });
  }
  await page.addScriptTag({ url: `${base}/node_modules/prismjs/prism.js` });
  await page.addScriptTag({ url: `${base}/node_modules/prismjs/components/prism-json.js` });
  const results = await page.evaluate(() => {
    const root = document.querySelector('main');
    const failures = [];
    for (const theme of ['github-light', 'github-dark', 'github-sepia']) {
      root.className = `markdown-body theme-${theme}`;
      for (const [language, source] of [
        ['json', '{"course":"数据新闻","students":[{"student_id":"S001","name":"张三"}]}'],
        ['javascript', 'const count = 1 + 2;'],
        ['css', 'a { content: "hello"; background: url("image.png"); }'],
        ['markup', '<p title="text">&amp;</p>'],
      ]) {
        root.innerHTML = `<pre class="language-${language}"><code class="language-${language}"></code></pre>`;
        const code = root.querySelector('code');
        code.textContent = source;
        Prism.highlightElement(code);
        if (code.textContent !== source) failures.push(`${theme}/${language}: text changed`);
        const tokens = code.querySelectorAll('.token');
        if (!tokens.length) failures.push(`${theme}/${language}: no highlighting`);
        for (const token of tokens) {
          const background = getComputedStyle(token).backgroundColor;
          if (background !== 'rgba(0, 0, 0, 0)') {
            failures.push(`${theme}/${language} ${token.textContent}: ${background}`);
          }
        }
      }
    }
    return failures;
  });
  if (results.length) throw new Error(results.join('\n'));
  return 'PASS: 3 themes × 4 languages; transparent token backgrounds and intact source text';
}
