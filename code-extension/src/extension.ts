import * as vscode from 'vscode';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export function activate(context: vscode.ExtensionContext) {
	// Register formatter for all supported languages
	let disposable = vscode.languages.registerDocumentFormattingEditProvider(
		['protobuf', 'hcl', 'terraform', 'dart'],
		{
			async provideDocumentFormattingEdits(document: vscode.TextDocument): Promise<vscode.TextEdit[]> {
				try {
                    // make sure retab is in the PATH
                    const retabPath = await execAsync('which retab');
                    if (retabPath.stderr) {
                        throw new Error('retab not found in PATH - please install retab at https://github.com/walteh/retab');
                    }

					// Run retab formatter
					const { stdout } = await execAsync(`retab fmt ${document.fileName} --stdout`);

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