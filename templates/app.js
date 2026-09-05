/* ============================================================
   homelab-reader · 共享前端逻辑
   ：主题切换 / 灵动岛导航 / 登录态渲染 / 工具函数
   ============================================================ */

const API = {
  books: '/api/books',
  upload: '/api/books',
  rss: '/api/rss',
  login: '/api/auth/login',
  register: '/api/auth/register',
  logout: '/api/auth/logout',
};

/* ---------------- 工具函数 ---------------- */
function escapeHtml(s) {
  return String(s == null ? '' : s).replace(/[&<>"']/g, c =>
    ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}
function fmtSize(n) {
  const v = Number(n || 0);
  if (v >= 1 << 30) return (v / (1 << 30)).toFixed(1) + ' GB';
  if (v >= 1 << 20) return (v / (1 << 20)).toFixed(1) + ' MB';
  if (v >= 1 << 10) return (v / (1 << 10)).toFixed(1) + ' KB';
  return v + ' B';
}
async function jfetch(url, opts = {}) {
  const res = await fetch(url, { credentials: 'include', ...opts });
  const ct = res.headers.get('content-type') || '';
  const body = ct.includes('json') ? await res.json() : await res.text();
  return { ok: res.ok, status: res.status, body };
}
function toast(msg, ms = 2200) {
  let el = document.getElementById('toast');
  if (!el) { el = document.createElement('div'); el.id = 'toast'; el.className = 'toast'; document.body.appendChild(el); }
  el.textContent = msg;
  el.classList.add('show');
  clearTimeout(el._t);
  el._t = setTimeout(() => el.classList.remove('show'), ms);
}

/* ---------------- 主题切换（白天/黑夜） ---------------- */
const THEME_KEY = 'hl_theme';
function getTheme() { return localStorage.getItem(THEME_KEY) || 'light'; }
function applyTheme(t) {
  document.documentElement.setAttribute('data-theme', t);
  localStorage.setItem(THEME_KEY, t);
  const btn = document.getElementById('themeToggle');
  if (btn) btn.textContent = t === 'dark' ? '☀' : '☾';
}
function initTheme() { applyTheme(getTheme()); }
function toggleTheme() { applyTheme(getTheme() === 'dark' ? 'light' : 'dark'); }

/* ---------------- 登录态（localStorage 记录用户名） ---------------- */
function getStoredUsername() { return localStorage.getItem('hl_username') || ''; }
function setStoredUsername(name) {
  if (name) localStorage.setItem('hl_username', name);
  else localStorage.removeItem('hl_username');
}
function renderAuthChip() {
  const wrap = document.getElementById('navRight');
  if (!wrap) return;
  const name = getStoredUsername();
  if (name) {
    wrap.innerHTML = `
      <div class="user-chip" title="${escapeHtml(name)}">
        <div class="avatar">${escapeHtml(name.slice(0, 1).toUpperCase())}</div>
        <span style="display:none">${escapeHtml(name)}</span>
        <button class="btn btn-ghost" id="btnLogout">退出</button>
      </div>`;
    const b = document.getElementById('btnLogout');
    if (b) b.addEventListener('click', logout);
  } else {
    wrap.innerHTML = `
      <button class="btn btn-ghost" data-auth-mode="login">登录</button>
      <button class="btn btn-primary" data-auth-mode="register">注册</button>`;
    wrap.querySelectorAll('[data-auth-mode]').forEach(el =>
      el.addEventListener('click', () => openAuthModal(el.dataset.authMode)));
  }
}

/* ---------------- 登出（服务端失效会话 + 本地重置） ---------------- */
async function logout() {
  // 请求服务端删除会话记录，并让浏览器清除 cookie
  try { await fetch(API.logout, { method: 'POST', credentials: 'include' }); } catch (e) {}
  setStoredUsername('');
  renderAuthChip();
  // 通知各页面（书架等）重新加载，切换为未登录态
  document.dispatchEvent(new CustomEvent('hl:logout'));
  toast('已退出登录');
}

/* ---------------- 登录/注册弹窗 ---------------- */
function openAuthModal(mode) {
  const mask = document.getElementById('authMask');
  if (!mask) { toast('请先在概览页登录'); return; }
  setAuthMode(mode);
  mask.classList.add('open');
}
function setAuthMode(mode) {
  const mask = document.getElementById('authMask');
  if (!mask) return;
  const isLogin = mode === 'login';
  const mt = document.getElementById('modalTitle');
  if (mt) mt.textContent = isLogin ? '登录' : '注册';
  const sb = document.getElementById('submitBtn');
  if (sb) sb.textContent = isLogin ? '登 录' : '注 册';
  const tl = document.getElementById('tabLogin');
  const tr = document.getElementById('tabRegister');
  if (tl) tl.classList.toggle('active', isLogin);
  if (tr) tr.classList.toggle('active', !isLogin);
  const msg = document.getElementById('authMsg');
  if (msg) msg.textContent = '';
}
function initAuth(loginSuccessCb) {
  const mask = document.getElementById('authMask');
  if (!mask) return;
  const close = () => mask.classList.remove('open');
  const cx = document.getElementById('closeAuth');
  if (cx) cx.addEventListener('click', close);
  mask.addEventListener('click', e => { if (e.target === mask) close(); });
  const tl = document.getElementById('tabLogin');
  const tr = document.getElementById('tabRegister');
  if (tl) tl.addEventListener('click', () => setAuthMode('login'));
  if (tr) tr.addEventListener('click', () => setAuthMode('register'));

  const form = document.getElementById('authForm');
  form.addEventListener('submit', async e => {
    e.preventDefault();
    const username = document.getElementById('username').value.trim();
    const password = document.getElementById('password').value;
    const msg = document.getElementById('authMsg');
    if (!username || !password) { msg.className = 'msg err'; msg.textContent = '用户名和密码不能为空'; return; }
    const isLogin = document.getElementById('submitBtn').textContent.trim().startsWith('登');
    msg.className = 'msg'; msg.textContent = '正在提交…';
    const endpoint = isLogin ? API.login : API.register;
    const r = await jfetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });
    if (r.ok) {
      if (isLogin) {
        setStoredUsername(username);
        msg.className = 'msg ok'; msg.textContent = '登录成功';
        renderAuthChip();
        form.reset();
        setTimeout(() => { close(); toast('欢迎回来，' + username); if (loginSuccessCb) loginSuccessCb(); }, 300);
      } else {
        msg.className = 'msg ok'; msg.textContent = '注册成功，请切换登录';
      }
    } else {
      msg.className = 'msg err';
      msg.textContent = (typeof r.body === 'string' && r.body) || ('请求失败 ' + r.status);
    }
  });
}

/* ---------------- 灵动岛导航高亮 ---------------- */
function initNav() {
  const path = location.pathname;
  const map = { '/dashboard': 'overview', '/dashboard/books': 'shelf', '/dashboard/rss': 'rss', '/dashboard/user': 'profile' };
  const active = map[path] || 'overview';
  document.querySelectorAll('.nav-link').forEach(l =>
    l.classList.toggle('active', l.dataset.view === active));
  const tt = document.getElementById('themeToggle');
  if (tt) tt.addEventListener('click', toggleTheme);
}

/* ---------------- 启动 ---------------- */
document.addEventListener('DOMContentLoaded', () => {
  initTheme();
  initNav();
  renderAuthChip();
});