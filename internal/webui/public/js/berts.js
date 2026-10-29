
      let globalChartInstance = null;
      let globalHourlyCumulativeChartInstance = null;
      let channelChartInstance = null;
      let channelHourlyChartInstance = null;
      let channelHourlyCumulativeChartInstance = null;
      let userChartInstance = null;
      let userHourlyChartInstance = null;
      let userHourlyCumulativeChartInstance = null;

      window.globalEmoteMemoryCache = {};
      window.channelListCache = [];
      window.userListCache = [];

      const ZAZA_EMOTE = "https://cdn.7tv.app/emote/01JBHY1T7RWKHNMC4TV7C7A54T/2x.webp";
      const ZAZAL_EMOTE = "https://cdn.7tv.app/emote/01KRRDPG9MCPC4NEAHJSBQDPPY/2x.webp";

      function zazaEmoteHTML(cls) {
        return `<img src="${ZAZA_EMOTE}" style="height:1.1em;vertical-align:middle;" class="${cls || ""}" />`;
      }
      function zazaLEmoteHTML(cls) {
        return `<img src="${ZAZAL_EMOTE}" style="height:1.1em;vertical-align:middle;" class="${cls || ""}" />`;
      }
      function zazaDoubleEmoteHTML() {
        return `<span style="position:relative;display:inline-block;width:3.2em;height:1.1em;vertical-align:middle;"><img src="${ZAZA_EMOTE}" style="height:1.1em;position:absolute;left:0;top:0;" /><img src="${ZAZAL_EMOTE}" style="height:1.1em;position:absolute;left:0.8em;top:0;" /></span>`;
      }

      const COMPOSITE_BERTS = {
        "BigBert": ["bigbert1", "bigbert2"],
      };

      function renderBertEmote(bert, emotesMap, extraClass) {
        const parts = COMPOSITE_BERTS[bert];
        if (parts) {
          const ids = parts.map(p => emotesMap[p]).filter(Boolean);
          if (ids.length === parts.length) {
            return `<div class="emote-wrapper" style="display:inline-flex;gap:0;">${ids.map(id => `<img src="https://cdn.7tv.app/emote/${id}/2x.webp" alt="${bert}" class="twitch-emote ${extraClass || ''}" style="height:32px;" />`).join("")}<span class="emote-tooltip">${bert}</span></div>`;
          }
        }
        const emId = emotesMap[bert];
        if (emId) {
          return `<div class="emote-wrapper"><img src="https://cdn.7tv.app/emote/${emId}/2x.webp" alt="${bert}" class="twitch-emote ${extraClass || ''}" /><span class="emote-tooltip">${bert}</span></div>`;
        }
        return null;
      }

      function switchTab(tabId, btn) {
        document.querySelectorAll(".panel").forEach(p => p.classList.remove("active"));
        document.querySelectorAll(".tab").forEach(t => t.classList.remove("active"));
        document.getElementById(tabId).classList.add("active");
        btn.classList.add("active");
      }

      function getCountValue(value) {
        if (typeof value === "object" && value !== null) {
          return value.count !== undefined ? value.count : value.Count !== undefined ? value.Count : 0;
        }
        return typeof value === "number" ? value : 0;
      }

      function navigateTo(type, name) {
        if (type === "channel") {
          const tab = [...document.querySelectorAll(".tab")].find(t => t.textContent.includes("Channel"));
          switchTab("channel", tab);
          document.getElementById("chInput").value = name;
          loadChannel();
        } else if (type === "user") {
          const tab = [...document.querySelectorAll(".tab")].find(t => t.textContent.includes("User"));
          switchTab("user", tab);
          document.getElementById("uInput").value = name;
          loadUser();
        }
      }

      let chAutocompleteIdx = -1;

      function filterChannelAutocomplete() {
        const input = document.getElementById("chInput");
        const dropdown = document.getElementById("chAutocomplete");
        const query = input.value.toLowerCase().trim();

        if (!query && window.channelListCache.length === 0) {
          dropdown.classList.remove("open");
          return;
        }

        const matches = window.channelListCache.filter(c => c.name.includes(query));
        if (matches.length === 0) {
          dropdown.classList.remove("open");
          return;
        }

        dropdown.innerHTML = matches.slice(0, 20).map((c, i) =>
          `<div class="ch-autocomplete-item" data-channel="${c.name}" onclick="selectChannel('${c.name}')">
            <span>#${c.name}</span>
            <small>x${c.total}</small>
          </div>`
        ).join("");
        dropdown.classList.add("open");
        chAutocompleteIdx = -1;
      }

      function selectChannel(name) {
        document.getElementById("chInput").value = name;
        document.getElementById("chAutocomplete").classList.remove("open");
        loadChannel();
      }

      function clearInput(inputId, autocompleteId, filterFn) {
        const input = document.getElementById(inputId);
        input.value = "";
        input.focus();
        if (autocompleteId) document.getElementById(autocompleteId).classList.remove("open");
        if (filterFn) setTimeout(filterFn, 0);
      }

      function handleChannelKeydown(e) {
        const dropdown = document.getElementById("chAutocomplete");
        const items = dropdown.querySelectorAll(".ch-autocomplete-item");
        if (!dropdown.classList.contains("open") || items.length === 0) {
          if (e.key === "Enter") { e.preventDefault(); loadChannel(); }
          return;
        }

        if (e.key === "ArrowDown") {
          e.preventDefault();
          chAutocompleteIdx = Math.min(chAutocompleteIdx + 1, items.length - 1);
          items.forEach((el, i) => el.classList.toggle("active", i === chAutocompleteIdx));
        } else if (e.key === "ArrowUp") {
          e.preventDefault();
          chAutocompleteIdx = Math.max(chAutocompleteIdx - 1, 0);
          items.forEach((el, i) => el.classList.toggle("active", i === chAutocompleteIdx));
        } else if (e.key === "Enter") {
          e.preventDefault();
          if (chAutocompleteIdx >= 0) selectChannel(items[chAutocompleteIdx].dataset.channel);
          else loadChannel();
        } else if (e.key === "Escape") {
          dropdown.classList.remove("open");
        }
      }

      // ── User channel autocomplete ──
      let uChAutocompleteIdx = -1;

      function filterUserChannelAutocomplete() {
        const input = document.getElementById("uChannelInput");
        const dropdown = document.getElementById("uChannelAutocomplete");
        const query = input.value.toLowerCase().trim();
        const list = window.userChannelListCache || [];

        if (!query && list.length === 0) {
          dropdown.classList.remove("open");
          return;
        }

        const matches = list.filter(c => c.name.includes(query));
        if (matches.length === 0) {
          dropdown.classList.remove("open");
          return;
        }

        dropdown.innerHTML = matches.slice(0, 20).map((c) =>
          `<div class="ch-autocomplete-item" data-channel="${c.name}" onclick="selectUserChannel('${c.name}')">
            <span>#${c.name}</span>
            <small>x${c.total}</small>
          </div>`
        ).join("");
        dropdown.classList.add("open");
        uChAutocompleteIdx = -1;
      }

      function selectUserChannel(name) {
        document.getElementById("uChannelInput").value = name;
        document.getElementById("uChannelAutocomplete").classList.remove("open");
        loadUserChannelSelection();
      }

      function handleUserChannelKeydown(e) {
        const dropdown = document.getElementById("uChannelAutocomplete");
        const items = dropdown.querySelectorAll(".ch-autocomplete-item");
        if (!dropdown.classList.contains("open") || items.length === 0) {
          if (e.key === "Enter") { e.preventDefault(); loadUserChannelSelection(); }
          return;
        }

        if (e.key === "ArrowDown") {
          e.preventDefault();
          uChAutocompleteIdx = Math.min(uChAutocompleteIdx + 1, items.length - 1);
          items.forEach((el, i) => el.classList.toggle("active", i === uChAutocompleteIdx));
        } else if (e.key === "ArrowUp") {
          e.preventDefault();
          uChAutocompleteIdx = Math.max(uChAutocompleteIdx - 1, 0);
          items.forEach((el, i) => el.classList.toggle("active", i === uChAutocompleteIdx));
        } else if (e.key === "Enter") {
          e.preventDefault();
          if (uChAutocompleteIdx >= 0) selectUserChannel(items[uChAutocompleteIdx].dataset.channel);
          else loadUserChannelSelection();
        } else if (e.key === "Escape") {
          dropdown.classList.remove("open");
        }
      }

      async function loadUserChannelSelection() {
        if (!currentUserData) return;
        const chan = document.getElementById("uChannelInput").value.toLowerCase().trim();
        const channelStats = currentUserData.channels[chan];
        if (!channelStats) return;
        updateUserChannelView();
      }

      document.addEventListener("click", function(e) {
        if (!e.target.closest(".search-row")) {
          document.getElementById("chAutocomplete").classList.remove("open");
          document.getElementById("uAutocomplete").classList.remove("open");
          document.getElementById("uChannelAutocomplete").classList.remove("open");
        }
      });

      // ── User autocomplete ──
      let uAutocompleteIdx = -1;

      function filterUserAutocomplete() {
        const input = document.getElementById("uInput");
        const dropdown = document.getElementById("uAutocomplete");
        const query = input.value.toLowerCase().trim();

        if (!query && window.userListCache.length === 0) {
          dropdown.classList.remove("open");
          return;
        }

        const matches = window.userListCache.filter(c => c.name.includes(query));
        if (matches.length === 0) {
          dropdown.classList.remove("open");
          return;
        }

        dropdown.innerHTML = matches.slice(0, 20).map((c, i) =>
          `<div class="u-autocomplete-item" data-user="${c.name}" onclick="selectUser('${c.name}')">
            <span>${c.name}</span>
            <small>x${c.total}</small>
          </div>`
        ).join("");
        dropdown.classList.add("open");
        uAutocompleteIdx = -1;
      }

      function selectUser(name) {
        document.getElementById("uInput").value = name;
        document.getElementById("uAutocomplete").classList.remove("open");
        loadUser();
      }

      function handleUserKeydown(e) {
        const dropdown = document.getElementById("uAutocomplete");
        const items = dropdown.querySelectorAll(".u-autocomplete-item");
        if (!dropdown.classList.contains("open") || items.length === 0) {
          if (e.key === "Enter") { e.preventDefault(); loadUser(); }
          return;
        }

        if (e.key === "ArrowDown") {
          e.preventDefault();
          uAutocompleteIdx = Math.min(uAutocompleteIdx + 1, items.length - 1);
          items.forEach((el, i) => el.classList.toggle("active", i === uAutocompleteIdx));
        } else if (e.key === "ArrowUp") {
          e.preventDefault();
          uAutocompleteIdx = Math.max(uAutocompleteIdx - 1, 0);
          items.forEach((el, i) => el.classList.toggle("active", i === uAutocompleteIdx));
        } else if (e.key === "Enter") {
          e.preventDefault();
          if (uAutocompleteIdx >= 0) selectUser(items[uAutocompleteIdx].dataset.user);
          else loadUser();
        } else if (e.key === "Escape") {
          dropdown.classList.remove("open");
        }
      }

      function pad2(n) { return String(n).padStart(2, "0"); }

      function fillHourlyTimeline(timeline) {
        const keys = Object.keys(timeline);
        if (keys.length === 0) return {};
        const sorted = keys.sort();
        const start = new Date(sorted[0]);
        const end = new Date(sorted[sorted.length - 1]);
        const result = {};
        const cur = new Date(start);
        while (cur <= end) {
          const key = cur.getFullYear() + "-" + pad2(cur.getMonth() + 1) + "-" + pad2(cur.getDate()) + " " + pad2(cur.getHours()) + ":00";
          result[key] = key in timeline ? timeline[key] : 0;
          cur.setHours(cur.getHours() + 1);
        }
        return result;
      }

      function fillDailyTimeline(timeline) {
        const keys = Object.keys(timeline);
        if (keys.length === 0) return {};
        const sorted = keys.sort();
        const start = new Date(sorted[0]);
        const end = new Date(sorted[sorted.length - 1]);
        const result = {};
        const cur = new Date(start);
        while (cur <= end) {
          const key = cur.getFullYear() + "-" + pad2(cur.getMonth() + 1) + "-" + pad2(cur.getDate());
          result[key] = key in timeline ? timeline[key] : 0;
          cur.setDate(cur.getDate() + 1);
        }
        return result;
      }

      function getLast24Hours(timeline) {
        const result = {};
        const now = new Date();
        now.setMinutes(0, 0, 0);
        for (let i = 23; i >= 0; i--) {
          const d = new Date(now);
          d.setHours(d.getHours() - i);
          const key = d.getFullYear() + "-" + pad2(d.getMonth() + 1) + "-" + pad2(d.getDate()) + " " + pad2(d.getHours()) + ":00";
          result[key] = key in timeline ? timeline[key] : 0;
        }
        return result;
      }

      function getTopHTML(obj, limit, nameType) {
        if (!obj) return "<i>None</i>";
        let sorted = Object.entries(obj)
          .map(([name, val]) => [name, getCountValue(val)])
          .sort((a, b) => b[1] - a[1])
          .slice(0, limit);
        if (sorted.length === 0) return "<i>None</i>";
        return sorted
          .map(([name, count], i) =>
            `<span class="badge">#${i + 1} <span class="link-name" onclick="navigateTo('${nameType}', '${name}')">${name}</span> <small>x${count}</small></span>`
          )
          .join("<br>");
      }

      function formatDate(iso) {
        if (!iso) return "?";
        return new Date(iso).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
      }

      function formatDateTime(iso) {
        if (!iso) return "?";
        const d = new Date(iso);
        return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" }) + " " +
               d.toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit" });
      }

      function renderGoldenzazaHall(events) {
        if (!events || events.length === 0) return "<i>Nobody yet</i>";
        return '<div class="hall-list">' +
          events.map((e, i) =>
            `<div class="hall-entry">
              <span class="hall-rank">#${i + 1}</span>
              <span class="hall-user link-name" onclick="navigateTo('user', '${e.username}')">${e.username}</span>
              <span class="hall-channel link-channel" onclick="navigateTo('channel', '${(e.channel || "").toLowerCase()}')">#${(e.channel || "").toLowerCase()}</span>
              <span class="hall-time">${formatDateTime(e.timestamp)}</span>
            </div>`
          ).join("") + "</div>";
      }

      function renderFirstDiscover(data) {
        if (!data) return "<i>None</i>";
        let entries = [];
        for (const [ch, berts] of Object.entries(data)) {
          for (const [bert, info] of Object.entries(berts)) {
            entries.push({ bert, user: info.user, timestamp: info.timestamp, channel: ch });
          }
        }
        if (entries.length === 0) return "<i>None</i>";
        entries.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
        return '<div class="first-list">' +
          entries.slice(0, 15).map(e =>
            `<div class="first-entry">
              <span class="first-bert">${e.bert}</span>
              <span class="first-user link-name" onclick="navigateTo('user', '${e.user}')">${e.user}</span>
              <span class="link-channel" onclick="navigateTo('channel', '${e.channel}')">#${e.channel}</span>
              <span class="first-date">${formatDate(e.timestamp)}</span>
            </div>`
          ).join("") + "</div>";
      }

      function connectLiveFeed() {
        const eventSource = new EventSource("/api/live");
        eventSource.onmessage = function (event) {
          try {
            const data = JSON.parse(event.data);
            const emptyLabel = document.getElementById("feedEmptyText");
            if (emptyLabel) emptyLabel.remove();
            const feedList = document.getElementById("liveFeedList");
            const item = document.createElement("div");
            item.className = "feed-item";

            let bertRender = `<span class="badge">${data.bert}</span>`;
            if (data.emote_ids && data.emote_ids.length >= 2) {
              bertRender = `<div class="emote-wrapper" style="display:inline-flex;gap:0;">${data.emote_ids.map(id => `<img src="https://cdn.7tv.app/emote/${id}/2x.webp" alt="${data.bert}" class="twitch-emote" style="height:32px;" />`).join("")}<span class="emote-tooltip">${data.bert}</span></div>`;
            } else {
              let emoteId = data.emote_id || window.globalEmoteMemoryCache[data.bert];
              if (emoteId) {
                bertRender = `<div class="emote-wrapper"><img src="https://cdn.7tv.app/emote/${emoteId}/2x.webp" alt="${data.bert}" class="twitch-emote" /><span class="emote-tooltip">${data.bert}</span></div>`;
              }
            }
            let zazaOverlay = "";
            if (data.is_zaza && data.is_zaza_l) {
              zazaOverlay = `<span style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;pointer-events:none;"><img src="${ZAZA_EMOTE}" style="height:32px;position:absolute;left:2px;top:3px;" /><img src="${ZAZAL_EMOTE}" style="height:32px;position:absolute;left:10px;top:3px;" /></span>`;
            } else if (data.is_zaza) {
              zazaOverlay = `<span style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;pointer-events:none;"><img src="${ZAZA_EMOTE}" style="height:32px;top:3px;" /></span>`;
            } else if (data.is_zaza_l) {
              zazaOverlay = `<span style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;pointer-events:none;"><img src="${ZAZAL_EMOTE}" style="height:32px;top:3px;" /></span>`;
            }
            const timestamp = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" });

            item.innerHTML = `
              <div class="feed-meta">
                <span class="link-channel" onclick="navigateTo('channel', '${data.channel.toLowerCase()}')">#${data.channel.toLowerCase()}</span>
                <span class="feed-time">${timestamp}</span>
              </div>
              <div>
                <strong style="color: var(--accent);" class="link-name" onclick="navigateTo('user', '${data.username}')">${data.username}</strong> ${bertRender}${zazaOverlay}
              </div>`;
            feedList.insertBefore(item, feedList.firstChild);
            while (feedList.children.length > 25) feedList.removeChild(feedList.lastChild);
          } catch (e) {}
        };
        eventSource.onerror = function () {};
      }

      // ── GLOBAL ──
      async function loadGlobal() {
        const res = await fetch("/api/global");
        if (!res.ok) return;
        const data = await res.json();

        document.getElementById("gTotal").innerText = data.total_activations;
        document.getElementById("gChans").innerText = data.channel_count;
        document.getElementById("gZazas").innerHTML = zazaEmoteHTML() + " <strong>" + (data.total_zazas || 0) + "</strong>";
        document.getElementById("gZazaLs").innerHTML = zazaLEmoteHTML() + " <strong>" + (data.total_zaza_ls || 0) + "</strong>";
        document.getElementById("gDoubleZazas").innerHTML = zazaDoubleEmoteHTML() + " <strong>" + (data.total_double_zazas || 0) + "</strong>";
        document.getElementById("gUsers").innerHTML = getTopHTML(data.users, 10, "user");
        document.getElementById("gBerts").innerHTML = getTopHTML(data.berts, 10, "bert");

        window.channelListCache = Object.entries(data.channel_totals || {})
          .map(([name, total]) => ({ name, total }))
          .sort((a, b) => b.total - a.total);

        window.userListCache = Object.entries(data.users || {})
          .map(([name, total]) => ({ name, total }))
          .sort((a, b) => b.total - a.total);

        onNewChannelData();

        const sortedBerts = Object.entries(data.berts || {})
          .map(([name, val]) => [name, getCountValue(val)])
          .sort((a, b) => b[1] - a[1])
          .slice(0, 7);

        if (globalChartInstance) globalChartInstance.destroy();
        globalChartInstance = new Chart(document.getElementById("globalChart").getContext("2d"), {
          type: "bar",
          data: {
            labels: sortedBerts.map(b => b[0]),
            datasets: [{
              label: "Total Drops Globally",
              data: sortedBerts.map(b => b[1]),
              backgroundColor: "rgba(145, 70, 255, 0.6)",
              borderColor: "#9146ff",
              borderWidth: 1,
              borderRadius: 4,
            }],
          },
          options: {
            indexAxis: "y",
            responsive: true,
            plugins: { legend: { display: false } },
            scales: {
              x: { grid: { color: "#1e1e22" }, ticks: { color: "#90909b", stepSize: 1 } },
              y: { grid: { display: false }, ticks: { color: "#90909b" } },
            },
          },
        });

        const globalHourlyTimeline = fillHourlyTimeline(data.hourly_timeline || {});
        const allGlobalSortedHours = Object.keys(globalHourlyTimeline);
        const gHourlyTimelineSum = Object.values(globalHourlyTimeline).reduce((s, v) => s + v, 0);
        let gRunningTotal = (data.total_activations || 0) - gHourlyTimelineSum;

        if (globalHourlyCumulativeChartInstance) globalHourlyCumulativeChartInstance.destroy();
        globalHourlyCumulativeChartInstance = new Chart(document.getElementById("globalHourlyCumulativeChart").getContext("2d"), {
          type: "line",
          data: {
            labels: allGlobalSortedHours,
            datasets: [{
              label: "Total Accumulated Berts",
              data: allGlobalSortedHours.map(h => { gRunningTotal += globalHourlyTimeline[h]; return gRunningTotal; }),
              borderColor: "#9146ff",
              backgroundColor: "rgba(145, 70, 255, 0.08)",
              borderWidth: 2,
              pointRadius: allGlobalSortedHours.length > 50 ? 0 : 2,
              tension: 0.1,
              fill: true,
            }],
          },
          options: {
            responsive: true,
            plugins: { legend: { display: false } },
            scales: {
              x: { grid: { color: "#161618" }, ticks: { color: "#90909b", maxTicksLimit: 12 } },
              y: { grid: { color: "#1e1e22" }, ticks: { color: "#90909b" } },
            },
          },
        });

        document.getElementById("gFirstDiscover").innerHTML = renderFirstDiscover(data.first_person_per_bert);
      }

      // ── CHANNEL ──
      async function loadChannel() {
        const name = document.getElementById("chInput").value;
        if (!name) return;
        document.getElementById("chAutocomplete").classList.remove("open");
        const res = await fetch(`/api/channel?name=${name}`);
        if (!res.ok) { alert("Channel not found!"); return; }
        const data = await res.json();

        let emotesMap = {};
        try {
          const emoteRes = await fetch(`/api/emotes?channel=${name}`);
          if (emoteRes.ok) {
            (await emoteRes.json()).forEach(em => {
              emotesMap[em.name] = em.id;
              window.globalEmoteMemoryCache[em.name] = em.id;
            });
          }
        } catch (err) {}
        window.currentChannelEmotes = emotesMap;

        document.getElementById("chResults").style.display = "block";
        document.getElementById("chTitle").innerText = `Stats for #${data.channel}`;
        document.getElementById("chTotal").innerText = data.total_activations;
        document.getElementById("chZazas").innerHTML = zazaEmoteHTML() + " <strong>" + (data.total_zazas || 0) + "</strong>";
        document.getElementById("chZazaLs").innerHTML = zazaLEmoteHTML() + " <strong>" + (data.total_zaza_ls || 0) + "</strong>";
        document.getElementById("chDoubleZazas").innerHTML = zazaDoubleEmoteHTML() + " <strong>" + (data.total_double_zazas || 0) + "</strong>";
        document.getElementById("chUsers").innerHTML = getTopHTML(data.users, 10, "user");

        document.getElementById("chBertsList").innerHTML = (data.active_berts || [])
          .map(b => {
            return renderBertEmote(b, emotesMap) || `<span class="badge">${b}</span>`;
          }).join(" ");

        const timeline = fillDailyTimeline(data.daily_activations || data.daily_timeline || {});
        const sortedDates = Object.keys(timeline);

        if (channelChartInstance) channelChartInstance.destroy();
        channelChartInstance = new Chart(document.getElementById("channelChart").getContext("2d"), {
          type: "line",
          data: {
            labels: sortedDates,
            datasets: [{
              label: "Daily Interactions",
              data: sortedDates.map(d => timeline[d]),
              borderColor: "#10b981",
              backgroundColor: "rgba(16, 185, 129, 0.1)",
              borderWidth: 2,
              tension: 0.2,
              fill: true,
            }],
          },
          options: {
            responsive: true,
            plugins: { legend: { display: false } },
            scales: {
              x: { grid: { color: "#1e1e22" }, ticks: { color: "#90909b" } },
              y: { grid: { color: "#1e1e22" }, ticks: { color: "#90909b", stepSize: 1 } },
            },
          },
        });

        const hourlyTimeline = data.hourly_activations || data.hourly_timeline || {};
        const last24Data = getLast24Hours(hourlyTimeline);
        const last24HoursLabels = Object.keys(last24Data);

        if (channelHourlyChartInstance) channelHourlyChartInstance.destroy();
        channelHourlyChartInstance = new Chart(document.getElementById("channelHourlyChart").getContext("2d"), {
          type: "bar",
          data: {
            labels: last24HoursLabels.map(h => h.split(" ")[1] || h),
            datasets: [{
              label: "Hourly Hits",
              data: last24HoursLabels.map(h => last24Data[h]),
              backgroundColor: "rgba(59, 130, 246, 0.6)",
              borderColor: "#3b82f6",
              borderWidth: 1,
              borderRadius: 3,
            }],
          },
          options: {
            responsive: true,
            plugins: { legend: { display: false } },
            scales: {
              x: { grid: { display: false }, ticks: { color: "#90909b" } },
              y: { grid: { color: "#1e1e22" }, ticks: { color: "#90909b", stepSize: 1 } },
            },
          },
        });

        const filledHourlyTimeline = fillHourlyTimeline(hourlyTimeline);
        const allFilledHours = Object.keys(filledHourlyTimeline);
        const chHourlyTimelineSum = Object.values(filledHourlyTimeline).reduce((s, v) => s + v, 0);
        let chRunningTotal = (data.total_activations || 0) - chHourlyTimelineSum;

        if (channelHourlyCumulativeChartInstance) channelHourlyCumulativeChartInstance.destroy();
        channelHourlyCumulativeChartInstance = new Chart(document.getElementById("channelHourlyCumulativeChart").getContext("2d"), {
          type: "line",
          data: {
            labels: allFilledHours,
            datasets: [{
              label: "Total Historical Berts",
              data: allFilledHours.map(h => { chRunningTotal += filledHourlyTimeline[h]; return chRunningTotal; }),
              borderColor: "#3b82f6",
              backgroundColor: "rgba(59, 130, 246, 0.08)",
              borderWidth: 2,
              pointRadius: allFilledHours.length > 50 ? 0 : 2,
              tension: 0.1,
              fill: true,
            }],
          },
          options: {
            responsive: true,
            plugins: { legend: { display: false } },
            scales: {
              x: { grid: { color: "#161618" }, ticks: { color: "#90909b", maxTicksLimit: 12 } },
              y: { grid: { color: "#1e1e22" }, ticks: { color: "#90909b", stepSize: 1 } },
            },
          },
        });

        document.getElementById("chFirstDiscover").innerHTML = (() => {
          const fpb = data.first_person_per_bert;
          if (!fpb) return "<i>None</i>";
          let entries = Object.entries(fpb).map(([bert, info]) => ({ bert, user: info.user, timestamp: info.timestamp }));
          if (entries.length === 0) return "<i>None</i>";
        entries.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
          return '<div class="first-list">' +
            entries.map(e =>
              `<div class="first-entry">
                <span class="first-bert">${e.bert}</span>
                <span class="first-user link-name" onclick="navigateTo('user', '${e.user}')">${e.user}</span>
                <span class="first-date">${formatDate(e.timestamp)}</span>
              </div>`
            ).join("") + "</div>";
        })();
      }

      // ── USER ──
      let currentUserData = null;

      async function loadUser() {
        const name = document.getElementById("uInput").value;
        if (!name) return;
        document.getElementById("uAutocomplete").classList.remove("open");
        const res = await fetch(`/api/user?name=${name}`);
        if (!res.ok) { alert("User not found!"); return; }
        currentUserData = await res.json();

        document.getElementById("uResults").style.display = "block";
        document.getElementById("uTitle").innerText = `Player: ${currentUserData.username}`;
        document.getElementById("uTotal").innerText = currentUserData.global_total;
        document.getElementById("uZazas").innerHTML = zazaEmoteHTML() + " <strong>" + (currentUserData.global_zazas || 0) + "</strong>";
        document.getElementById("uZazaLs").innerHTML = zazaLEmoteHTML() + " <strong>" + (currentUserData.global_zaza_ls || 0) + "</strong>";
        document.getElementById("uDoubleZazas").innerHTML = zazaDoubleEmoteHTML() + " <strong>" + (currentUserData.global_double_zazas || 0) + "</strong>";

        const channels = Object.entries(currentUserData.channels).map(([ch, stats]) => {
          return { name: ch, total: stats.total || 0 };
        });
        channels.sort((a, b) => b.total - a.total);
        window.userChannelListCache = channels;

        if (channels.length > 0) {
          document.getElementById("uChannelInput").value = channels[0].name;
          loadUserChannelSelection();
        }
      }

      async function updateUserChannelView() {
        if (!currentUserData) return;
        const chan = document.getElementById("uChannelInput").value.toLowerCase().trim();
        const channelStats = currentUserData.channels[chan];
        if (!channelStats) return;

        let emotesMap = {};
        try {
          const emoteRes = await fetch(`/api/emotes?channel=${chan}`);
          if (emoteRes.ok) {
            (await emoteRes.json()).forEach(em => {
              emotesMap[em.name] = em.id;
              window.globalEmoteMemoryCache[em.name] = em.id;
            });
          }
        } catch (err) {}

        document.getElementById("uCollected").innerHTML =
          Object.entries(channelStats.collected || {})
            .map(([bert, info]) => {
              const count = getCountValue(info);
              const zazaCount = info.zaza_count || 0;
              const zazaLCount = info.zaza_l_count || 0;
              const doubleCount = info.double_zaza_count || 0;
              const dateStr = info.first_seen ? formatDate(info.first_seen) : "Unknown";
              const zazaParts = [];
              if (zazaCount > 0) zazaParts.push(`zaza x${zazaCount}`);
              if (zazaLCount > 0) zazaParts.push(`zazaL x${zazaLCount}`);
              if (doubleCount > 0) zazaParts.push(`double x${doubleCount}`);
              const zazaStr = zazaParts.length > 0 ? `, ${zazaParts.join(", ")}` : "";
              const parts = COMPOSITE_BERTS[bert];
              if (parts) {
                const ids = parts.map(p => emotesMap[p]).filter(Boolean);
                if (ids.length === parts.length) {
                  return `<div class="emote-wrapper" style="display:inline-flex;gap:0;">${ids.map(id => `<img src="https://cdn.7tv.app/emote/${id}/2x.webp" alt="${bert}" class="twitch-emote" style="height:32px;" />`).join("")}<small style="color: var(--text-muted); font-weight: 600; font-family: 'DM Mono', monospace; margin-left: 2px;">x${count}</small><span class="emote-tooltip">${bert} (x${count}${zazaStr}, first: ${dateStr})</span></div>`;
                }
              }
              const emId = emotesMap[bert];
              return emId
                ? `<div class="emote-wrapper"><img src="https://cdn.7tv.app/emote/${emId}/2x.webp" alt="${bert}" class="twitch-emote" /><small style="color: var(--text-muted); font-weight: 600; font-family: 'DM Mono', monospace; margin-left: 2px;">x${count}</small><span class="emote-tooltip">${bert} (x${count}${zazaStr}, first: ${dateStr})</span></div>`
                : `<span class="badge" title="x${count}${zazaStr}, first: ${dateStr}">${bert} <small>x${count}</small></span>`;
            }).join(" ") || "<i>None yet!</i>";

        document.getElementById("uMissing").innerHTML =
          (channelStats.missing || [])
            .map(bert => {
              return renderBertEmote(bert, emotesMap) || `<span class="badge missing">${bert}</span>`;
            }).join(" ") || "<i>Got 'em all!</i>";

        const userTimeline = fillDailyTimeline(channelStats.daily_timeline || {});
        const sortedDates = Object.keys(userTimeline);

        if (userChartInstance) userChartInstance.destroy();
        userChartInstance = new Chart(document.getElementById("userChart").getContext("2d"), {
          type: "line",
          data: {
            labels: sortedDates,
            datasets: [{
              label: "Daily Interactions",
              data: sortedDates.map(d => userTimeline[d]),
              borderColor: "#9146ff",
              backgroundColor: "rgba(145, 70, 255, 0.1)",
              borderWidth: 2,
              tension: 0.3,
              fill: true,
            }],
          },
          options: {
            responsive: true,
            plugins: { legend: { display: false } },
            scales: {
              x: { grid: { color: "#1e1e22" }, ticks: { color: "#90909b" } },
              y: { grid: { color: "#1e1e22" }, ticks: { color: "#90909b", stepSize: 1 } },
            },
          },
        });

        const userHourlyTimeline = channelStats.hourly_timeline || {};
        const userLast24Data = getLast24Hours(userHourlyTimeline);
        const userLast24Labels = Object.keys(userLast24Data);

        if (userHourlyChartInstance) userHourlyChartInstance.destroy();
        userHourlyChartInstance = new Chart(document.getElementById("userHourlyChart").getContext("2d"), {
          type: "bar",
          data: {
            labels: userLast24Labels.map(h => h.split(" ")[1] || h),
            datasets: [{
              label: "User Hourly Counts",
              data: userLast24Labels.map(h => userLast24Data[h]),
              backgroundColor: "rgba(236, 72, 153, 0.6)",
              borderColor: "#ec4899",
              borderWidth: 1,
              borderRadius: 3,
            }],
          },
          options: {
            responsive: true,
            plugins: { legend: { display: false } },
            scales: {
              x: { grid: { display: false }, ticks: { color: "#90909b" } },
              y: { grid: { color: "#1e1e22" }, ticks: { color: "#90909b", stepSize: 1 } },
            },
          },
        });

        const filledUserHourlyTimeline = fillHourlyTimeline(userHourlyTimeline);
        const allFilledUserHours = Object.keys(filledUserHourlyTimeline);
        const hourlyTimelineSum = Object.values(filledUserHourlyTimeline).reduce((s, v) => s + v, 0);
        let uRunningTotal = (channelStats.total || 0) - hourlyTimelineSum;

        if (userHourlyCumulativeChartInstance) userHourlyCumulativeChartInstance.destroy();
        userHourlyCumulativeChartInstance = new Chart(document.getElementById("userHourlyCumulativeChart").getContext("2d"), {
          type: "line",
          data: {
            labels: allFilledUserHours,
            datasets: [{
              label: "Total Collection Progress",
              data: allFilledUserHours.map(h => { uRunningTotal += filledUserHourlyTimeline[h]; return uRunningTotal; }),
              borderColor: "#ec4899",
              backgroundColor: "rgba(236, 72, 153, 0.08)",
              borderWidth: 2,
              pointRadius: allFilledUserHours.length > 50 ? 0 : 2,
              tension: 0.1,
              fill: true,
            }],
          },
          options: {
            responsive: true,
            plugins: { legend: { display: false } },
            scales: {
              x: { grid: { color: "#161618" }, ticks: { color: "#90909b", maxTicksLimit: 12 } },
              y: { grid: { color: "#1e1e22" }, ticks: { color: "#90909b" } },
            },
          },
        });
      }

      // ── GOLDENZAZA ──
      let gzHallData = null;

      async function loadGoldenzaza() {
        if (gzHallData) return;
        const res = await fetch("/api/global");
        if (!res.ok) return;
        const data = await res.json();
        gzHallData = data.goldenzazabert_hall || [];

        document.getElementById("gzEmpty").style.display = "none";
        document.getElementById("gzResults").style.display = "block";
        document.getElementById("gzGlobalHall").innerHTML = renderGoldenzazaHall(gzHallData);
      }

      // ── Fun mode: bert rain (DOM-based for GIF support) ──
      let funModeActive = false;
      let funAnimId = null;
      let funBertEmotes = []; // [{name, src}]
      let funDrops = [];
      const FUN_H = 64;      // emote display height
      const FUN_SPACING = 100;
      const FUN_SPEED_MIN = 0.5;
      const FUN_SPEED_MAX = 1.5;
      const FUN_ALPHA = 0.18;

      function pickRandomBert() {
        if (funBertEmotes.length === 0) return null;
        return funBertEmotes[Math.floor(Math.random() * funBertEmotes.length)];
      }

      function createDropEl(x, startY) {
        const bert = pickRandomBert();
        const img = document.createElement("img");
        img.src = bert ? bert.src : "";
        img.style.cssText = `position:absolute;height:${FUN_H}px;opacity:${FUN_ALPHA};will-change:transform;`;
        document.getElementById("bertBgContainer").appendChild(img);
        return {
          el: img,
          x: x,
          y: startY !== undefined ? startY : -FUN_H - Math.random() * window.innerHeight,
          speed: FUN_SPEED_MIN + Math.random() * (FUN_SPEED_MAX - FUN_SPEED_MIN),
        };
      }

      function recycleDrop(drop) {
        const bert = pickRandomBert();
        if (bert) drop.el.src = bert.src;
        drop.y = -FUN_H - Math.random() * 200;
        drop.speed = FUN_SPEED_MIN + Math.random() * (FUN_SPEED_MAX - FUN_SPEED_MIN);
      }

      async function toggleFunMode() {
        const btn = document.getElementById("funToggle");
        const container = document.getElementById("bertBgContainer");

        if (funModeActive) {
          funModeActive = false;
          btn.classList.remove("active");
          container.style.opacity = "0";
          if (funAnimId) cancelAnimationFrame(funAnimId);
          funAnimId = null;
          // Clean up drops after fade out
          setTimeout(() => {
            for (const d of funDrops) d.el.remove();
            funDrops = [];
          }, 400);
          return;
        }

        funModeActive = true;
        btn.classList.add("active");

        // Load bert emotes if not cached
        if (funBertEmotes.length === 0) {
          const channels = (window.channelListCache || []).map(c => c.name);
          const seen = new Set();
          for (const ch of channels.slice(0, 10)) {
            try {
              const [emoteRes, channelRes] = await Promise.all([
                fetch(`/api/emotes?channel=${ch}`),
                fetch(`/api/channel?name=${ch}`),
              ]);
              if (emoteRes.ok && channelRes.ok) {
                const emotes = await emoteRes.json();
                const channelData = await channelRes.json();
                const activeBerts = new Set(channelData.active_berts || []);
                const emoteMap = {};
                for (const em of emotes) emoteMap[em.name] = em.id;
                for (const bertName of activeBerts) {
                  const emId = emoteMap[bertName];
                  if (emId && !seen.has(emId)) {
                    seen.add(emId);
                    funBertEmotes.push({ name: bertName, src: `https://cdn.7tv.app/emote/${emId}/2x.webp` });
                  }
                }
              }
            } catch (e) {}
          }
          if (funBertEmotes.length === 0) {
            funBertEmotes.push({ name: "fallback", src: "https://cdn.7tv.app/emote/01K6B3Y3SG5ZBAWXXVKX7BKBE5/4x.webp" });
          }
        }

        // Create rain drops
        funDrops = [];
        const cols = Math.ceil(window.innerWidth / FUN_SPACING) + 2;
        for (let i = 0; i < cols; i++) {
          funDrops.push(createDropEl(i * FUN_SPACING + Math.random() * 30 - 15));
        }

        // Let browser render images, then fade in
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            container.style.opacity = "1";
            funAnimate();
          });
        });
      }

      function funAnimate() {
        if (!funModeActive) return;
        const maxH = window.innerHeight;

        for (const drop of funDrops) {
          drop.y += drop.speed;
          if (drop.y > maxH + FUN_H) recycleDrop(drop);
          drop.el.style.transform = `translate3d(${drop.x}px,${drop.y}px,0)`;
        }

        funAnimId = requestAnimationFrame(funAnimate);
      }

      function onNewChannelData() {}

      // Bootstrap
      loadGlobal();
      connectLiveFeed();
    