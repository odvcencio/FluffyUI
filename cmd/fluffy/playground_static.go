package main

const playgroundHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>FluffyUI Playground</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
:root{
  --bg:#0d1117;--bg-secondary:#161b22;--bg-tertiary:#21262d;
  --border:#30363d;--text:#e6edf3;--text-secondary:#8b949e;
  --accent:#58a6ff;--accent-hover:#79c0ff;--success:#3fb950;
  --danger:#f85149;--warning:#d29922;
}
body{background:var(--bg);color:var(--text);font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;height:100vh;display:flex;flex-direction:column;overflow:hidden}
header{background:var(--bg-secondary);border-bottom:1px solid var(--border);padding:8px 16px;display:flex;align-items:center;gap:12px;flex-shrink:0}
header .logo{font-weight:700;font-size:16px;color:var(--accent)}
header .logo span{color:var(--text-secondary);font-weight:400}
.toolbar{display:flex;gap:8px;margin-left:auto;align-items:center}
.toolbar button{border:1px solid var(--border);border-radius:6px;padding:5px 14px;font-size:13px;cursor:pointer;font-weight:500;transition:all .15s}
#runBtn{background:var(--success);color:#fff;border-color:var(--success)}
#runBtn:hover{background:#2ea043}
#runBtn:disabled{opacity:.5;cursor:not-allowed}
#stopBtn{background:var(--danger);color:#fff;border-color:var(--danger);display:none}
#stopBtn:hover{background:#da3633}
.status{font-size:12px;color:var(--text-secondary)}
.status.running{color:var(--success)}
.status.error{color:var(--danger)}
.status.building{color:var(--warning)}
main{display:flex;flex:1;overflow:hidden}
.pane{display:flex;flex-direction:column;flex:1;min-width:0}
.pane-header{background:var(--bg-secondary);border-bottom:1px solid var(--border);padding:6px 12px;font-size:12px;font-weight:600;color:var(--text-secondary);text-transform:uppercase;letter-spacing:.5px;flex-shrink:0}
.divider{width:3px;background:var(--border);cursor:col-resize;flex-shrink:0;transition:background .15s}
.divider:hover,.divider.dragging{background:var(--accent)}
#editor{flex:1}
#terminal-container{flex:1;padding:4px;background:#000}
.output-pane .pane-header{display:flex;justify-content:space-between;align-items:center}
.output-tabs{display:flex;gap:2px}
.output-tab{padding:2px 8px;border-radius:4px;cursor:pointer;font-size:11px;font-weight:500}
.output-tab.active{background:var(--bg-tertiary);color:var(--text)}
#build-output{flex:1;padding:12px;font-family:'JetBrains Mono','Fira Code',monospace;font-size:13px;overflow:auto;white-space:pre-wrap;background:var(--bg);display:none;color:var(--danger)}
.kbd{background:var(--bg-tertiary);border:1px solid var(--border);border-radius:3px;padding:1px 5px;font-size:11px;font-family:inherit}
</style>
<link rel="stylesheet" data-name="vs/editor/editor.main" href="https://cdn.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs/editor/editor.main.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@xterm/xterm@5.5.0/css/xterm.css">
</head>
<body>
<header>
  <div class="logo">FluffyUI <span>Playground</span></div>
  <div class="toolbar">
    <span class="status" id="status">Ready</span>
    <button id="runBtn">&#9654; Run <span class="kbd">Ctrl+Enter</span></button>
    <button id="stopBtn">&#9632; Stop</button>
  </div>
</header>
<main>
  <div class="pane">
    <div class="pane-header">main.go</div>
    <div id="editor"></div>
  </div>
  <div class="divider" id="divider"></div>
  <div class="pane output-pane">
    <div class="pane-header">
      <span>Output</span>
      <div class="output-tabs">
        <div class="output-tab active" data-tab="terminal" id="tab-terminal">Terminal</div>
        <div class="output-tab" data-tab="build" id="tab-build">Build</div>
      </div>
    </div>
    <div id="terminal-container"></div>
    <div id="build-output"></div>
  </div>
</main>

<script src="https://cdn.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs/loader.js"></script>
<script src="https://cdn.jsdelivr.net/npm/@xterm/xterm@5.5.0/lib/xterm.js"></script>
<script src="https://cdn.jsdelivr.net/npm/@xterm/addon-fit@0.10.0/lib/addon-fit.js"></script>
<script>
const defaultCode = ` + "`" + `package main

import (
	"context"

	"github.com/odvcencio/fluffyui/fluffy"
)

func main() {
	app, _ := fluffy.NewApp()

	counter := fluffy.Signal(0)

	app.Root(fluffy.VStack(
		fluffy.Label("Welcome to FluffyUI!"),
		fluffy.ReactiveText(func() string {
			return fmt.Sprintf("Count: %d", counter.Get())
		}, counter),
		fluffy.Button("Click me!", func() {
			counter.Set(counter.Get() + 1)
		}),
	))

	app.Run(context.Background())
}
` + "`" + `;

// State
let editor, term, fitAddon, ws, sessionID;

