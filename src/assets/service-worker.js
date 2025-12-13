const BASE_PATH = new URL(self.registration.scope).pathname.replace(/\/$/, '');
const CACHE_VERSION = 'v1';
const CACHE_NAME = `readn-${CACHE_VERSION}`;
const PRECACHE_URLS = [
  `${BASE_PATH}/`,
  `${BASE_PATH}/static/stylesheets/bootstrap.min.css`,
  `${BASE_PATH}/static/stylesheets/app.css`,
  `${BASE_PATH}/static/javascripts/vue.min.js`,
  `${BASE_PATH}/static/javascripts/marked.min.js`,
  `${BASE_PATH}/static/javascripts/app.js`,
  `${BASE_PATH}/static/javascripts/api.js`,
  `${BASE_PATH}/static/javascripts/key.js`,
  `${BASE_PATH}/static/graphicarts/favicon.png`,
  `${BASE_PATH}/static/graphicarts/icon-192.png`,
  `${BASE_PATH}/static/graphicarts/icon-512.png`,
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(PRECACHE_URLS)),
  );
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) =>
        Promise.all(
          keys
            .filter((key) => key.startsWith('readn-') && key !== CACHE_NAME)
            .map((key) => caches.delete(key)),
        ),
      ),
  );
  self.clients.claim();
});

self.addEventListener('fetch', (event) => {
  const { request } = event;
  if (request.method !== 'GET') {
    return;
  }

  const url = new URL(request.url);
  if (url.origin !== self.location.origin) {
    return;
  }

  const path = url.pathname;
  const inScope =
    BASE_PATH === '' ? true : path === BASE_PATH || path.startsWith(`${BASE_PATH}/`);
  if (!inScope) {
    return;
  }

  const cacheable =
    path === `${BASE_PATH}/` ||
    path === BASE_PATH ||
    path.startsWith(`${BASE_PATH}/static/`);

  if (!cacheable) {
    return;
  }

  event.respondWith(
    caches.match(request).then((cached) => {
      if (cached) {
        return cached;
      }
      return fetch(request)
        .then((response) => {
          if (response && response.ok) {
            const copy = response.clone();
            caches.open(CACHE_NAME).then((cache) => cache.put(request, copy));
          }
          return response;
        })
        .catch(() => cached);
    }),
  );
});
