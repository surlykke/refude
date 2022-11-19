// Simple serviceworker to cache react.production.min.js and react-dom.production.min.js.
// This to be able to run after pc-boot, before wifi-connection is established.

self.addEventListener('install', event => event.waitUntil(
  caches.open("v1") .then(v1 => {
    v1.addAll(["https://unpkg.com/react@18.2.0/umd/react.development.js", "https://unpkg.com/react-dom@18.2.0/umd/react-dom.development.js"])
  })
))

self.addEventListener('fetch', (event) => 
  event.respondWith(caches.match(event.request).then(responseFromCache => responseFromCache || fetch(event.request))));
