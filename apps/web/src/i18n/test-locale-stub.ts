// jsdom's navigator.language defaults to 'en-US'. The language detector runs
// during i18next's init() as soon as './config' is imported, so by the time
// any test-side changeLanguage('ja') call could run, detection has already
// resolved to 'en' instead of the app's 'ja' fallback. Stub the locale here,
// in a module with no further imports, so it is in place before './config'
// (imported after this in test-init.ts) triggers detection.
Object.defineProperty(window.navigator, 'language', {
  value: 'ja',
  configurable: true,
})
Object.defineProperty(window.navigator, 'languages', {
  value: ['ja'],
  configurable: true,
})
