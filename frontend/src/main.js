import * as Go from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { t, setLang, getLang, apply } from './i18n';

const SOURCE_URL = 'https://github.com/justech0/TelTun';
const STEPS = ['stepPrepare', 'stepDevice', 'stepPush', 'stepPhone', 'stepLink', 'stepTest', 'stepTun'];

const $ = (id) => document.getElementById(id);
let settings = {};
let status = { state: 'idle' };
let devices = [];
let stats = { down: 0, up: 0, totalDown: 0, totalUp: 0, streams: 0 };
let hotspot = null;
let hsBusy = false;
let doneSteps = new Set();
const hist = { down: new Array(60).fill(0), up: new Array(60).fill(0) };

// ---------- helpers ----------
function mbps(bps) { return (bps * 8 / 1e6); }
function fmtMbps(bps) { const v = mbps(bps); return v >= 100 ? v.toFixed(0) : v.toFixed(1); }
function fmtBytes(n) {
  if (!n) return '0 MB';
  if (n < 1e9) return (n / 1e6).toFixed(1) + ' MB';
  return (n / 1e9).toFixed(2) + ' GB';
}
function fmtDur(s) {
  const h = Math.floor(s / 3600), m = Math.floor(s % 3600 / 60), x = Math.floor(s % 60);
  return (h ? h + ':' : '') + String(m).padStart(h ? 2 : 1, '0') + ':' + String(x).padStart(2, '0');
}
function copyText(s) { navigator.clipboard.writeText(s).catch(() => {}); }
function toast(msg) {
  const el = $('toast');
  el.textContent = msg; el.hidden = false;
  clearTimeout(toast.t); toast.t = setTimeout(() => (el.hidden = true), 2600);
}

// ---------- navigation ----------
document.querySelectorAll('.nav').forEach(b => b.addEventListener('click', () => {
  document.querySelectorAll('.nav').forEach(x => x.classList.toggle('active', x === b));
  document.querySelectorAll('.page').forEach(p => p.classList.toggle('active', p.id === 'page-' + b.dataset.page));
  if (b.dataset.page === 'hotspot') refreshHotspot();
  if (b.dataset.page === 'logs') { const l = $('logs'); l.scrollTop = l.scrollHeight; }
}));

// ---------- title bar ----------
$('btnMin').onclick = () => Go.Minimise();
$('btnClose').onclick = () => (settings.closeToTray ? Go.Hide() : Go.Quit());

// ---------- language ----------
function useLang(l, save = true) {
  setLang(l);
  apply();
  document.querySelectorAll('[data-lang]').forEach(b => b.classList.toggle('on', b.dataset.lang === getLang()));
  renderStatus(); renderDevices(); renderHotspot();
  if (save) { settings.language = getLang(); Go.SaveSettings(settings); }
}
document.querySelectorAll('[data-lang]').forEach(b => b.addEventListener('click', () => useLang(b.dataset.lang)));

// ---------- connect ----------
$('orb').onclick = () => {
  if (status.state === 'connected') Go.Disconnect();
  else if (status.state === 'idle') { doneSteps.clear(); Go.Connect(); }
};

function renderStatus() {
  const s = status.state;
  const orb = $('orb');
  orb.dataset.state = s;
  orb.disabled = s === 'connecting' || s === 'disconnecting';
  $('orbLabel').textContent = s === 'connected' ? t('btnDisconnect') : s === 'idle' ? t('btnConnect') : t('btnWait');

  const key = { idle: 'stIdle', connecting: 'stConnecting', connected: 'stConnected', disconnecting: 'stDisconnecting' }[s];
  $('stateText').textContent = t(key);
  $('sideState').textContent = t(key);
  $('statePill').dataset.s = s;
  $('sideState').dataset.s = s;

  // idle: show the phone that is ready, or say none is found
  const ready = (devices || []).find(d => d.state === 'device' && (!settings.device || d.serial === settings.device));
  $('heroTitle').textContent = s === 'connected' ? (status.device || t('stConnected'))
    : s === 'idle' ? (ready ? (ready.model || ready.serial) : t('heroWait')) : t(key);
  $('heroSub').textContent = s === 'connected' ? t('subConnected') : s === 'idle' ? (ready ? t('subReady') : t('subIdle')) : '';

  // progress list while connecting
  const ol = $('steps');
  ol.innerHTML = '';
  if (s === 'connecting') {
    if (status.step) {
      const i = STEPS.indexOf(status.step);
      STEPS.slice(0, i).forEach(k => doneSteps.add(k));
    }
    STEPS.forEach(k => {
      const li = document.createElement('li');
      li.textContent = t(k);
      if (k === status.step) li.className = 'now';
      else if (doneSteps.has(k)) li.className = 'done';
      ol.appendChild(li);
    });
  }

  const al = $('alert');
  if (s === 'idle' && status.error) {
    al.hidden = false;
    al.innerHTML = '';
    al.append(t(status.error));
    if (status.detail) { const sm = document.createElement('small'); sm.textContent = status.detail; al.append(sm); }
  } else al.hidden = true;

  $('fDevice').textContent = status.device || t('none');
  renderIp();
  if (s !== 'connected') {
    $('fUp').textContent = t('none'); $('fConn').textContent = t('none');
  }
  $('devices').hidden = s !== 'idle';
}

