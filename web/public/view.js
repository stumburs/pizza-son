// Let links between top-level sections switch tabs in the shell instead of
// loading the shell inside its own iframe.
(function () {
  if (window.parent === window) return;

  var routes = {
    "/": "overview",
    "/bertstats": "berts",
    // Old location kept working for existing links.
    "/berts": "berts",
    "/quotes": "quotes",
    "/docs": "docs",
    "/vote": "vote",
    "/panel": "panel",
  };

  document.addEventListener("click", function (event) {
    var anchor = event.target.closest("a[href]");
    if (!anchor) return;
    if (anchor.target && anchor.target !== "_self") return;
    if (anchor.origin !== window.location.origin) return;

    var path = anchor.pathname.replace(/\/+$/, "") || "/";
    var tab = routes[path];
    if (!tab) return;

    event.preventDefault();
    window.parent.postMessage(
      { type: "pizzason:navigate", tab: tab },
      window.location.origin
    );
  });
})();