// Monaco
require.config({ paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs' }});
require(['vs/editor/editor.main'], function() {
  editor = monaco.editor.create(document.getElementById('editor'), {
    value: defaultCode,
    language: 'go',
    theme: 'vs-dark',
    minimap: { enabled: false },
    fontSize: 14,
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
    lineNumbers: 'on',
    renderLineHighlight: 'line',
    scrollBeyondLastLine: false,
    automaticLayout: true,
    tabSize: 4,
    insertSpaces: false,
    padding: { top: 8 },
  });

  editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, run);
});

// xterm.js
term = new Terminal({
  cursorBlink: true,
  fontSize: 14,
  fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
  theme: {
    background: '#000000',
    foreground: '#e6edf3',
    cursor: '#58a6ff',
    selectionBackground: '#264f78',
  },
});
fitAddon = new FitAddon.FitAddon();
term.loadAddon(fitAddon);
term.open(document.getElementById('terminal-container'));
fitAddon.fit();
term.write('\x1b[2m[ Click Run or press Ctrl+Enter to start ]\x1b[0m\r\n');

window.addEventListener('resize', () => fitAddon.fit());

// Tab switching
document.querySelectorAll('.output-tab').forEach(tab => {
  tab.addEventListener('click', () => {
    document.querySelectorAll('.output-tab').forEach(t => t.classList.remove('active'));
    tab.classList.add('active');
    const target = tab.dataset.tab;
    document.getElementById('terminal-container').style.display = target === 'terminal' ? '' : 'none';
    document.getElementById('build-output').style.display = target === 'build' ? '' : 'none';
    if (target === 'terminal') fitAddon.fit();
  });
});

// Divider drag
const divider = document.getElementById('divider');
const mainEl = document.querySelector('main');
let dragging = false;
divider.addEventListener('mousedown', (e) => {
  dragging = true;
  divider.classList.add('dragging');
  e.preventDefault();
});
document.addEventListener('mousemove', (e) => {
  if (!dragging) return;
  const panes = mainEl.querySelectorAll('.pane');
  const rect = mainEl.getBoundingClientRect();
  const x = e.clientX - rect.left;
  const pct = (x / rect.width) * 100;
  if (pct > 20 && pct < 80) {
    panes[0].style.flex = 'none';
    panes[0].style.width = pct + '%';
    panes[1].style.flex = '1';
  }
});
document.addEventListener('mouseup', () => {
  dragging = false;
  divider.classList.remove('dragging');
  fitAddon.fit();
});

// Status
const statusEl = document.getElementById('status');
const runBtn = document.getElementById('runBtn');
const stopBtn = document.getElementById('stopBtn');
const buildOutput = document.getElementById('build-output');

function setStatus(text, cls) {
  statusEl.textContent = text;
  statusEl.className = 'status ' + (cls || '');
}

function showTab(name) {
  document.querySelectorAll('.output-tab').forEach(t => {
    t.classList.toggle('active', t.dataset.tab === name);
  });
  document.getElementById('terminal-container').style.display = name === 'terminal' ? '' : 'none';
  document.getElementById('build-output').style.display = name === 'build' ? '' : 'none';
  if (name === 'terminal') fitAddon.fit();
}

// Run
async function run() {
  if (ws) { ws.close(); ws = null; }
  term.reset();
  buildOutput.textContent = '';
  setStatus('Building...', 'building');
  runBtn.disabled = true;
  showTab('terminal');

  try {
    const resp = await fetch('/api/run', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code: editor.getValue() }),
    });
    const data = await resp.json();
    if (!resp.ok || data.error) {
      buildOutput.textContent = data.error || 'Build failed';
      setStatus('Build failed', 'error');
      showTab('build');
      runBtn.disabled = false;
      return;
    }
    sessionID = data.session;
    setStatus('Running', 'running');
    runBtn.style.display = 'none';
    stopBtn.style.display = '';
    connectWS(sessionID);
  } catch (e) {
    setStatus('Error: ' + e.message, 'error');
    runBtn.disabled = false;
  }
}

function connectWS(sid) {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  ws = new WebSocket(proto + '//' + location.host + '/ws/' + sid);
  ws.binaryType = 'arraybuffer';

  ws.onopen = () => {
    fitAddon.fit();
    // Send initial size
    ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }));
  };

  ws.onmessage = (e) => {
    if (e.data instanceof ArrayBuffer) {
      term.write(new Uint8Array(e.data));
    } else {
      term.write(e.data);
    }
  };

  ws.onclose = () => {
    setStatus('Stopped', '');
    runBtn.style.display = '';
    runBtn.disabled = false;
    stopBtn.style.display = 'none';
    ws = null;
  };

  ws.onerror = () => {
    setStatus('Connection error', 'error');
  };

  // Forward terminal input to WebSocket
  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(data);
    }
  });

  // Forward resize events
  term.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'resize', cols, rows }));
    }
  });
}

// Stop
async function stop() {
  if (!sessionID) return;
  try {
    await fetch('/api/stop/' + sessionID, { method: 'POST' });
  } catch (_) {}
  if (ws) { ws.close(); ws = null; }
  sessionID = null;
  setStatus('Stopped', '');
  runBtn.style.display = '';
  runBtn.disabled = false;
  stopBtn.style.display = 'none';
}

runBtn.addEventListener('click', run);
stopBtn.addEventListener('click', stop);

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
  if (e.ctrlKey && e.key === 'Enter') {
    e.preventDefault();
    run();
  }
});
</script>
</body>
</html>`
