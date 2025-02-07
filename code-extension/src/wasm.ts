import * as path from 'path';
import * as vscode from 'vscode';
import * as fs from 'fs';

declare global {

	interface result {
		result: string;
		error: string | undefined;
	}
	interface retab {
		fmt: (formatter: string, filename: string, fileContent: string, editorConfigContent: string) => result;
	}


    var retab_initialized: boolean;
	var retab: retab;
}

let go: any = null;
const outputChannel = vscode.window.createOutputChannel("retab-wasm");

// Wait for WASM initialization
async function waitForWasmInit(timeout: number = 5000): Promise<void> {
    const start = Date.now();
    while (!globalThis.retab_initialized) {
        if (Date.now() - start > timeout) {
            throw new Error('Timeout waiting for WASM initialization');
        }
        await new Promise(resolve => setTimeout(resolve, 100));
    }
}

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
        outputChannel.appendLine("WASM module started");
        
        // Wait for initialization to complete
        await waitForWasmInit();
        outputChannel.appendLine("WASM module fully initialized");
    } catch (err) {
        outputChannel.appendLine(`Error initializing WASM: ${err}`);
        throw err;
    }
}

export async function formatWithWasm(content: string, filePath: string, formatType: string, editorConfig: string): Promise<string> {
    outputChannel.appendLine(`\nFormatting with WASM: ${filePath} (${formatType})`);
    outputChannel.appendLine(`Content length: ${content.length}, EditorConfig length: ${editorConfig.length}`);
	
    // Ensure WASM is initialized
    if (!globalThis.retab_initialized) {
        const error = 'WASM module not fully initialized';
        outputChannel.appendLine(error);
        throw new Error(error);
    }

    if (!go || !globalThis.retab?.fmt) {
        const error = 'WASM module not properly initialized or retab.fmt not available';
        outputChannel.appendLine(error);
        throw new Error(error);
    }
    
    try {
        outputChannel.appendLine("Calling retab.fmt...");
        const response = globalThis.retab.fmt(formatType, filePath, content, editorConfig );
		outputChannel.appendLine(`retab.fmt response: ${response}`);

		if (response.error) {
			outputChannel.appendLine(response.error);
			throw new Error(response.error);
		}
    

        outputChannel.appendLine(`Formatting complete. Result length: ${(response.result as string).length}`);
        return response.result as string;
    } catch (err) {
        const error = `WASM formatting error: ${err}`;
        outputChannel.appendLine(error);
        throw new Error(error);
    }
}
