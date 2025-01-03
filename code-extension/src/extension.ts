import * as vscode from 'vscode';
import { exec, spawn } from 'child_process';
import { promisify } from 'util';
import * as path from 'path';

const execAsync = promisify(exec);

function resolveRetabPath(configuredPath: string | undefined): string {
	// If no path configured, use 'retab' from PATH
	if (!configuredPath) {
		return 'retab';
	}

	// If it's an absolute path, use it as is
	if (path.isAbsolute(configuredPath)) {
		return configuredPath;
	}

	// Get the workspace folder
	const workspaceFolders = vscode.workspace.workspaceFolders;
	if (!workspaceFolders || workspaceFolders.length === 0) {
		return configuredPath;
	}

	// Resolve relative to the first workspace folder
	return path.resolve(workspaceFolders[0].uri.fsPath, configuredPath);
}

async function formatWithStdin(retabPath: string, content: string, filePath: string): Promise<string> {
	return new Promise((resolve, reject) => {
		const child = spawn(retabPath, ['fmt', '--stdin', filePath]);
		let stdout = '';
		let stderr = '';

		child.stdout.on('data', (data) => {
			stdout += data;
		});

		child.stderr.on('data', (data) => {
			stderr += data;
		});

		child.on('error', (err) => {
			reject(new Error(`Failed to spawn retab: ${err.message}`));
		});

		child.on('close', (code) => {
			if (code !== 0) {
				reject(new Error(`retab failed with code ${code}: ${stderr}`));
			} else {
				resolve(stdout);
			}
		});

		// Write the content to stdin and close it
		child.stdin.write(content);
		child.stdin.end();
	});
}

export function activate(context: vscode.ExtensionContext) {
	// Register formatter for all supported languages
	let disposable = vscode.languages.registerDocumentFormattingEditProvider(
		['protobuf', 'hcl', 'terraform', 'dart'],
		{
			async provideDocumentFormattingEdits(document: vscode.TextDocument): Promise<vscode.TextEdit[]> {
				try {
					// Get the configured retab path
					const config = vscode.workspace.getConfiguration('retab');
					const configuredPath = config.get<string>('executable');
					const retabPath = resolveRetabPath(configuredPath);

					// Get the document content
					const content = document.getText();

					// Format using stdin
					const formatted = await formatWithStdin(retabPath, content, document.fileName);

					// Return the edit
					return [vscode.TextEdit.replace(
						new vscode.Range(0, 0, document.lineCount, 0),
						formatted
					)];
				} catch (err) {
					vscode.window.showErrorMessage(`Retab formatting failed: ${err}`);
					return [];
				}
			}
		}
	);

	context.subscriptions.push(disposable);
}

export function deactivate() {} 