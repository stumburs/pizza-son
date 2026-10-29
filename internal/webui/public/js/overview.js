
      const numberFormat = new Intl.NumberFormat();

      const htmlEscape = (value) =>
        String(value ?? "").replace(/[&<>"']/g, (char) => ({
          "&": "&amp;",
          "<": "&lt;",
          ">": "&gt;",
          '"': "&quot;",
          "'": "&#39;",
        })[char]);

      function formatNumber(value) {
        return numberFormat.format(Number(value) || 0);
      }

      function formatUptime(seconds) {
        const total = Math.max(0, Number(seconds) || 0);
        const days = Math.floor(total / 86400);
        const hours = Math.floor((total % 86400) / 3600);
        const minutes = Math.floor((total % 3600) / 60);
        if (days > 0) return `${days}d ${hours}h`;
        if (hours > 0) return `${hours}h ${minutes}m`;
        return `${minutes}m`;
      }

      function formatHours(seconds) {
        const hours = (Number(seconds) || 0) / 3600;
        if (hours >= 100) return `${numberFormat.format(Math.round(hours))} h`;
        return `${hours.toFixed(1)} h`;
      }

      function formatUpdated(iso) {
        if (!iso) return "loading";
        const date = new Date(iso);
        if (Number.isNaN(date.getTime())) return "loading";
        return `updated ${date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`;
      }

      function renderChannels(channels) {
        const body = document.getElementById("channelBody");
        if (!channels || channels.length === 0) {
          body.innerHTML = `
            <tr>
              <td colspan="7" style="text-align: center; color: var(--text-secondary)">
                No channels recorded yet.
              </td>
            </tr>`;
          return;
        }

        body.innerHTML = channels
          .map((channel) => {
            const name = htmlEscape(channel.name);
            const live = Boolean(channel.live);
            return `
              <tr>
                <td>
                  <a
                    class="channel-name"
                    href="https://twitch.tv/${encodeURIComponent(channel.name)}"
                    target="_blank"
                    rel="noopener"
                  >
                    <span class="channel-dot ${live ? "live" : ""}"></span>
                    <span>${name}</span>
                  </a>
                </td>
                <td>${formatNumber(channel.messages)}</td>
                <td>${formatNumber(channel.commands)}</td>
                <td>${formatNumber(channel.listeners)}</td>
                <td>${formatNumber(channel.users)}</td>
                <td>${formatNumber(channel.berts)}</td>
                <td>${formatHours(channel.streamed_seconds)}</td>
              </tr>`;
          })
          .join("");
      }

      function thumbnailFor(channel) {
        if (channel.thumbnail) {
          return channel.thumbnail
            .replace("{width}", "640")
            .replace("{height}", "360");
        }
        return `https://static-cdn.jtvnw.net/previews-ttv/live_user_${encodeURIComponent(channel.name)}-640x360.jpg`;
      }

      function renderLive(channels) {
        const liveChannels = (channels || []).filter((channel) => channel.live);
        document.getElementById("liveCount").textContent =
          liveChannels.length > 0
            ? `${liveChannels.length} streaming`
            : "none streaming";

        const grid = document.getElementById("liveGrid");
        if (liveChannels.length === 0) {
          grid.innerHTML = `<div class="empty">Nobody from the channel list is live right now.</div>`;
          return;
        }

        grid.innerHTML = liveChannels
          .map((channel) => {
            const name = htmlEscape(channel.name);
            const title = htmlEscape(channel.title || "Untitled stream");
            const game = htmlEscape(channel.game || "No category");
            return `
              <a
                class="live-card"
                href="https://twitch.tv/${encodeURIComponent(channel.name)}"
                target="_blank"
                rel="noopener"
              >
                <div class="live-thumb">
                  <img
                    src="${htmlEscape(thumbnailFor(channel))}"
                    alt="${name} stream preview"
                    loading="lazy"
                    onerror="this.style.display='none'"
                  />
                  <span class="live-badge">live</span>
                  <span class="viewers">${formatNumber(channel.viewers)} viewers</span>
                </div>
                <div class="live-body">
                  <span class="live-channel">#${name}</span>
                  <p class="live-title">${title}</p>
                  <div class="live-meta">
                    <span class="game">${game}</span>
                    <span>${formatUptime(channel.uptime_seconds)}</span>
                  </div>
                </div>
              </a>`;
          })
          .join("");
      }

      function renderLeaderboard(elementId, entries, prefix = "") {
        const element = document.getElementById(elementId);
        if (!entries || entries.length === 0) {
          element.innerHTML = `<div class="empty" style="padding: 16px">Nothing here yet.</div>`;
          return;
        }

        element.innerHTML = entries
          .map(
            (entry, index) => `
              <div class="leader-row">
                <span class="leader-rank">${index + 1}</span>
                <span class="leader-name" title="${htmlEscape(entry.name)}">${prefix}${htmlEscape(entry.name)}</span>
                <span class="leader-count">${formatNumber(entry.count)}</span>
              </div>`
          )
          .join("");
      }

      function render(data) {
        const totals = data.totals || {};
        document.getElementById("statChannels").textContent = formatNumber(
          data.bot?.twitch_channels ?? totals.channels
        );
        document.getElementById("statDiscord").textContent = formatNumber(
          data.bot?.discord_channels
        );
        document.getElementById("statLive").textContent = formatNumber(totals.live_channels);
        document.getElementById("statStreamed").textContent = formatHours(totals.streamed_seconds);
        document.getElementById("statMessages").textContent = formatNumber(totals.messages);
        document.getElementById("statCommands").textContent = formatNumber(totals.commands);
        document.getElementById("statListeners").textContent = formatNumber(
          totals.listener_activations
        );
        document.getElementById("statUsers").textContent = formatNumber(totals.unique_users);
        document.getElementById("statBerts").textContent = formatNumber(totals.berts);

        renderLive(data.channels);
        renderChannels(data.channels);
        renderLeaderboard("topBerts", data.top_berts);
        renderLeaderboard("topCommands", data.top_commands, "!");
        renderLeaderboard("topListeners", data.top_listeners);

        document.getElementById("updatedAt").textContent = formatUpdated(data.generated_at);
      }

      async function loadSummary() {
        try {
          const response = await fetch("/api/site/summary", { cache: "no-store" });
          if (!response.ok) throw new Error(`HTTP ${response.status}`);
          render(await response.json());
        } catch (error) {
          document.getElementById("updatedAt").textContent = "update failed";
        }
      }

      loadSummary();
      setInterval(loadSummary, 30000);

      document.addEventListener("visibilitychange", () => {
        if (!document.hidden) loadSummary();
      });
    