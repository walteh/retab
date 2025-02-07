import * as path from 'path';
import * as vscode from 'vscode';
import * as fs from 'fs';

declare global {
    interface Window {
        Go: any;
        retab_run: (formatter: string, filename: string, editorConfigContent: string, fileContent: string) => boolean;
        retab_run_result_error: string;
        retab_run_result_success: string;
    }
    var retab_run: (formatter: string, filename: string, editorConfigContent: string, fileContent: string) => boolean;
    var retab_run_result_error: string;
    var retab_run_result_success: string;
}

let go: any = null;
const outputChannel = vscode.window.createOutputChannel("retab-wasm");

export async function initWasm(context: vscode.ExtensionContext): Promise<void> {
    outputChannel.appendLine("Initializing WASM module...");
    try {
        // Load and execute wasm_exec.js
        const wasmExecPath = path.join(context.extensionPath, 'out', 'wasm_exec.js');
        outputChannel.appendLine(`Loading wasm_exec.js from ${wasmExecPath}`);
        const wasmExecContent = await fs.promises.readFile(wasmExecPath, 'utf8');
        
        // Create a new context for the WASM execution
        outputChannel.appendLine("Creating Go runtime...");
        go = new (Function(`
            ${wasmExecContent}
            return Go;
        `)())();

        // Load and instantiate the WASM module
        const wasmPath = path.join(context.extensionPath, 'out', 'retab.wasm');
        outputChannel.appendLine(`Loading WASM module from ${wasmPath}`);
        const wasmBuffer = await vscode.workspace.fs.readFile(vscode.Uri.file(wasmPath));
        outputChannel.appendLine(`WASM module loaded, size: ${wasmBuffer.length} bytes`);
        
        const wasmModule = await WebAssembly.compile(wasmBuffer);
        outputChannel.appendLine("WASM module compiled");
        
        const instance = await WebAssembly.instantiate(wasmModule, go.importObject);
        outputChannel.appendLine("WASM module instantiated");
        
        go.run(instance);
        outputChannel.appendLine("WASM module initialized and running");
    } catch (err) {
        outputChannel.appendLine(`Error initializing WASM: ${err}`);
        throw err;
    }
}

export async function formatWithWasm(content: string, filePath: string, formatType: string, editorConfig: string): Promise<string> {
    outputChannel.appendLine(`\nFormatting with WASM: ${filePath} (${formatType})`);
    outputChannel.appendLine(`Content length: ${content.length}, EditorConfig length: ${editorConfig.length}`);

    if (!go || !(globalThis as any).retab_run) {
        const error = 'WASM module not initialized or retab_run not available';
        outputChannel.appendLine(error);
        throw new Error(error);
    }

    // Reset global variables
    (globalThis as any).retab_run_result_error = undefined;
    (globalThis as any).retab_run_result_success = undefined;
    
    try {
        outputChannel.appendLine("Calling retab_run...");
        const success = (globalThis as any).retab_run(formatType, filePath, editorConfig, content);
        outputChannel.appendLine(`retab_run result: ${success}`);
        
        if (!success) {
            const error = (globalThis as any).retab_run_result_error || 'WASM formatting failed';
            outputChannel.appendLine(error);
            throw new Error(error);
        }

        const result = (globalThis as any).retab_run_result_success;
        if (!result) {
            throw new Error('No result returned from WASM');
        }

        outputChannel.appendLine(`Formatting complete. Result length: ${result.length}`);
        return result;
    } catch (err) {
        const error = `WASM formatting error: ${err}`;
        outputChannel.appendLine(error);
        throw new Error(error);
    }
} 