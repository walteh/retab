import * as vscode from 'vscode';
import { exec } from 'child_process';
import { promisify } from 'util';
import * as path from 'path';

const execAsync = promisify(exec);

function resolveRetabPath(configuredPath: string): string {
	// If it's an absolute path, use it as is
	if (path.isAbsolute(configuredPath)) {
		return configuredPath;
	}

	// Get the workspace folder
	const workspaceFolders = vscode.workspace.workspaceFolders;
	if (!workspaceFolders || workspaceFolders.length === 0) {
		return configuredPath; // Fall back to PATH-based resolution
	}

	// Resolve relative to the first workspace folder
	return path.resolve(workspaceFolders[0].uri.fsPath, configuredPath);
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
					const configuredPath = config.get<string>('executable', 'retab');
					const retabPath = resolveRetabPath(configuredPath);

					// Check if retab exists at the configured path
					try {
						await execAsync(`${retabPath} --version`);
					} catch (err) {
						throw new Error(`retab not found at '${retabPath}' - please install retab or configure the correct path in settings`);
					}

					// Run retab formatter
					const { stdout } = await execAsync(`${retabPath} fmt ${document.fileName} --stdout`);

					// Return the edit
					return [vscode.TextEdit.replace(
						new vscode.Range(0, 0, document.lineCount, 0),
						stdout
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