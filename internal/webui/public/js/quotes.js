
      const API_URL = "/api/quotes";
      const REFRESH_MS = 60_000;

      const state = {
        channels: [],
        total: 0,
        active: (new URLSearchParams(location.search).get("channel") || "").trim().toLowerCase(),
        query: "",
        loading: false,
      };

      const statusEl = document.getElementById("status");
      const totalEl = document.getElementById("quoteTotal");
      const tabsEl = document.getElementById("channelTabs");
      const listEl = document.getElementById("quoteList");
      const searchEl = document.getElementById("search");
      const refreshEl = document.getElementById("refresh");

      function escapeHtml(value) {
        return String(value ?? "").replace(/[&<>"']/g, (char) => {
          switch (char) {
            case "&":
              return "&amp;";
            case "<":
              return "&lt;";
            case ">":
              return "&gt;";
            case '"':
              return "&quot;";
            default:
              return "&#39;";
          }
        });
      }

      function formatCount(value) {
        return new Intl.NumberFormat().format(Number(value) || 0);
      }

      function selectedChannel() {
        return state.channels.find((channel) => channel.name === state.active);
      }

      function renderTabs() {
        tabsEl.innerHTML = state.channels
          .map(
            (channel) => `
            <button
              class="channel-tab${channel.name === state.active ? " active" : ""}"
              type="button"
              data-channel="${escapeHtml(channel.name)}"
              aria-pressed="${channel.name === state.active}"
            >
              <span>${escapeHtml(channel.name)}</span>
              <span class="count">${formatCount(channel.count)}</span>
            </button>`
          )
          .join("");

        totalEl.textContent = state.total
          ? `${formatCount(state.total)} total`
          : "";
      }

      function visibleQuotes() {
        const channel = selectedChannel();
        if (!channel) return [];

        const query = state.query.trim().toLowerCase();
        return channel.quotes.filter((quote) => {
          if (!query) return true;
          return (
            quote.text.toLowerCase().includes(query) ||
            (quote.added_by || "").toLowerCase().includes(query)
          );
        });
      }

      function renderQuotes() {
        const entries = visibleQuotes();

        if (entries.length === 0) {
          let message = "No quotes have been saved yet.";
          if (state.total > 0) {
            message = state.query.trim()
              ? "No quotes match that search."
              : "No quotes in this channel yet.";
          }
          listEl.innerHTML = `<div class="empty">${message}</div>`;
          return;
        }

        listEl.innerHTML = entries
          .map(
            (quote) => `
            <article class="quote">
              <span class="quote-index">${quote.number}</span>
              <div>
                <p class="quote-text">${quote.html}</p>
                <div class="quote-meta">
                  <span class="added-by">${escapeHtml(quote.added_by || "unknown")}</span>
                  ${quote.created_at ? `<span class="dot">·</span><span>${escapeHtml(quote.created_at)}</span>` : ""}
                </div>
              </div>
            </article>`
          )
          .join("");
      }

      function setStatus(message, isError) {
        statusEl.textContent = message;
        statusEl.classList.toggle("error", Boolean(isError));
      }

      async function load(initial) {
        if (state.loading) return;
        state.loading = true;
        refreshEl.disabled = true;
        if (initial) setStatus("Loading quotes…", false);

        try {
          const response = await fetch(API_URL, { cache: "no-store" });
          if (!response.ok) throw new Error(`HTTP ${response.status}`);

          const data = await response.json();
          state.channels = Array.isArray(data.channels) ? data.channels : [];
          state.total = Number(data.total) || 0;

          if (!selectedChannel()) {
            state.active = state.channels[0]?.name || "";
          }

          renderTabs();
          renderQuotes();

          const stamp = data.generated_at
            ? new Date(data.generated_at).toLocaleTimeString([], {
                hour: "2-digit",
                minute: "2-digit",
              })
            : "";
          setStatus(stamp ? `Updated ${stamp}` : "Updated", false);
        } catch (error) {
          setStatus("Could not load quotes. Retrying automatically.", true);
          if (state.total === 0) {
            listEl.innerHTML = `<div class="empty">Quotes could not be loaded. They will retry automatically.</div>`;
          }
        } finally {
          state.loading = false;
          refreshEl.disabled = false;
        }
      }

      tabsEl.addEventListener("click", (event) => {
        const button = event.target.closest("[data-channel]");
        if (!button) return;
        state.active = button.dataset.channel;
        renderTabs();
        renderQuotes();
      });

      searchEl.addEventListener("input", () => {
        state.query = searchEl.value;
        renderQuotes();
      });

      refreshEl.addEventListener("click", () => load(false));

      document.addEventListener("visibilitychange", () => {
        if (!document.hidden) load(false);
      });

      setInterval(() => {
        if (!document.hidden) load(false);
      }, REFRESH_MS);

      load(true);
    