import * as vscode from 'vscode';

export function activate(context: vscode.ExtensionContext): void {
  const openDocs = vscode.commands.registerCommand('fluffyui.openDocs', async () => {
    const root = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
    if (!root) {
      vscode.window.showInformationMessage('Open docs: no workspace folder found.');
      return;
    }
    const docsPath = vscode.Uri.file(root + '/docs/getting-started.md');
    try {
      const doc = await vscode.workspace.openTextDocument(docsPath);
      await vscode.window.showTextDocument(doc, { preview: true });
    } catch {
      vscode.window.showInformationMessage('Open docs: docs/getting-started.md not found.');
    }
  });

  const runTests = vscode.commands.registerCommand('fluffyui.runTests', () => {
    const terminal = vscode.window.createTerminal('FluffyUI Tests');
    terminal.show(true);
    terminal.sendText('go test ./...');
  });

  context.subscriptions.push(openDocs, runTests);
}

export function deactivate(): void {}
