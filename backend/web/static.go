//go:build !js

package web

import (
	"fmt"
	"strings"
)

// indexHTML generates the terminal HTML page.
func indexHTML(title, wsPath string, chromeless bool) []byte {
	headerCSS := ""
	footerCSS := ""
	containerCSS := `padding: 8px;`
	headerHTML := fmt.Sprintf(`<div id="header">
        <div>%s</div>
        <div id="status">
            <span id="status-text">Connecting...</span>
            <div id="status-indicator"></div>
        </div>
    </div>`, title)
	footerHTML := `<div id="footer">
        <span>FluffyUI Web Terminal</span>
        <span id="dimensions">80x24</span>
    </div>`

	if chromeless {
		headerCSS = `#header { display: none !important; }`
		footerCSS = `#footer { display: none !important; }`
		containerCSS = `padding: 0;`
		headerHTML = `<div id="header" style="display:none"></div>`
		footerHTML = `<div id="footer" style="display:none"><span id="dimensions"></span></div>`
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            background: #1e1e1e;
            height: 100vh;
            display: flex;
            flex-direction: column;
        }
        #header {
            background: #2d2d2d;
            color: #cccccc;
            padding: 8px 16px;
            font-family: system-ui, -apple-system, sans-serif;
            font-size: 14px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            border-bottom: 1px solid #3e3e3e;
        }
        #status {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        #status-indicator {
            width: 8px;
            height: 8px;
            border-radius: 50%%%%;
            background: #858585;
        }
        #status-indicator.connected {
            background: #4ec9b0;
        }
        #status-indicator.disconnected {
            background: #f48771;
        }
        #terminal-container {
            flex: 1;
            %s
            overflow: hidden;
        }
        .xterm {
            height: 100%%%%;
        }
        #footer {
            background: #2d2d2d;
            color: #858585;
            padding: 4px 16px;
            font-family: system-ui, -apple-system, sans-serif;
            font-size: 12px;
            border-top: 1px solid #3e3e3e;
            display: flex;
            justify-content: space-between;
        }
        %s
        %s
    </style>
</head>
<body>
    %s
    <div id="terminal-container"></div>
    %s

    <script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/xterm-addon-web-links@0.9.0/lib/xterm-addon-web-links.min.js"></script>
    <script>
        (function() {
            const wsPath = '%s';
            const statusText = document.getElementById('status-text');
            const statusIndicator = document.getElementById('status-indicator');
            const dimensions = document.getElementById('dimensions');

            // Initialize terminal
            const term = new Terminal({
                cursorBlink: true,
                cursorStyle: 'block',
                fontFamily: 'Consolas, "Courier New", monospace',
                fontSize: 14,
                theme: {
                    background: '#1e1e1e',
                    foreground: '#cccccc',
                    cursor: '#cccccc',
                    selectionBackground: '#264f78',
                    black: '#000000',
                    red: '#cd3131',
                    green: '#0dbc79',
                    yellow: '#e5e510',
                    blue: '#2472c8',
                    magenta: '#bc3fbc',
                    cyan: '#11a8cd',
                    white: '#e5e5e5',
                    brightBlack: '#666666',
                    brightRed: '#f14c4c',
                    brightGreen: '#23d18b',
                    brightYellow: '#f5f543',
                    brightBlue: '#3b8eea',
                    brightMagenta: '#d670d6',
                    brightCyan: '#29b8db',
                    brightWhite: '#e5e5e5'
                },
                allowProposedApi: true
            });

            // Add addons
            const fitAddon = new FitAddon.FitAddon();
            term.loadAddon(fitAddon);
            term.loadAddon(new WebLinksAddon.WebLinksAddon());

            // Open terminal
            term.open(document.getElementById('terminal-container'));
            fitAddon.fit();

            // WebSocket setup
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + wsPath;
            let ws = null;
            let reconnectAttempts = 0;
            const maxReconnectAttempts = 5;

            function connect() {
                ws = new WebSocket(wsUrl);
                ws.binaryType = 'arraybuffer';

                ws.onopen = function() {
                    if (statusText) statusText.textContent = 'Connected';
                    if (statusIndicator) statusIndicator.className = 'connected';
                    reconnectAttempts = 0;

                    // Send terminal dimensions
                    const cols = term.cols;
                    const rows = term.rows;
                    if (dimensions) dimensions.textContent = cols + 'x' + rows;
                };

                ws.onmessage = function(event) {
                    if (event.data instanceof ArrayBuffer) {
                        // Binary data: write to terminal
                        const data = new Uint8Array(event.data);
                        term.write(data);
                    } else {
                        // Text data: JSON control messages
                        try {
                            const msg = JSON.parse(event.data);
                            handleControlMessage(msg);
                        } catch (e) {
                            console.error('Invalid JSON:', event.data);
                        }
                    }
                };

                ws.onclose = function() {
                    if (statusText) statusText.textContent = 'Disconnected';
                    if (statusIndicator) statusIndicator.className = 'disconnected';

                    // Attempt reconnection
                    if (reconnectAttempts < maxReconnectAttempts) {
                        reconnectAttempts++;
                        if (statusText) statusText.textContent = 'Reconnecting... (' + reconnectAttempts + '/' + maxReconnectAttempts + ')';
                        setTimeout(connect, 1000 * reconnectAttempts);
                    }
                };

                ws.onerror = function(error) {
                    console.error('WebSocket error:', error);
                    if (statusText) statusText.textContent = 'Error';
                    if (statusIndicator) statusIndicator.className = 'disconnected';
                };
            }

            function handleControlMessage(msg) {
                switch (msg.type) {
                    case 'error':
                        term.writeln('\r\n\x1b[31mError: ' + (msg.message || 'Unknown error') + '\x1b[0m');
                        break;
                    case 'resize':
                        if (msg.width && msg.height) {
                            term.resize(msg.width, msg.height);
                            if (dimensions) dimensions.textContent = msg.width + 'x' + msg.height;
                        }
                        break;
                }
            }

            // Forward terminal input to WebSocket
            term.onData(function(data) {
                if (ws && ws.readyState === WebSocket.OPEN) {
                    // Convert string to Uint8Array
                    const encoder = new TextEncoder();
                    ws.send(encoder.encode(data));
                }
            });

            // Forward mouse wheel as scroll keys
            document.getElementById('terminal-container').addEventListener('wheel', function(e) {
                if (ws && ws.readyState === WebSocket.OPEN) {
                    e.preventDefault();
                    const encoder = new TextEncoder();
                    const lines = Math.max(1, Math.abs(Math.round(e.deltaY / 40)));
                    const key = e.deltaY < 0 ? '\x1b[A' : '\x1b[B'; // Up or Down arrow
                    for (let i = 0; i < lines; i++) {
                        ws.send(encoder.encode(key));
                    }
                }
            }, { passive: false });

            // Handle resize
            window.addEventListener('resize', function() {
                fitAddon.fit();
                const cols = term.cols;
                const rows = term.rows;
                if (dimensions) dimensions.textContent = cols + 'x' + rows;

                // Send resize message
                if (ws && ws.readyState === WebSocket.OPEN) {
                    ws.send(JSON.stringify({
                        type: 'resize',
                        width: cols,
                        height: rows
                    }));
                }
            });

            // Initial connection
            connect();

            // Cleanup on page unload
            window.addEventListener('beforeunload', function() {
                if (ws) {
                    ws.close();
                }
            });
        })();
    </script>
</body>
</html>
`, title, containerCSS, headerCSS, footerCSS, headerHTML, footerHTML, wsPath))

	return []byte(b.String())
}
