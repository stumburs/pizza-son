
      const DOCS_URL = "/api/docs";

      const state = {
        data: null,
        commandQuery: "",
        category: "all",
        listenerQuery: "",
      };

      const commandCountEl = document.getElementById("commandCount");
      const listenerCountEl = document.getElementById("listenerCount");
      const commandsPanel = document.getElementById("commandsPanel");
      const listenersPanel = document.getElementById("listenersPanel");
      const commandGroupsEl = document.getElementById("commandGroups");
      const listenerListEl = document.getElementById("listenerList");
      const commandStatusEl = document.getElementById("commandStatus");
      const listenerStatusEl = document.getElementById("listenerStatus");
      const commandSearchEl = document.getElementById("commandSearch");
      const listenerSearchEl = document.getElementById("listenerSearch");
      const categoryEl = document.getElementById("categoryFilter");
      const permissionRowsEl = document.getElementById("permissionRows");

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

      function setStatus(el, message, isError) {
        el.textContent = message;
        el.classList.toggle("error", Boolean(isError));
      }

      function commandMatches(command, query) {
        if (!query) return true;
        const examples = (command.examples || [])
          .flatMap((example) => [example.input, example.output])
          .join(" ");
        return [
          command.name,
          command.description,
          command.usage,
          command.permission,
          examples,
        ]
          .join(" ")
          .toLowerCase()
          .includes(query);
      }

      function renderCommands() {
        const data = state.data;
        if (!data) return;

        const query = state.commandQuery.trim().toLowerCase();
        const groups = data.commands
          .filter((group) => state.category === "all" || group.name === state.category)
          .map((group) => ({
            name: group.name,
            commands: group.commands.filter((command) => commandMatches(command, query)),
          }))
          .filter((group) => group.commands.length > 0);

        const visible = groups.reduce((sum, group) => sum + group.commands.length, 0);
        const filtering = Boolean(query) || state.category !== "all";
        setStatus(
          commandStatusEl,
          filtering
            ? `${formatCount(visible)} of ${formatCount(data.total_commands)} commands`
            : `${formatCount(data.total_commands)} commands`,
          false
        );

        if (groups.length === 0) {
          commandGroupsEl.innerHTML = `<div class="empty">No commands match those filters.</div>`;
          return;
        }

        commandGroupsEl.innerHTML = groups
          .map(
            (group) => `
            <section class="group">
              <div class="group-head">
                <h2>${escapeHtml(group.name)}</h2>
                <span class="group-count">${formatCount(group.commands.length)}</span>
              </div>
              ${group.commands.map(renderCommand).join("")}
            </section>`
          )
          .join("");
      }

      function renderCommand(command) {
        const examples = command.examples || [];
        return `
          <article class="entry">
            <div class="entry-head">
              <h3><code>${escapeHtml(state.data.prefix || "!")}${escapeHtml(command.name)}</code></h3>
              <span class="entry-perm">${escapeHtml(command.permission)}</span>
            </div>
            ${command.description ? `<p class="entry-desc">${escapeHtml(command.description)}</p>` : ""}
            ${command.usage ? `<div class="entry-usage"><code>${escapeHtml(command.usage)}</code></div>` : ""}
            ${command.cooldown ? `<div class="entry-meta">Cooldown ${escapeHtml(command.cooldown)}</div>` : ""}
            ${examples.length ? renderExamples(examples) : ""}
          </article>`;
      }

      function renderExamples(examples) {
        return `
          <details class="examples">
            <summary>${formatCount(examples.length)} ${examples.length === 1 ? "example" : "examples"}</summary>
            <div class="example-list">
              ${examples
                .map(
                  (example) => `
                  <div class="example">
                    <code>${escapeHtml(example.input)}</code>
                    ${example.output ? `<span class="arrow">→</span><code>${escapeHtml(example.output)}</code>` : ""}
                  </div>`
                )
                .join("")}
            </div>
          </details>`;
      }

      function listenerMatches(listener, query) {
        if (!query) return true;
        return [listener.name, listener.description, listener.permission]
          .join(" ")
          .toLowerCase()
          .includes(query);
      }

      function renderListeners() {
        const data = state.data;
        if (!data) return;

        const query = state.listenerQuery.trim().toLowerCase();
        const listeners = data.listeners.filter((listener) => listenerMatches(listener, query));

        setStatus(
          listenerStatusEl,
          query
            ? `${formatCount(listeners.length)} of ${formatCount(data.total_listeners)} listeners`
            : `${formatCount(data.total_listeners)} listeners`,
          false
        );

        if (listeners.length === 0) {
          listenerListEl.innerHTML = `<div class="empty">No listeners match that search.</div>`;
          return;
        }

        listenerListEl.innerHTML = listeners
          .map(
            (listener) => `
            <article class="entry">
              <div class="entry-head">
                <h3><code>${escapeHtml(listener.name)}</code></h3>
                <span class="entry-perm">${escapeHtml(listener.permission)}</span>
              </div>
              ${listener.description ? `<p class="entry-desc">${escapeHtml(listener.description)}</p>` : ""}
              <div class="entry-meta">
                ${listener.cooldown ? `Cooldown ${escapeHtml(listener.cooldown)}` : "No cooldown"}
                ${listener.passive ? " · passive" : ""}
              </div>
            </article>`
          )
          .join("");
      }

      function renderPermissions() {
        const levels = state.data?.permission_levels || [];
        permissionRowsEl.innerHTML = levels
          .map(
            (level) => `
            <div class="perm-row">
              <span class="perm-name">${escapeHtml(level.name)}</span>
              <span class="perm-desc">${escapeHtml(level.description)}</span>
            </div>`
          )
          .join("");
      }

      function fillCategories() {
        const groups = state.data?.commands || [];
        categoryEl.innerHTML =
          `<option value="all">All categories</option>` +
          groups
            .map((group) => `<option value="${escapeHtml(group.name)}">${escapeHtml(group.name)}</option>`)
            .join("");
      }

      function selectSection(section) {
        const showCommands = section === "commands";
        commandsPanel.hidden = !showCommands;
        listenersPanel.hidden = showCommands;
        document.querySelectorAll(".section-tab").forEach((tab) => {
          const active = tab.dataset.section === section;
          tab.classList.toggle("active", active);
          tab.setAttribute("aria-selected", String(active));
        });
      }

      async function load() {
        try {
          const response = await fetch(DOCS_URL, { cache: "no-store" });
          if (!response.ok) throw new Error(`HTTP ${response.status}`);

          state.data = await response.json();
          commandCountEl.textContent = formatCount(state.data.total_commands);
          listenerCountEl.textContent = formatCount(state.data.total_listeners);

          fillCategories();
          renderPermissions();
          renderCommands();
          renderListeners();
        } catch (error) {
          setStatus(commandStatusEl, "Could not load docs.", true);
          setStatus(listenerStatusEl, "Could not load docs.", true);
          commandGroupsEl.innerHTML = `<div class="empty">Commands could not be loaded.</div>`;
          listenerListEl.innerHTML = `<div class="empty">Listeners could not be loaded.</div>`;
        }
      }

      document.querySelectorAll(".section-tab").forEach((tab) => {
        tab.addEventListener("click", () => selectSection(tab.dataset.section));
      });

      commandSearchEl.addEventListener("input", () => {
        state.commandQuery = commandSearchEl.value;
        renderCommands();
      });

      categoryEl.addEventListener("change", () => {
        state.category = categoryEl.value;
        renderCommands();
      });

      listenerSearchEl.addEventListener("input", () => {
        state.listenerQuery = listenerSearchEl.value;
        renderListeners();
      });

      load();
    