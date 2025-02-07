import * as vscode from 'vscode';
import { exec, spawn } from 'child_process';
import { promisify } from 'util';
import * as path from 'path';

const execAsync = promisify(exec);

// Create output channel
const outputChannel = vscode.window.createOutputChannel("retab");

async function checkRetabExists(path: string): Promise<boolean> {
	try {
		await execAsync(`${path} --version`);
		return true;
	} catch {
		return false;
	}
}

function resolveRetabPath(configuredPath: string | undefined): Promise<{ path: string; useGoRun: boolean }> {
	const config = vscode.workspace.getConfiguration('retab');
	const useGoTool = config.get<boolean>('run_as_go_tool');

	if (useGoTool) {
		return Promise.resolve({ path: 'go', useGoRun: false });
	}

	// If no path configured or it's the default "retab", use 'retab' from PATH
	if (!configuredPath || configuredPath === "" || configuredPath === "retab") {
		return checkRetabExists('retab').then(exists => {
			if (exists) {
				return { path: 'retab', useGoRun: false };
			}
			return { path: 'go', useGoRun: true };
		});
	}

	// If it's an absolute path, use it as is
	if (path.isAbsolute(configuredPath)) {
		return checkRetabExists(configuredPath).then(exists => {
			if (exists) {
				return { path: configuredPath, useGoRun: false };
			}
			return { path: 'go', useGoRun: true };
		});
	}

	// Get the workspace folder
	const workspaceFolders = vscode.workspace.workspaceFolders;
	if (!workspaceFolders || workspaceFolders.length === 0) {
		return checkRetabExists(configuredPath).then(exists => {
			if (exists) {
				return { path: configuredPath, useGoRun: false };
			}
			return { path: 'go', useGoRun: true };
		});
	}

	// Resolve relative to the first workspace folder
	const resolvedPath = path.resolve(workspaceFolders[0].uri.fsPath, configuredPath);
	return checkRetabExists(resolvedPath).then(exists => {
		if (exists) {
			return { path: resolvedPath, useGoRun: false };
		}
		return { path: 'go', useGoRun: true };
	});
}

async function formatWithStdin(retabPath: string, useGoRun: boolean, content: string, filePath: string): Promise<string> {
	const config = vscode.workspace.getConfiguration('retab');
	const useGoTool = config.get<boolean>('run_as_go_tool');
	const extensionVersion = vscode.extensions.getExtension('walteh.retab-vscode')?.packageJSON.version || 'latest';

	if (useGoTool) {
		return new Promise((resolve, reject) => {
			const args = ['tool', 'github.com/walteh/retab/v2/cmd/retab', 'fmt', '--stdin', filePath];
			const child = spawn(retabPath, args);
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
					if (stderr.includes('bad tool name')) {
						reject(new Error(
							'The run_as_go_tool option requires:\n' +
							'1. Go 1.24 or higher installed\n' +
							'2. The tool to be present in one of your workspace\'s go.mod files as:\n' +
							'   tool github.com/walteh/retab/v2/cmd/retab\n\n' +
							'Please ensure these requirements are met or disable the run_as_go_tool option.'
						));
					} else {
						reject(new Error(`retab failed with code ${code}: ${stderr}`));
					}
				} else {
					resolve(stdout);
				}
			});

			child.stdin.write(content);
			child.stdin.end();
		});
	}

	return new Promise((resolve, reject) => {
		const args = useGoRun ? 
			['run', `github.com/walteh/retab/v2/cmd/retab@${extensionVersion}`, 'fmt', '--stdin', filePath] :
			['fmt', '--stdin', filePath];

		const child = spawn(retabPath, args);
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

		child.stdin.write(content);
		child.stdin.end();
	});
}

export function activate(context: vscode.ExtensionContext) {
	outputChannel.appendLine("Retab formatter activated");

	// Register formatter for all supported languages
	let disposable = vscode.languages.registerDocumentFormattingEditProvider(
		['proto', 'proto3', 'hcl', 'hcl2', 'terraform', 'tf', 'tfvars', 'dart'],
		{
			async provideDocumentFormattingEdits(document: vscode.TextDocument): Promise<vscode.TextEdit[]> {
				try {
					// Get the configured retab path
					const config = vscode.workspace.getConfiguration('retab');
					const configuredPath = config.get<string>('executable');
					const { path: retabPath, useGoRun } = await resolveRetabPath(configuredPath);

					// Get the document content
					const content = document.getText();

					// Format using stdin
					const formatted = await formatWithStdin(retabPath, useGoRun, content, document.fileName);

					// Return the edit
					return [vscode.TextEdit.replace(
						new vscode.Range(0, 0, document.lineCount, 0),
						formatted
					)];
				} catch (err) {
					// Log error to output channel instead of showing to user
					outputChannel.appendLine(`Error formatting ${document.fileName}: ${err}`);
					return [];
				}
			}
		}
	);

	context.subscriptions.push(disposable);
}

export function deactivate() {
	outputChannel.dispose();
} 