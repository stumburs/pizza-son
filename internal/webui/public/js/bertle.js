
      const MAX_GUESSES = 6;
      const MIN_WORD_LENGTH = 4;
      const MAX_WORD_LENGTH = 10;
      const STORAGE_PREFIX = "bertle:";
      const ENDLESS_PREFIX = "bertle:endless:";
      const MODE_KEY = "bertle:mode";
      const STREAK_PREFIX = "bertle:endless:streak:";
      const COMPOSITE_BERTS = {
        bigbert: ["bigbert1", "bigbert2"],
      };

      const channelSelect = document.getElementById("channelSelect");
      const channelLabel = document.getElementById("channelLabel");
      const gameStatus = document.getElementById("gameStatus");
      const messageEl = document.getElementById("message");
      const board = document.getElementById("board");
      const emptyState = document.getElementById("emptyState");
      const resultEl = document.getElementById("result");
      const resultEmote = document.getElementById("resultEmote");
      const resultTitle = document.getElementById("resultTitle");
      const resultName = document.getElementById("resultName");
      const resultChannel = document.getElementById("resultChannel");
      const giveUpBtn = document.getElementById("giveUpBtn");
      const nextBtn = document.getElementById("nextBtn");
      const shareBtn = document.getElementById("shareBtn");
      const keyboardEl = document.getElementById("keyboard");
      const modeButtons = [...document.querySelectorAll(".mode-btn")];

      let mode = "daily";
      let channel = "";
      let candidates = [];
      let emoteMap = {};
      let state = null;
      let typed = "";
      let animateRow = -1;
      let loadToken = 0;
      let messageTimer = null;

      function dateKey(date = new Date()) {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, "0");
        const day = String(date.getDate()).padStart(2, "0");
        return `${year}-${month}-${day}`;
      }

      function formatDate(key) {
        const parts = String(key || "").split("-").map(Number);
        if (parts.length !== 3 || !parts[0]) return key || "";
        return new Date(parts[0], parts[1] - 1, parts[2]).toLocaleDateString(undefined, {
          month: "short",
          day: "numeric",
          year: "numeric",
        });
      }

      // FNV-1a: keeps the daily word stable for a channel without a backend.
      function hashSeed(text) {
        let hash = 2166136261;
        for (let i = 0; i < text.length; i++) {
          hash ^= text.charCodeAt(i);
          hash = Math.imul(hash, 16777619);
        }
        return hash >>> 0;
      }

      function escapeHTML(value) {
        return String(value ?? "").replace(/[&<>"']/g, (char) => ({
          "&": "&amp;",
          "<": "&lt;",
          ">": "&gt;",
          '"': "&quot;",
          "'": "&#39;",
        })[char]);
      }

      function validTarget(target) {
        return (
          typeof target === "string" &&
          /^[a-z]+$/.test(target) &&
          target.length >= MIN_WORD_LENGTH &&
          target.length <= MAX_WORD_LENGTH
        );
      }

      function cleanGuesses(guesses, target) {
        if (!Array.isArray(guesses)) return [];
        return guesses
          .filter(
            (guess) =>
              typeof guess === "string" &&
              guess.length === target.length &&
              /^[a-z]+$/.test(guess)
          )
          .slice(0, MAX_GUESSES);
      }

      function loadDailyState(channelName, words) {
        try {
          const raw = localStorage.getItem(STORAGE_PREFIX + channelName);
          if (raw) {
            const saved = JSON.parse(raw);
            if (
              saved &&
              saved.date === dateKey() &&
              validTarget(saved.target) &&
              Array.isArray(saved.guesses)
            ) {
              return {
                date: saved.date,
                target: saved.target,
                guesses: cleanGuesses(saved.guesses, saved.target),
                gaveUp: Boolean(saved.gaveUp),
              };
            }
          }
        } catch (error) {
          // Storage can be unavailable in private modes; fall through.
        }

        const target = words[hashSeed(`${channelName}:${dateKey()}`) % words.length];
        return { date: dateKey(), target, guesses: [], gaveUp: false };
      }

      function randomTarget(words, avoid) {
        const pool = words.filter((word) => word !== avoid);
        const options = pool.length > 0 ? pool : words;
        return options[Math.floor(Math.random() * options.length)];
      }

      function newEndlessState(words, avoid) {
        return {
          date: dateKey(),
          target: randomTarget(words, avoid),
          guesses: [],
          gaveUp: false,
        };
      }

      function loadEndlessState(channelName, words) {
        try {
          const raw = localStorage.getItem(ENDLESS_PREFIX + channelName);
          if (raw) {
            const saved = JSON.parse(raw);
            if (saved && validTarget(saved.target) && Array.isArray(saved.guesses)) {
              return {
                date: saved.date || dateKey(),
                target: saved.target,
                guesses: cleanGuesses(saved.guesses, saved.target),
                gaveUp: Boolean(saved.gaveUp),
              };
            }
          }
        } catch (error) {
          // Storage can be unavailable in private modes; fall through.
        }

        return newEndlessState(words);
      }

      function storageKey() {
        return (mode === "endless" ? ENDLESS_PREFIX : STORAGE_PREFIX) + channel;
      }

      function saveState() {
        if (!state || !channel) return;
        try {
          localStorage.setItem(storageKey(), JSON.stringify(state));
        } catch (error) {
          // The game still works without persistence.
        }
      }

      function getStreak() {
        if (!channel) return 0;
        try {
          return Number(localStorage.getItem(STREAK_PREFIX + channel)) || 0;
        } catch (error) {
          return 0;
        }
      }

      function setStreak(value) {
        if (!channel) return;
        try {
          localStorage.setItem(STREAK_PREFIX + channel, String(Math.max(0, value)));
        } catch (error) {
          // Streak tracking is optional.
        }
      }

      function updateEndlessStreak(won) {
        if (mode !== "endless") return;
        setStreak(won ? getStreak() + 1 : 0);
      }

      function hasWon() {
        return Boolean(state && state.guesses.length > 0 && state.guesses[state.guesses.length - 1] === state.target);
      }

      function isGameOver() {
        if (!state) return true;
        return Boolean(state.gaveUp) || hasWon() || state.guesses.length >= MAX_GUESSES;
      }

      function scoreGuess(guess, target) {
        const marks = new Array(target.length).fill("absent");
        const remaining = {};

        for (let i = 0; i < target.length; i++) {
          if (guess[i] === target[i]) {
            marks[i] = "correct";
          } else {
            remaining[target[i]] = (remaining[target[i]] || 0) + 1;
          }
        }

        for (let i = 0; i < target.length; i++) {
          if (marks[i] === "correct") continue;
          const letter = guess[i];
          if (remaining[letter] > 0) {
            marks[i] = "present";
            remaining[letter]--;
          }
        }

        return marks;
      }

      function keyStates() {
        const rank = { absent: 1, present: 2, correct: 3 };
        const states = {};
        if (!state) return states;
        for (const guess of state.guesses) {
          const marks = scoreGuess(guess, state.target);
          for (let i = 0; i < guess.length; i++) {
            const letter = guess[i];
            const mark = marks[i];
            if (!states[letter] || rank[mark] > rank[states[letter]]) {
              states[letter] = mark;
            }
          }
        }
        return states;
      }

      function setMessage(text, kind) {
        clearTimeout(messageTimer);
        messageEl.textContent = text || "";
        messageEl.className = "message" + (kind ? ` ${kind}` : "");
        if (text) {
          messageTimer = setTimeout(() => {
            messageEl.textContent = "";
            messageEl.className = "message";
          }, 2400);
        }
      }

      function fitBoard() {
        if (!state) return;
        const gameEl = document.querySelector(".game");
        const styles = getComputedStyle(gameEl);
        const innerWidth = gameEl.clientWidth - parseFloat(styles.paddingLeft) - parseFloat(styles.paddingRight);
        const available = Math.min(innerWidth || window.innerWidth - 48, 520);
        const cols = state.target.length;
        const gap = 6;
        const tile = Math.max(20, Math.min(56, Math.floor((available - gap * (cols - 1)) / cols)));
        board.style.setProperty("--tile", `${tile}px`);
      }

      function renderBoard() {
        board.innerHTML = "";
        if (!state) return;

        const targetLength = state.target.length;
        const animate = animateRow;
        animateRow = -1;

        board.style.setProperty("--cols", targetLength);

        for (let rowIndex = 0; rowIndex < MAX_GUESSES; rowIndex++) {
          const row = document.createElement("div");
          row.className = "board-row";

          const guess = state.guesses[rowIndex];
          const isActiveRow = !guess && rowIndex === state.guesses.length && !isGameOver();
          const marks = guess ? scoreGuess(guess, state.target) : null;

          for (let col = 0; col < targetLength; col++) {
            const tile = document.createElement("div");
            tile.className = "tile";
            tile.style.setProperty("--i", col);

            const letter = guess ? guess[col] : isActiveRow ? typed[col] : "";
            if (letter) {
              tile.textContent = letter;
              tile.classList.add("filled");
            }

            if (marks) {
              tile.classList.add(marks[col]);
              if (rowIndex === animate) tile.classList.add("reveal");
            }

            row.appendChild(tile);
          }

          board.appendChild(row);
        }
      }

      function renderKeyboard() {
        const states = keyStates();
        const over = isGameOver();

        keyboardEl.querySelectorAll(".key[data-key]").forEach((key) => {
          const letter = key.dataset.key;
          key.classList.remove("correct", "present", "absent");
          if (states[letter]) key.classList.add(states[letter]);
          key.setAttribute("aria-disabled", over ? "true" : "false");
        });

        keyboardEl.classList.toggle("done", over);
      }

      function renderStatus() {
        if (!state) return;

        const streak = mode === "endless" ? ` · streak ${getStreak()}` : "";
        const prefix = mode === "endless" ? "Endless · " : `${formatDate(state.date)} · `;

        if (hasWon()) {
          gameStatus.textContent = `Solved in ${state.guesses.length}/${MAX_GUESSES}${streak}`;
        } else if (state.gaveUp) {
          gameStatus.textContent = `Gave up after ${state.guesses.length}/${MAX_GUESSES}${streak}`;
        } else if (state.guesses.length >= MAX_GUESSES) {
          gameStatus.textContent = `Out of guesses${streak}`;
        } else {
          gameStatus.textContent = `${prefix}${state.target.length} letters · ${state.guesses.length}/${MAX_GUESSES} guesses${streak}`;
        }
      }

      function renderResult() {
        if (!state || !isGameOver()) {
          resultEl.hidden = true;
          return;
        }

        resultEl.hidden = false;

        const compositeParts = COMPOSITE_BERTS[state.target];
        const compositeIds = compositeParts
          ? compositeParts.map((part) => emoteMap[part]).filter(Boolean)
          : [];
        const emoteId = emoteMap[state.target];

        if (compositeParts && compositeIds.length === compositeParts.length) {
          resultEmote.innerHTML = `<span class="composite">${compositeIds
            .map(
              (id) =>
                `<img src="https://cdn.7tv.app/emote/${encodeURIComponent(id)}/2x.webp" alt="${escapeHTML(state.target)}" />`
            )
            .join("")}</span>`;
        } else if (emoteId) {
          resultEmote.innerHTML = `<img src="https://cdn.7tv.app/emote/${encodeURIComponent(emoteId)}/2x.webp" alt="${escapeHTML(state.target)}" />`;
        } else {
          resultEmote.innerHTML = `<span class="result-letter">${escapeHTML(state.target[0] || "?")}</span>`;
        }

        resultTitle.textContent = hasWon()
          ? `Solved in ${state.guesses.length}/${MAX_GUESSES}`
          : "The bert was…";
        resultName.textContent = state.target;
        resultChannel.textContent =
          mode === "endless" ? `#${channel} · streak ${getStreak()}` : `#${channel}`;
      }

      function render() {
        if (!state) return;
        renderBoard();
        renderKeyboard();
        renderStatus();
        renderResult();
        fitBoard();

        giveUpBtn.disabled = isGameOver();
        shareBtn.disabled = !isGameOver();
        nextBtn.hidden = !(mode === "endless" && isGameOver());
      }

      function submitGuess() {
        if (!state || isGameOver()) return;

        if (typed.length !== state.target.length) {
          setMessage(`Guess must be ${state.target.length} letters.`, "error");
          return;
        }

        const rowIndex = state.guesses.length;
        state.guesses.push(typed);
        typed = "";
        animateRow = rowIndex;

        if (mode === "endless" && (hasWon() || state.guesses.length >= MAX_GUESSES)) {
          updateEndlessStreak(hasWon());
        }

        saveState();
        render();

        if (hasWon()) {
          setMessage("You got it!", "success");
        } else if (state.guesses.length >= MAX_GUESSES) {
          setMessage("Out of guesses.", "error");
        }
      }

      function handleKey(key) {
        if (!state || isGameOver()) return;

        if (key === "enter") {
          submitGuess();
          return;
        }

        if (key === "backspace") {
          typed = typed.slice(0, -1);
          renderBoard();
          return;
        }

        if (/^[a-z]$/.test(key) && typed.length < state.target.length) {
          typed += key;
          clearTimeout(messageTimer);
          messageEl.textContent = "";
          messageEl.className = "message";
          renderBoard();
        }
      }

      function shareText() {
        const score = hasWon() ? `${state.guesses.length}/${MAX_GUESSES}` : `X/${MAX_GUESSES}`;
        const label =
          mode === "endless"
            ? `Bertle endless · #${channel} · streak ${getStreak()}`
            : `Bertle ${state.date} · #${channel}`;
        const lines = [`${label} · ${score}`];

        for (const guess of state.guesses) {
          lines.push(
            scoreGuess(guess, state.target)
              .map((mark) => (mark === "correct" ? "🟩" : mark === "present" ? "🟨" : "⬛"))
              .join("")
          );
        }

        return lines.join("\n");
      }

      async function copyText(text) {
        try {
          await navigator.clipboard.writeText(text);
          return true;
        } catch (error) {
          const area = document.createElement("textarea");
          area.value = text;
          area.setAttribute("readonly", "");
          area.style.position = "fixed";
          area.style.opacity = "0";
          document.body.appendChild(area);
          area.select();
          let copied = false;
          try {
            copied = document.execCommand("copy");
          } catch (copyError) {
            copied = false;
          }
          area.remove();
          return copied;
        }
      }

      function buildKeyboard() {
        const rows = [
          "qwertyuiop".split(""),
          "asdfghjkl".split(""),
          ["enter", ..."zxcvbnm".split(""), "backspace"],
        ];

        keyboardEl.innerHTML = rows
          .map((row) => {
            const keys = row
              .map((key) => {
                if (key === "enter") {
                  return `<button type="button" class="key wide" data-key="enter">enter</button>`;
                }
                if (key === "backspace") {
                  return `<button type="button" class="key wide" data-key="backspace" aria-label="Backspace">⌫</button>`;
                }
                return `<button type="button" class="key" data-key="${key}">${key}</button>`;
              })
              .join("");
            return `<div class="key-row">${keys}</div>`;
          })
          .join("");

        keyboardEl.addEventListener("click", (event) => {
          const key = event.target.closest(".key");
          if (!key) return;
          handleKey(key.dataset.key);
        });
      }

      function showEmpty(text) {
        state = null;
        candidates = [];
        typed = "";
        board.hidden = true;
        keyboardEl.hidden = true;
        emptyState.hidden = false;
        emptyState.textContent = text;
        resultEl.hidden = true;
        giveUpBtn.disabled = true;
        nextBtn.hidden = true;
        shareBtn.disabled = true;
        setModeButtonsDisabled(true);
      }

      function setModeButtonsDisabled(disabled) {
        modeButtons.forEach((button) => {
          button.disabled = disabled;
        });
      }

      function updateModeButtons() {
        modeButtons.forEach((button) => {
          const active = button.dataset.mode === mode;
          button.classList.toggle("active", active);
          button.setAttribute("aria-pressed", active ? "true" : "false");
        });
      }

      function switchMode(newMode) {
        if (mode === newMode) return;
        mode = newMode;
        try {
          localStorage.setItem(MODE_KEY, mode);
        } catch (error) {
          // Mode persistence is optional.
        }
        updateModeButtons();
        typed = "";
        animateRow = -1;
        setMessage("");

        if (!channel || candidates.length === 0) {
          state = null;
          return;
        }

        state =
          mode === "daily"
            ? loadDailyState(channel, candidates)
            : loadEndlessState(channel, candidates);
        render();
      }

      async function loadChannel(name) {
        const token = ++loadToken;

        channel = name;
        emoteMap = {};
        state = null;
        typed = "";
        animateRow = -1;
        board.innerHTML = "";
        board.hidden = false;
        keyboardEl.hidden = false;
        emptyState.hidden = true;
        resultEl.hidden = true;
        giveUpBtn.disabled = true;
        nextBtn.hidden = true;
        shareBtn.disabled = true;
        channelSelect.disabled = true;
        setModeButtonsDisabled(true);
        channelLabel.textContent = `#${name}`;
        gameStatus.textContent = "Loading berts…";
        setMessage("");

        try {
          localStorage.setItem(STORAGE_PREFIX + "channel", name);
        } catch (error) {
          // Ignore storage errors.
        }

        try {
          const [channelResponse, emoteResponse] = await Promise.all([
            fetch(`/api/channel?name=${encodeURIComponent(name)}`),
            fetch(`/api/emotes?channel=${encodeURIComponent(name)}`).catch(() => null),
          ]);

          if (token !== loadToken) return;
          if (!channelResponse.ok) throw new Error(`HTTP ${channelResponse.status}`);

          const data = await channelResponse.json();
          const emotes = emoteResponse && emoteResponse.ok ? await emoteResponse.json() : [];
          if (token !== loadToken) return;

          emoteMap = {};
          (Array.isArray(emotes) ? emotes : []).forEach((emote) => {
            if (emote && emote.name && emote.id) {
              emoteMap[String(emote.name).toLowerCase()] = emote.id;
            }
          });

          candidates = [
            ...new Set(
              (Array.isArray(data.active_berts) ? data.active_berts : [])
                .map((bert) => String(bert).toLowerCase())
                .filter(
                  (bert) =>
                    /^[a-z]+$/.test(bert) &&
                    bert.length >= MIN_WORD_LENGTH &&
                    bert.length <= MAX_WORD_LENGTH
                )
            ),
          ].sort();

          if (candidates.length === 0) {
            channelSelect.disabled = false;
            gameStatus.textContent = `No playable berts in #${name}`;
            showEmpty(
              `#${name} has no bert names with ${MIN_WORD_LENGTH}–${MAX_WORD_LENGTH} letters (a–z). Try another channel.`
            );
            return;
          }

          state =
            mode === "daily"
              ? loadDailyState(name, candidates)
              : loadEndlessState(name, candidates);
          channelSelect.disabled = false;
          setModeButtonsDisabled(false);
          render();
        } catch (error) {
          if (token !== loadToken) return;
          channelSelect.disabled = false;
          gameStatus.textContent = `Could not load #${name}`;
          showEmpty("Failed to load berts for this channel. Try again in a moment.");
        }
      }

      async function bootstrap() {
        buildKeyboard();

        try {
          if (localStorage.getItem(MODE_KEY) === "endless") mode = "endless";
        } catch (error) {
          // Default to daily mode.
        }
        updateModeButtons();

        try {
          const response = await fetch("/api/global");
          if (!response.ok) throw new Error(`HTTP ${response.status}`);
          const data = await response.json();

          const totals = data && data.channel_totals && typeof data.channel_totals === "object"
            ? data.channel_totals
            : {};
          const counts = new Map();

          Object.entries(totals).forEach(([name, count]) => {
            const key = String(name).toLowerCase();
            if (!key || /^\d{17,20}$/.test(key)) return;
            counts.set(key, (counts.get(key) || 0) + (Number(count) || 0));
          });

          const names = [...counts.keys()].sort(
            (a, b) => (counts.get(b) || 0) - (counts.get(a) || 0)
          );

          if (names.length === 0) {
            channelSelect.innerHTML = `<option>No channels</option>`;
            gameStatus.textContent = "No channels available";
            showEmpty("No channels with berts have been recorded yet.");
            return;
          }

          channelSelect.innerHTML = names
            .map((name) => `<option value="${escapeHTML(name)}">#${escapeHTML(name)}</option>`)
            .join("");
          channelSelect.disabled = false;

          let remembered = "";
          try {
            remembered = localStorage.getItem(STORAGE_PREFIX + "channel") || "";
          } catch (error) {
            remembered = "";
          }

          const initial = names.includes(remembered) ? remembered : names[0];
          channelSelect.value = initial;
          channelSelect.addEventListener("change", () => loadChannel(channelSelect.value));

          await loadChannel(initial);
        } catch (error) {
          channelSelect.disabled = true;
          gameStatus.textContent = "Could not load channels";
          showEmpty("The site API is unavailable right now. Try again in a moment.");
        }
      }

      window.addEventListener("keydown", (event) => {
        if (event.ctrlKey || event.metaKey || event.altKey) return;
        if (event.target instanceof Element && event.target.closest("input, select, textarea")) return;

        if (event.key === "Enter") {
          event.preventDefault();
          handleKey("enter");
        } else if (event.key === "Backspace") {
          event.preventDefault();
          handleKey("backspace");
        } else if (/^[a-zA-Z]$/.test(event.key)) {
          event.preventDefault();
          handleKey(event.key.toLowerCase());
        }
      });

      window.addEventListener("resize", fitBoard);

      modeButtons.forEach((button) => {
        button.addEventListener("click", () => switchMode(button.dataset.mode));
      });

      giveUpBtn.addEventListener("click", () => {
        if (!state || isGameOver()) return;
        state.gaveUp = true;
        typed = "";
        updateEndlessStreak(false);
        saveState();
        render();
        setMessage("The bert has been revealed.");
      });

      nextBtn.addEventListener("click", () => {
        if (mode !== "endless" || !state || !isGameOver() || candidates.length === 0) return;
        state = newEndlessState(candidates, state.target);
        typed = "";
        animateRow = -1;
        saveState();
        render();
        setMessage("New bert. Good luck!", "success");
      });

      shareBtn.addEventListener("click", async () => {
        if (!state || !isGameOver()) return;
        const copied = await copyText(shareText());
        setMessage(
          copied ? "Result copied to clipboard." : "Could not copy the result.",
          copied ? "success" : "error"
        );
      });

      bootstrap();
    