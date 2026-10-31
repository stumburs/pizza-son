
      let sessionToken = null;
      let currentChannel = null;
      let changes = {};

      async function checkSession() {
        try {
          const res = await fetch('/api/auth/me', {
            headers: { 'X-Panel-Session': sessionToken || '' },
          });
          if (!res.ok) return false;
          const data = await res.json();
          sessionToken = data.token || sessionToken;
          showPanel(data);
          return true;
        } catch (e) {
          return false;
        }
      }

      // Handle OAuth redirect (fragment with access_token)
      function handleOAuthRedirect() {
        const hash = window.location.hash.substring(1);
        if (!hash) return;
        const params = new URLSearchParams(hash);
        const token = params.get('access_token');
        if (!token) return;
        window.location.hash = '';
        sendToken(token);
      }

      async function sendToken(token) {
        setLoggingIn(true);
        document.getElementById('loginDesc').textContent = 'Completing login...';
        try {
          const res = await fetch('/api/auth/callback', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ token }),
          });
          if (!res.ok) {
            const text = await res.text();
            document.getElementById('loginScreen').innerHTML = `
              <h1>Control panel</h1>
              <p style="color: var(--red);">Login failed: ${text}</p>
              <button class="login-btn" onclick="login()">Try again</button>
            `;
            return;
          }
          const data = await res.json();
          sessionToken = data.token;
          showPanel(data);
        } catch (e) {
          document.getElementById('loginScreen').innerHTML = `
            <h1>Control panel</h1>
            <p style="color: var(--red);">Connection error</p>
            <button class="login-btn" onclick="login()">Try again</button>
          `;
        }
      }

      function setLoggingIn(loading) {
        const btn = document.getElementById('loginBtn');
        const text = document.getElementById('loginBtnText');
        const spinner = document.getElementById('loginLoading');
        btn.disabled = loading;
        text.textContent = loading ? 'Redirecting...' : 'Log in with Twitch';
        spinner.classList.toggle('hidden', !loading);
      }

      function login() {
        setLoggingIn(true);
        fetch('/api/auth/login')
          .then(r => r.json())
          .then(data => {
            const redirect = encodeURIComponent(window.location.origin + '/panel');
            const url = 'https://id.twitch.tv/oauth2/authorize' +
              '?client_id=' + data.client_id +
              '&redirect_uri=' + redirect +
              '&response_type=token' +
              '&scope=user:read:email';
            // Twitch refuses to render its consent screen inside an iframe.
            (window.top || window).location.href = url;
          })
          .catch(() => { setLoggingIn(false); });
      }

      function showPanel(data) {
        document.getElementById('loginScreen').classList.add('hidden');
        document.getElementById('panelView').classList.remove('hidden');
        document.getElementById('userDisplay').innerHTML = 'Logged in as <strong>' + (data.display_name || data.username) + '</strong>';

        const bar = document.getElementById('channelBar');
        bar.innerHTML = data.channels.map(ch =>
          `<button class="channel-btn" onclick="selectChannel('${ch}')">#${ch}</button>`
        ).join('');
        if (data.channels.length > 0) selectChannel(data.channels[0]);
      }

      let pendingChanges = { commands: {}, listeners: {} };
      let bertChannel = null;

      function selectChannel(channel) {
        currentChannel = channel;
        document.querySelectorAll('.channel-btn').forEach(b => {
          b.classList.toggle('active', b.textContent === '#' + channel);
        });
        pendingChanges = { commands: {}, listeners: {} };
        changes = {};
        updateSaveStatus();
        if (bertChannel !== channel) {
          bertChannel = channel;
          resetBertTab();
          if (isPanelTabActive('bertcheck')) loadBertOverview();
        }
        loadCommands(channel);
        loadListeners(channel);
      }

      function isPanelTabActive(name) {
        const view = document.getElementById(name + 'Tab');
        return view !== null && !view.classList.contains('hidden');
      }

      function selectPanelTab(name) {
        ['settings', 'bertcheck'].forEach(tab => {
          const view = document.getElementById(tab + 'Tab');
          const btn = document.getElementById(tab === 'settings' ? 'tabBtnSettings' : 'tabBtnBertcheck');
          if (view) view.classList.toggle('hidden', tab !== name);
          if (btn) {
            btn.classList.toggle('active', tab === name);
            btn.setAttribute('aria-selected', tab === name ? 'true' : 'false');
          }
        });
        if (name === 'bertcheck') loadBertOverview();
      }

      function toggleChanged(type, name, enabled) {
        const key = type + ':' + name;
        changes[key] = enabled;
        const hasChanges = Object.keys(changes).length > 0;
        document.getElementById('saveBtn').disabled = !hasChanges;
        if (!hasChanges) {
          document.getElementById('saveStatus').textContent = '';
        } else {
          const count = Object.keys(changes).length;
          document.getElementById('saveStatus').textContent = count + ' change(s) pending';
          document.getElementById('saveStatus').className = '';
        }
      }

      async function loadCommands(channel) {
        const div = document.getElementById('commandsList');
        div.className = 'loading';
        div.textContent = 'Loading commands...';
        try {
          const res = await fetch('/api/panel/commands?channel=' + channel, {
            headers: { 'X-Panel-Session': sessionToken },
          });
          if (!res.ok) { div.textContent = 'Failed to load commands.'; return; }
          const items = await res.json();
          document.getElementById('cmdCount').textContent = '(' + items.length + ')';
          div.className = '';
          div.innerHTML = items.map(cmd => {
            const enabled = changes['command:' + cmd.name] !== undefined ? changes['command:' + cmd.name] : cmd.enabled;
            return `
              <div class="item">
                <div class="item-info">
                  <span class="item-name">!${cmd.name}</span>
                  <span class="item-desc">${cmd.description}</span>
                  ${!cmd.can_disable ? '<span class="item-protected">protected</span>' : ''}
                </div>
                <label class="toggle">
                  <input type="checkbox" ${enabled ? 'checked' : ''} ${!cmd.can_disable ? 'disabled' : ''}
                    onchange="changeItem('command', '${cmd.name}', this.checked, ${cmd.can_disable})">
                  <span class="slider"></span>
                </label>
              </div>
            `;
          }).join('');
        } catch (e) {
          div.textContent = 'Error loading commands.';
        }
      }

      async function loadListeners(channel) {
        const div = document.getElementById('listenersList');
        div.className = 'loading';
        div.textContent = 'Loading listeners...';
        try {
          const res = await fetch('/api/panel/listeners?channel=' + channel, {
            headers: { 'X-Panel-Session': sessionToken },
          });
          if (!res.ok) { div.textContent = 'Failed to load listeners.'; return; }
          const items = await res.json();
          document.getElementById('listenerCount').textContent = '(' + items.length + ')';
          div.className = '';
          div.innerHTML = items.map(l => {
            const enabled = changes['listener:' + l.name] !== undefined ? changes['listener:' + l.name] : l.enabled;
            return `
              <div class="item">
                <div class="item-info">
                  <span class="item-name">${l.name}</span>
                  <span class="item-desc">${l.description}</span>
                  ${!l.can_disable ? '<span class="item-protected">protected</span>' : ''}
                </div>
                <label class="toggle">
                  <input type="checkbox" ${enabled ? 'checked' : ''} ${!l.can_disable ? 'disabled' : ''}
                    onchange="changeItem('listener', '${l.name}', this.checked, ${l.can_disable})">
                  <span class="slider"></span>
                </label>
              </div>
            `;
          }).join('');
        } catch (e) {
          div.textContent = 'Error loading listeners.';
        }
      }

      function changeItem(type, name, enabled, canDisable) {
        if (!canDisable) return;
        const key = type + ':' + name;
        changes[key] = enabled;
        updateSaveStatus();
      }

      function getDefaultEnabled(type, name) {
        return true; // will be overridden by change handler
      }

      function updateSaveStatus() {
        const count = Object.keys(changes).length;
        document.getElementById('saveBtn').disabled = count === 0;
        const status = document.getElementById('saveStatus');
        if (count === 0) {
          status.textContent = '';
          status.className = '';
        } else {
          status.textContent = count + ' change(s) pending';
          status.className = '';
        }
      }

      async function saveChanges() {
        const btn = document.getElementById('saveBtn');
        const status = document.getElementById('saveStatus');
        btn.disabled = true;
        status.textContent = 'Saving...';
        status.className = '';

        const changeList = Object.entries(changes).map(([key, enable]) => {
          const [type, ...nameParts] = key.split(':');
          return { type, name: nameParts.join(':'), enable };
        });

        try {
          const res = await fetch('/api/panel/save', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'X-Panel-Session': sessionToken,
            },
            body: JSON.stringify({ channel: currentChannel, changes: changeList }),
          });
          if (!res.ok) {
            status.textContent = 'Failed to save changes.';
            status.className = 'error';
            btn.disabled = false;
            return;
          }
          const data = await res.json();
          const parts = [];
          if (data.commands_updated > 0) parts.push(data.commands_updated + ' command(s)');
          if (data.listeners_updated > 0) parts.push(data.listeners_updated + ' listener(s)');
          status.textContent = 'Saved: ' + parts.join(' and ') + ' updated.';
          status.className = '';
          changes = {};
          updateSaveStatus();
          loadCommands(currentChannel);
          loadListeners(currentChannel);
        } catch (e) {
          status.textContent = 'Network error.';
          status.className = 'error';
          btn.disabled = false;
        }
      }

      async function logout() {
        await fetch('/api/auth/logout', {
          method: 'POST',
          headers: { 'X-Panel-Session': sessionToken },
        });
        sessionToken = null;
        resetLoginScreen();
        document.getElementById('panelView').classList.add('hidden');
        document.getElementById('loginScreen').classList.remove('hidden');
      }

      function resetLoginScreen() {
        document.getElementById('loginDesc').textContent = 'Log in with your Twitch account to manage commands and listeners.';
        document.getElementById('loginBtnText').textContent = 'Log in with Twitch';
        document.getElementById('loginLoading').classList.add('hidden');
        document.getElementById('loginBtn').disabled = false;
      }

      // ---- Bertcheck management ----

      function resetBertTab() {
        const overview = document.getElementById('bertOverview');
        overview.className = 'loading';
        overview.textContent = 'Loading bertcheck data...';
        document.getElementById('bertOverviewMeta').textContent = '';
        document.getElementById('bertUserSelect').innerHTML = '';
        clearBertUser();
        setBertStatus('', false);
        document.getElementById('bertPartial').checked = false;
        document.getElementById('bertPartial').disabled = false;
        document.getElementById('bertApproximate').checked = false;
        const countInput = document.getElementById('bertCount');
        countInput.value = '';
        countInput.disabled = true;
        setBertLastHours(2);
      }

      // Approximate mode ignores the log coverage requirement, so the partial
      // option is meaningless while it is on.
      function toggleBertApproximate() {
        const approximate = document.getElementById('bertApproximate').checked;
        const partial = document.getElementById('bertPartial');
        if (approximate) partial.checked = false;
        partial.disabled = approximate;

        const count = document.getElementById('bertCount');
        count.disabled = !approximate;
        if (!approximate) count.value = '';
      }

      function clearBertUser() {
        document.getElementById('bertUserSummary').innerHTML = '';
        document.getElementById('bertRecords').innerHTML = '';
        document.getElementById('bertLog').innerHTML = '';
      }

      async function loadBertOverview() {
        if (!currentChannel) return;
        const overview = document.getElementById('bertOverview');
        overview.className = 'loading';
        overview.textContent = 'Loading bertcheck data...';
        try {
          const res = await fetch('/api/panel/bert/overview?channel=' + encodeURIComponent(currentChannel), {
            headers: { 'X-Panel-Session': sessionToken },
          });
          if (!res.ok) {
            overview.className = 'loading error';
            overview.textContent = await panelErrorText(res, 'Failed to load bertcheck data.');
            return;
          }
          const data = await res.json();
          renderBertOverview(data);
          const select = document.getElementById('bertUserSelect');
          const previous = select.value;
          if (data.users.length === 0) {
            select.innerHTML = '<option value="">No users</option>';
          } else {
            select.innerHTML = data.users.map(u =>
              `<option value="${escapeHtml(u.username)}">${escapeHtml(u.username)} (${u.total})</option>`
            ).join('');
            if (previous && data.users.some(u => u.username === previous)) {
              select.value = previous;
            }
          }
        } catch (e) {
          overview.className = 'loading error';
          overview.textContent = 'Error loading bertcheck data.';
        }
      }

      function renderBertOverview(data) {
        document.getElementById('bertOverviewMeta').textContent = data.user_count + ' user(s)';
        const berts = data.berts || [];
        const logRange = data.log_entries
          ? formatBertTime(data.log_starts) + ' -> ' + formatBertTime(data.log_ends)
          : 'no activations logged yet';
        const chips = berts.length
          ? berts.map(b => `<span class="bert-chip">${escapeHtml(b)}</span>`).join('')
          : '<span class="hint">No active berts</span>';
        const overview = document.getElementById('bertOverview');
        overview.className = '';
        overview.innerHTML = `
          <div class="stat-grid">
            <div class="stat"><span>Total bert checks</span><strong>${data.total_activations.toLocaleString()}</strong></div>
            <div class="stat"><span>Users</span><strong>${data.user_count.toLocaleString()}</strong></div>
            <div class="stat"><span>Active berts</span><strong>${berts.length.toLocaleString()}</strong></div>
            <div class="stat"><span>Logged activations</span><strong>${data.log_entries.toLocaleString()}</strong></div>
          </div>
          <div class="hint">Activation log: ${escapeHtml(logRange)}</div>
          <div class="bert-chips">${chips}</div>
        `;
      }

      async function loadBertUser() {
        const user = document.getElementById('bertUserSelect').value;
        if (!currentChannel || !user) {
          setBertStatus('Choose a user first.', true);
          return;
        }
        setBertStatus('Loading user...', false);
        try {
          const url = '/api/panel/bert/user?channel=' + encodeURIComponent(currentChannel) +
            '&user=' + encodeURIComponent(user) + '&limit=50';
          const res = await fetch(url, {
            headers: { 'X-Panel-Session': sessionToken },
          });
          if (!res.ok) {
            setBertStatus(await panelErrorText(res, 'Failed to load user.'), true);
            return;
          }
          const data = await res.json();
          renderBertUser(data);
          setBertStatus('', false);
        } catch (e) {
          setBertStatus('Error loading user.', true);
        }
      }

      function renderBertUser(data) {
        const logNote = data.channel_log_entries === 0
          ? 'No activation log exists for this channel yet. Use approximate purge for older data; new activations will be logged automatically.'
          : (data.log_entries === 0
            ? 'No activations logged for this user yet.'
            : data.log_entries + ' logged activation(s): ' + formatBertTime(data.log_starts) + ' -> ' + formatBertTime(data.log_ends));

        document.getElementById('bertUserSummary').innerHTML = `
          <div class="bert-user-head">
            <strong>${escapeHtml(data.username)}</strong>
            <span>${data.total.toLocaleString()} bert check(s)</span>
          </div>
          <div class="hint">
            ${data.zazas.toLocaleString()} zazas · ${data.zaza_ls.toLocaleString()} zazaLs ·
            ${data.double_zazas.toLocaleString()} double zazas · ${escapeHtml(logNote)}
          </div>
        `;

        const records = document.getElementById('bertRecords');
        if (data.records.length === 0) {
          records.innerHTML = '<p class="hint">No bert records.</p>';
        } else {
          records.innerHTML = `
            <table class="bert-table">
              <thead><tr>
                <th>Bert</th><th>Count</th><th>Zazas</th><th>ZazaLs</th><th>Double</th>
                <th>First seen</th><th>Last seen</th>
              </tr></thead>
              <tbody>
                ${data.records.map(r => `<tr>
                  <td class="bert-name">${escapeHtml(r.bert)}</td>
                  <td>${r.count}</td>
                  <td>${r.zazas}</td>
                  <td>${r.zaza_ls}</td>
                  <td>${r.double_zazas}</td>
                  <td>${formatBertTime(r.first_seen)}</td>
                  <td>${formatBertTime(r.last_seen)}</td>
                </tr>`).join('')}
              </tbody>
            </table>`;
        }

        const log = document.getElementById('bertLog');
        if (data.log.length === 0) {
          log.innerHTML = '<p class="hint">No logged activations to show.</p>';
        } else {
          log.innerHTML = `
            <table class="bert-table bert-log">
              <thead><tr><th>Time</th><th>Bert</th><th>Zaza</th><th>ZazaL</th><th>Double</th></tr></thead>
              <tbody>
                ${data.log.map(a => `<tr>
                  <td>${formatBertTime(a.timestamp)}</td>
                  <td class="bert-name">${escapeHtml(a.bert)}</td>
                  <td>${markBertFlag(a.is_zaza)}</td>
                  <td>${markBertFlag(a.is_zaza_l)}</td>
                  <td>${markBertFlag(a.double_zaza)}</td>
                </tr>`).join('')}
              </tbody>
            </table>`;
        }
      }

      async function purgeBertUser() {
        const user = document.getElementById('bertUserSelect').value;
        if (!user) {
          setBertStatus('Choose a user first.', true);
          return;
        }

        const fromInput = document.getElementById('bertFrom').value;
        const untilInput = document.getElementById('bertUntil').value;
        if (!fromInput || !untilInput) {
          setBertStatus('Choose a start and end time.', true);
          return;
        }
        const from = new Date(fromInput);
        const until = new Date(untilInput);
        if (isNaN(from.getTime()) || isNaN(until.getTime()) || from >= until) {
          setBertStatus('The start of the range must be before the end.', true);
          return;
        }

        const partial = document.getElementById('bertPartial').checked;
        const approximate = document.getElementById('bertApproximate').checked;
        const countInput = document.getElementById('bertCount');
        let count = 0;
        if (approximate && countInput.value.trim() !== '') {
          count = Number(countInput.value);
          if (!Number.isInteger(count) || count <= 0) {
            setBertStatus('Enter a positive whole number of berts to remove, or leave it blank to estimate.', true);
            return;
          }
        }

        const range = formatBertTime(from) + ' and ' + formatBertTime(until);
        let message;
        if (approximate && count > 0) {
          message = `Approximately remove ${count} bertcheck(s) for ${user} between ${range}?\n\nBerts are picked at random. This updates the saved stats and cannot be undone.`;
        } else if (approximate) {
          message = `Approximately purge ${user}'s bertchecks between ${range}?\n\nThe number of berts is estimated from the hourly counters and picked at random. This updates the saved stats and cannot be undone.`;
        } else {
          message = `Delete ${user}'s bertchecks between ${range}?\n\nThis updates the saved stats and cannot be undone.`;
        }
        if (!confirm(message)) {
          return;
        }

        setBertStatus('Removing activations...', false);
        try {
          const res = await fetch('/api/panel/bert/purge', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'X-Panel-Session': sessionToken,
            },
            body: JSON.stringify({
              channel: currentChannel,
              user,
              from: from.toISOString(),
              until: until.toISOString(),
              partial,
              approximate,
              count,
            }),
          });
          if (!res.ok) {
            setBertStatus(await panelErrorText(res, 'Purge failed.'), true);
            return;
          }
          const data = await res.json();
          await loadBertOverview();
          const select = document.getElementById('bertUserSelect');
          if (Array.from(select.options).some(o => o.value === user)) {
            select.value = user;
            await loadBertUser();
          } else {
            clearBertUser();
          }
          renderPurgeResult(data, user);
        } catch (e) {
          setBertStatus('Network error while purging.', true);
        }
      }

      function renderPurgeResult(data, user) {
        const byBert = Object.entries(data.removed_by_bert || {})
          .map(([bert, count]) => `${escapeHtml(bert)} ×${count}`)
          .join(', ');
        let html = `Removed <strong>${data.removed}</strong> activation(s) for <strong>${escapeHtml(user)}</strong>.`;
        if (byBert) html += `<div class="hint">${byBert}</div>`;
        if (data.approximate) {
          html += '<div class="hint warn">Approximate purge: berts were chosen at random from the hourly counters. Zaza counts were clamped to the remaining activations.</div>';
        }
        if (data.partial) {
          html += `<div class="hint warn">Partial purge: the log starts at ${formatBertTime(data.log_starts)}, so older aggregates were left unchanged.</div>`;
        }
        if (data.user_removed) {
          html += '<div class="hint">The user has no activity left and was removed from the channel.</div>';
        }
        if (data.warnings && data.warnings.length) {
          html += `<div class="hint warn">${data.warnings.map(escapeHtml).join('<br>')}</div>`;
        }
        setBertStatus(html, false, true);
      }

      async function deleteBertUser() {
        const user = document.getElementById('bertUserSelect').value;
        if (!user) {
          setBertStatus('Choose a user first.', true);
          return;
        }
        if (!confirm(`Remove ${user} from ${currentChannel} entirely?\n\nThis deletes all of their bertcheck history in this channel and cannot be undone.`)) {
          return;
        }

        setBertStatus('Removing user...', false);
        try {
          const res = await fetch('/api/panel/bert/delete-user', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'X-Panel-Session': sessionToken,
            },
            body: JSON.stringify({ channel: currentChannel, user }),
          });
          if (!res.ok) {
            setBertStatus(await panelErrorText(res, 'Failed to delete user.'), true);
            return;
          }
          const data = await res.json();
          await loadBertOverview();
          clearBertUser();
          setBertStatus(
            `Removed <strong>${escapeHtml(user)}</strong> from the channel (${data.removed} logged activation(s)).`,
            false,
            true
          );
        } catch (e) {
          setBertStatus('Network error while deleting the user.', true);
        }
      }

      function setBertStatus(message, isError, isHtml) {
        const status = document.getElementById('bertActionStatus');
        if (isHtml) status.innerHTML = message;
        else status.textContent = message;
        status.className = 'bert-status' + (isError ? ' error' : '');
      }

      function setBertLastHours(hours) {
        const until = new Date();
        const from = new Date(until.getTime() - hours * 60 * 60 * 1000);
        document.getElementById('bertFrom').value = bertLocalInputValue(from);
        document.getElementById('bertUntil').value = bertLocalInputValue(until);
      }

      function bertLocalInputValue(date) {
        const pad = value => String(value).padStart(2, '0');
        return date.getFullYear() + '-' + pad(date.getMonth() + 1) + '-' + pad(date.getDate()) +
          'T' + pad(date.getHours()) + ':' + pad(date.getMinutes());
      }

      function formatBertTime(value) {
        if (!value) return '-';
        const date = new Date(value);
        if (isNaN(date.getTime()) || date.getFullYear() <= 1) return '-';
        return date.toLocaleString();
      }

      function markBertFlag(value) {
        return value ? 'yes' : '-';
      }

      function escapeHtml(value) {
        return String(value == null ? '' : value).replace(/[&<>"']/g, ch => ({
          '&': '&amp;',
          '<': '&lt;',
          '>': '&gt;',
          '"': '&quot;',
          "'": '&#39;',
        })[ch]);
      }

      async function panelErrorText(res, fallback) {
        try {
          const text = (await res.text()).trim();
          return text || fallback;
        } catch (e) {
          return fallback;
        }
      }

      // Bootstrap
      (async function() {
        handleOAuthRedirect();
        const hasSession = await checkSession();
        if (!hasSession) {
          document.getElementById('panelView').classList.add('hidden');
          document.getElementById('loginScreen').classList.remove('hidden');
        }
      })();
    