
      "use strict";
      var CHANNEL = "sir_lysergium";
      var LS_BALLOTS = "bertvote.ballots.v1";
      var LS_VOTER = "bertvote.voter.v1";
      // BigBert has no single emote; the channel uses bigbert1 + bigbert2 side by side.
      var COMPOSITE = { BigBert: ["bigbert1", "bigbert2"] };

      var allBerts = [];
      var emoteByName = {};
      var current = []; // ordered best -> worst
      var ballots = [];
      var dragIdx = null;
      var dataLoaded = false;

      function $(id) { return document.getElementById(id); }

      function updateSubmitState() {
        var hasName = $("voterInput").value.trim().length > 0;
        $("submitBtn").disabled = !(dataLoaded && hasName);
      }

      function loadLocal() {
        try {
          var raw = localStorage.getItem(LS_BALLOTS);
          if (raw) {
            var parsed = JSON.parse(raw);
            if (Array.isArray(parsed)) ballots = parsed.filter(isValidBallot);
          }
        } catch (e) { ballots = []; }
        try {
          var v = localStorage.getItem(LS_VOTER);
          if (v) $("voterInput").value = v;
        } catch (e) {}
        $("voterInput").addEventListener("input", function () {
          try { localStorage.setItem(LS_VOTER, $("voterInput").value.trim()); } catch (e) {}
          updateSubmitState();
        });
      }

      function saveLocal() {
        try { localStorage.setItem(LS_BALLOTS, JSON.stringify(ballots)); } catch (e) {}
      }

      function isValidBallot(b) {
        return b && Array.isArray(b.ranking) && b.ranking.length === 5 &&
          b.ranking.every(function (n) { return typeof n === "string" && n.length > 0; });
      }

      function emoteImg(name, cls) {
        var parts = COMPOSITE[name];
        if (parts && parts.every(function (p) { return emoteByName[p]; })) {
          return parts.map(function (p) {
            return '<img src="https://cdn.7tv.app/emote/' + emoteByName[p] + '/4x.webp" alt="' + escapeHtml(name) + '" loading="lazy" />';
          }).join("");
        }
        var id = emoteByName[name];
        if (!id) return "";
        // onerror falls back to smaller sizes in case a 4x render does not exist
        return '<img src="https://cdn.7tv.app/emote/' + id + '/4x.webp" alt="' + escapeHtml(name) + '" loading="lazy" class="' + (cls || "") + '"' +
          ' onerror="if(!this.dataset.fb){this.dataset.fb=1;this.src=\'https://cdn.7tv.app/emote/' + id + '/2x.webp\';}" />';
      }

      function escapeHtml(s) {
        return String(s).replace(/[&<>"']/g, function (c) {
          return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
        });
      }

      function sampleFive() {
        var pool = allBerts.slice();
        // shuffle
        for (var i = pool.length - 1; i > 0; i--) {
          var j = Math.floor(Math.random() * (i + 1));
          var t = pool[i]; pool[i] = pool[j]; pool[j] = t;
        }
        var pick = pool.slice(0, 5);
        // avoid serving the exact same set twice in a row when possible
        if (pool.length > 5 && current.length === 5) {
          var same = pick.slice().sort().join("|") === current.slice().sort().join("|");
          if (same) {
            var k = Math.floor(Math.random() * pool.length);
            pick[0] = pool[k];
          }
        }
        return pick;
      }

      function renderRound() {
        var list = $("rankList");
        list.innerHTML = "";
        var labels = ["1st · best · 4 pts", "2nd · 3 pts", "3rd · 2 pts", "4th · 1 pt", "5th · worst · 0 pts"];
        current.forEach(function (name, idx) {
          var li = document.createElement("li");
          li.className = "rank-item";
          li.draggable = true;
          li.dataset.idx = idx;
          li.innerHTML =
            '<span class="rank-pos">' + labels[idx].split(" · ")[0] + '<small>' + labels[idx].split(" · ").slice(1).join(" · ") + '</small></span>' +
            '<span class="rank-emote">' + (emoteImg(name) || escapeHtml(name)) + '</span>' +
            '<span class="rank-name">' + escapeHtml(name) + '</span>' +
            '<span class="rank-controls">' +
            '<button data-act="up" data-i="' + idx + '" aria-label="Move ' + escapeHtml(name) + ' up"' + (idx === 0 ? " disabled" : "") + '>↑</button>' +
            '<button data-act="down" data-i="' + idx + '" aria-label="Move ' + escapeHtml(name) + ' down"' + (idx === current.length - 1 ? " disabled" : "") + '>↓</button>' +
            '</span>';
          li.addEventListener("dragstart", function (e) {
            dragIdx = idx;
            li.classList.add("dragging");
            try { e.dataTransfer.effectAllowed = "move"; } catch (err) {}
          });
          li.addEventListener("dragend", function () {
            dragIdx = null;
            list.querySelectorAll(".rank-item").forEach(function (el) {
              el.classList.remove("dragging", "dragover");
            });
          });
          li.addEventListener("dragover", function (e) {
            e.preventDefault();
            li.classList.add("dragover");
          });
          li.addEventListener("dragleave", function () { li.classList.remove("dragover"); });
          li.addEventListener("drop", function (e) {
            e.preventDefault();
            if (dragIdx === null || dragIdx === idx) return;
            var item = current.splice(dragIdx, 1)[0];
            current.splice(idx, 0, item);
            renderRound();
          });
          list.appendChild(li);
        });
        list.querySelectorAll("button[data-act]").forEach(function (btn) {
          btn.addEventListener("click", function () {
            var i = parseInt(btn.dataset.i, 10);
            if (btn.dataset.act === "up" && i > 0) {
              var t = current[i - 1]; current[i - 1] = current[i]; current[i] = t;
            } else if (btn.dataset.act === "down" && i < current.length - 1) {
              var u = current[i + 1]; current[i + 1] = current[i]; current[i] = u;
            }
            renderRound();
          });
        });
      }

      function bordaTotals(list) {
        var totals = {}, seen = {};
        list.forEach(function (b) {
          var n = b.ranking.length;
          b.ranking.forEach(function (name, pos) {
            totals[name] = (totals[name] || 0) + (n - 1 - pos);
            seen[name] = (seen[name] || 0) + 1;
          });
        });
        return { totals: totals, seen: seen };
      }

      function standingsHTML(totals, seen, limit) {
        var rows = Object.keys(totals).map(function (name) {
          return { name: name, pts: totals[name], n: seen[name] || 0 };
        });
        rows.sort(function (a, b) { return b.pts - a.pts || a.name.localeCompare(b.name); });
        if (!rows.length) return '<span class="muted">No votes yet.</span>';
        var html = '<table class="standings"><thead><tr><th>#</th><th>Bert</th><th class="num">Points</th><th class="num">Rounds</th></tr></thead><tbody>';
        rows.slice(0, limit || rows.length).forEach(function (r, i) {
          var img = emoteImg(r.name, "mini-emote");
          html += '<tr><td>' + (i + 1) + '</td><td>' + (img || "") + ' ' + escapeHtml(r.name) + '</td>' +
            '<td class="num">' + r.pts + '</td><td class="num">' + r.n + '</td></tr>';
        });
        return html + '</tbody></table>';
      }

      function renderMine() {
        var t = bordaTotals(ballots);
        $("myRounds").textContent = ballots.length;
        $("mySeen").textContent = Object.keys(t.seen).length;
        $("myTable").innerHTML = standingsHTML(t.totals, t.seen, 30);
        $("roundStatus").textContent = ballots.length
          ? "Saved locally: " + ballots.length + " round" + (ballots.length === 1 ? "" : "s") + " so far."
          : "";
        $("roundStatus").className = "status";
      }

      function setStatus(msg, isError) {
        var el = $("roundStatus");
        el.textContent = msg;
        el.className = "status" + (isError ? " error" : "");
      }

      function nextRound() {
        current = sampleFive();
        renderRound();
      }

      function submitCurrent() {
        if (current.length !== 5) return;
        var voter = $("voterInput").value.trim();
        if (!voter) {
          setStatus("Enter a display name before submitting.", true);
          $("voterInput").focus();
          return;
        }
        var ballot = {
          voter: voter,
          ranking: current.slice(), // best -> worst
          ts: new Date().toISOString()
        };
        ballots.push(ballot);
        saveLocal();
        // best-effort shared tally; local save already done above
        fetch("/api/vote/submit", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ voter: ballot.voter, ranking: ballot.ranking })
        }).then(function (res) {
          if (!res.ok) throw new Error("server returned " + res.status);
          setStatus("Vote saved locally and sent to the shared tally. New berts below.");
          refreshShared();
        }).catch(function () {
          setStatus("Vote saved locally. Shared tally unreachable, your vote is still stored in this browser.");
        });
        try { localStorage.setItem(LS_VOTER, ballot.voter); } catch (e) {}
        renderMine();
        nextRound();
      }

      function refreshShared() {
        fetch("/api/vote/results").then(function (res) {
          if (!res.ok) throw new Error("status " + res.status);
          return res.json();
        }).then(function (data) {
          $("sharedRounds").textContent = data.total_ballots != null ? data.total_ballots : 0;
          $("sharedTable").innerHTML = standingsHTML(data.totals || {}, data.appearances || {}, 30);
        }).catch(function () {
          $("sharedRounds").textContent = "–";
          $("sharedTable").innerHTML = '<span class="muted">Shared tally unavailable (backend not reachable). Local votes above are unaffected.</span>';
        });
      }

      function init() {
        loadLocal();
        renderMine();
        updateSubmitState();
        $("submitBtn").addEventListener("click", submitCurrent);
        $("skipBtn").addEventListener("click", nextRound);
        $("refreshSharedBtn").addEventListener("click", refreshShared);
        $("resetBtn").addEventListener("click", function () {
          if (!ballots.length) return;
          if (!confirm("Delete all " + ballots.length + " locally saved round(s) in this browser?")) return;
          ballots = [];
          saveLocal();
          renderMine();
        });

        Promise.all([
          fetch("/api/channel?name=" + CHANNEL).then(function (r) {
            if (!r.ok) throw new Error("channel api: " + r.status);
            return r.json();
          }),
          fetch("/api/emotes?channel=" + CHANNEL).then(function (r) {
            if (!r.ok) throw new Error("emotes api: " + r.status);
            return r.json();
          })
        ]).then(function (res) {
          allBerts = res[0].active_berts || [];
          (res[1] || []).forEach(function (em) {
            if (em && em.name && em.id) emoteByName[em.name] = em.id;
          });
          if (!allBerts.length) throw new Error("no berts returned");
          $("loadingLabel").style.display = "none";
          dataLoaded = true;
          updateSubmitState();
          $("skipBtn").disabled = false;
          nextRound();
        }).catch(function (err) {
          $("loadingLabel").textContent = "Failed to load bert data: " + err.message;
          $("loadingLabel").className = "status error";
        });
        refreshShared();
      }

      document.addEventListener("DOMContentLoaded", init);
    