// Public IP is hidden until the user asks to see it.
let showIp = false;
function renderIp() {
  const ip = status.publicIp;
  $('fIp').textContent = !ip ? t('none') : showIp ? ip : '•••.•••.•••.•••';
  const eye = $('ipEye');
  eye.classList.toggle('on', showIp);
  eye.title = showIp ? t('hide') : t('show');
}
$('ipEye').onclick = () => { showIp = !showIp; renderIp(); };

function renderDevices() {
  const box = $('devices');
  box.innerHTML = '';
  if (!devices || !devices.length) {
    const d = document.createElement('span');
    d.className = 'dev none'; d.innerHTML = '<svg viewBox="0 0 24 24"><rect x="7" y="2.5" width="10" height="19" rx="2.5"/><path d="M11 18h2"/></svg>';
    d.append(t('noPhones'));
    box.appendChild(d);
  } else {
    devices.forEach(dv => {
      const d = document.createElement('span');
      d.className = 'dev';
      const ok = dv.state === 'device';
      d.innerHTML = '<svg viewBox="0 0 24 24"><rect x="7" y="2.5" width="10" height="19" rx="2.5"/><path d="M11 18h2"/></svg>';
      if (!ok) d.classList.add('wait');
      const stKey = { device: 'stDevice', unauthorized: 'stUnauthorized' }[dv.state] || 'stOffline';
      const name = document.createElement('b');
      name.textContent = dv.model || dv.serial;
      const st = document.createElement('span');
      st.textContent = t(stKey);
      d.append(name, st);
      box.appendChild(d);
    });
  }
  // settings dropdown
  const sel = $('deviceSel');
  const cur = settings.device || '';
  sel.innerHTML = '';
  const opt = (v, txt) => { const o = document.createElement('option'); o.value = v; o.textContent = txt; sel.appendChild(o); };
  opt('', t('anyPhone'));
  (devices || []).forEach(dv => opt(dv.serial, `${dv.model || dv.serial} (${dv.serial})`));
  if (cur && !(devices || []).some(d => d.serial === cur)) opt(cur, cur);
  sel.value = cur;
}
$('deviceSel').onchange = (e) => { settings.device = e.target.value; Go.SaveSettings(settings); };

// ---------- speed charts ----------
function drawChart(canvas, data, c1, c2) {
  const dpr = window.devicePixelRatio || 1;
  const w = canvas.clientWidth, h = canvas.clientHeight;
  if (canvas.width !== w * dpr) { canvas.width = w * dpr; canvas.height = h * dpr; }
  const ctx = canvas.getContext('2d');
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.clearRect(0, 0, w, h);
  const max = Math.max(1, ...data) * 1.15;
  const step = w / (data.length - 1);
  const pts = data.map((v, i) => [i * step, h - (v / max) * (h - 4) - 2]);
  const grad = ctx.createLinearGradient(0, 0, w, 0);
  grad.addColorStop(0, c1); grad.addColorStop(1, c2);
  const fill = ctx.createLinearGradient(0, 0, 0, h);
  fill.addColorStop(0, c2 + '55'); fill.addColorStop(1, c2 + '00');
  ctx.beginPath();
  pts.forEach(([x, y], i) => {
    if (!i) ctx.moveTo(x, y);
    else { const [px, py] = pts[i - 1]; const mx = (px + x) / 2; ctx.bezierCurveTo(mx, py, mx, y, x, y); }
  });
  ctx.strokeStyle = grad; ctx.lineWidth = 2; ctx.stroke();
  ctx.lineTo(w, h); ctx.lineTo(0, h); ctx.closePath();
  ctx.fillStyle = fill; ctx.fill();
}
function renderStats() {
  $('downVal').textContent = fmtMbps(stats.down);
  $('upVal').textContent = fmtMbps(stats.up);
  drawChart($('downChart'), hist.down, '#22d3ee', '#6366f1');
  drawChart($('upChart'), hist.up, '#a855f7', '#ec4899');
  if (status.state === 'connected') {
    $('fData').textContent = fmtBytes(stats.totalDown + stats.totalUp);
    $('fConn').textContent = stats.streams;
  }
}
setInterval(() => {
  if (status.state === 'connected' && status.since) $('fUp').textContent = fmtDur(Date.now() / 1000 - status.since);
}, 1000);
window.addEventListener('resize', renderStats);

