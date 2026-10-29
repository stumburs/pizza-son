
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

      function selectChannel(channel) {
        currentChannel = channel;
        document.querySelectorAll('.channel-btn').forEach(b => {
          b.classList.toggle('active', b.textContent === '#' + channel);
        });
        pendingChanges = { commands: {}, listeners: {} };
        changes = {};
        updateSaveStatus();
        loadCommands(channel);
        loadListeners(channel);
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

      // Bootstrap
      (async function() {
        handleOAuthRedirect();
        const hasSession = await checkSession();
        if (!hasSession) {
          document.getElementById('panelView').classList.add('hidden');
          document.getElementById('loginScreen').classList.remove('hidden');
        }
      })();
    