// ---------- hotspot ----------
let reveal = false;
async function refreshHotspot() {
  try { hotspot = await Go.HotspotStatus(); } catch (e) { hotspot = { state: 'Unknown', msg: String(e) }; }
  renderHotspot();
}
function renderHotspot() {
  const on = hotspot && hotspot.state === 'On';
  $('hsIcon').classList.toggle('on', on);
  $('hsState').textContent = hsBusy ? t('hsBusy') : on ? t('hsOn') : t('hsOff');
  $('hsMsg').textContent = hotspot && hotspot.msg ? hotspot.msg : (hotspot && hotspot.profile ? hotspot.profile : '');
  const b = $('hsBtn');
  b.textContent = on ? t('hsStop') : t('hsStart');
  b.disabled = hsBusy;
  $('hsSsid').textContent = hotspot && hotspot.ssid || t('none');
  const pass = hotspot && hotspot.pass;
  $('hsPassView').textContent = pass ? (reveal ? pass : '•'.repeat(Math.min(pass.length, 12))) : t('none');
  $('hsPassView').dataset.value = pass || '';
  $('hsReveal').textContent = reveal ? t('hide') : t('show');
  $('hsClients').textContent = hotspot && hotspot.max ? `${hotspot.clients} / ${hotspot.max}` : t('none');
}
$('hsBtn').onclick = async () => {
  const on = hotspot && hotspot.state === 'On';
  hsBusy = true; renderHotspot();
  try { hotspot = on ? await Go.HotspotStop() : await Go.HotspotStart(); }
  catch (e) { hotspot = Object.assign({}, hotspot || {}, { msg: String(e) }); toast(String(e)); }
  hsBusy = false; renderHotspot();
};
$('hsReveal').onclick = () => { reveal = !reveal; renderHotspot(); };
document.querySelectorAll('[data-copy]').forEach(b => b.addEventListener('click', () => {
  const el = $(b.dataset.copy);
  copyText(el.dataset.value || el.textContent);
  toast(t('copied'));
}));
$('hsSave').onclick = () => {
  const ssid = $('hsSsidIn').value.trim(), pass = $('hsPassIn').value;
  if (pass && pass.length < 8) { toast(t('hsPass') + ' ≥ 8'); return; }
  settings.hotspotSsid = ssid; settings.hotspotPass = pass;
  Go.SaveSettings(settings); toast(t('saved'));
};

// ---------- settings toggles ----------
document.querySelectorAll('[data-set]').forEach(el => el.addEventListener('change', () => {
  settings[el.dataset.set] = el.checked;
  Go.SaveSettings(settings);
}));
$('autostart').onchange = async (e) => {
  try { await Go.SetAutostart(e.target.checked); }
  catch (err) { e.target.checked = !e.target.checked; toast(String(err)); }
};

// ---------- logs ----------
function addLog(line) {
  const l = $('logs');
  const stick = l.scrollTop + l.clientHeight >= l.scrollHeight - 20;
  l.textContent += (l.textContent ? '\n' : '') + line;
  if (stick) l.scrollTop = l.scrollHeight;
}
$('logsCopy').onclick = () => { copyText($('logs').textContent); toast(t('copied')); };
$('logsOpen').onclick = () => Go.OpenLogFolder();
$('abSource').onclick = () => Go.OpenURL(SOURCE_URL);
$('abMail').onclick = () => Go.OpenURL('mailto:iletisim@enestunc.com.tr');
$('btnSpeedtest').onclick = () => Go.OpenURL('https://www.speedtest.net/');

// ---------- events from Go ----------
EventsOn('status', (s) => {
  if (s.state === 'connected' && status.state !== 'connected') {
    hist.down.fill(0); hist.up.fill(0);
  }
  status = s; renderStatus();
});
EventsOn('stats', (s) => {
  stats = s;
  hist.down.push(mbps(s.down)); hist.down.shift();
  hist.up.push(mbps(s.up)); hist.up.shift();
  renderStats();
});
EventsOn('devices', (d) => { devices = d || []; if (status.state === 'idle') { renderDevices(); renderStatus(); } });
EventsOn('log', addLog);
EventsOn('lost', () => toast(t('lost')));
EventsOn('hotspot', (h) => { hotspot = h; renderHotspot(); });

// ---------- boot ----------
(async () => {
  const snap = await Go.Snapshot();
  settings = snap.settings;
  status = snap.status;
  devices = snap.devices || [];
  $('abVer').textContent = snap.version;
  $('autostart').checked = snap.autostart;
  document.querySelectorAll('[data-set]').forEach(el => { el.checked = !!settings[el.dataset.set]; });
  $('hsSsidIn').value = settings.hotspotSsid || '';
  $('hsPassIn').value = settings.hotspotPass || '';
  (snap.logs || []).forEach(addLog);
  let l = settings.language;
  if (!l) l = (navigator.language || 'en').toLowerCase().startsWith('tr') ? 'tr' : 'en';
  useLang(l, !settings.language);
  renderStats();
})();
