import * as path from "path";
import * as vscode from "vscode";
import * as fs from "fs";
import { RetabEngine, RetabFormat, RetabFormatter } from "./extension";
import * as editorconfig from "editorconfig";
import { Visited } from "editorconfig";

declare global {
	interface result {
		result: string;
		error: string | undefined;
	}
	interface retab {
		fmt: (formatter: RetabFormat, filename: string, fileContent: string, editorConfigContent: string) => result;
	}

	interface Go {
		importObject: WebAssembly.Imports & {
			gojs: {
				"syscall/js.finalizeRef": (v_ref: any) => void;
			};
		};
		run: (instance: any) => void;
	}

	var retab_initialized: boolean;
	var retab: retab;
}

export class WasmFormatter implements RetabFormatter {
	private go: Go | null = null;
	private initialized = false;
	private outputChannel: vscode.OutputChannel;

	constructor(outputChannel: vscode.OutputChannel) {
		this.outputChannel = outputChannel;
	}

	private log(message: string) {
		this.outputChannel.appendLine(`[${RetabEngine.WASM}] ${message}`);
	}

	private async waitForInit(timeout: number = 5000): Promise<void> {
		this.log("Waiting for WASM initialization...");
		const start = Date.now();
		while (!globalThis.retab_initialized) {
			if (Date.now() - start > timeout) {
				throw new Error("Timeout waiting for WASM initialization");
			}
			await new Promise((resolve) => setTimeout(resolve, 100));
		}
		this.log("WASM initialization complete");
	}

	async initialize(context: vscode.ExtensionContext): Promise<void> {
		if (this.initialized) {
			return;
		}

		this.log("Initializing WASM module...");
		try {
			// Load and execute wasm_exec.js
			let wasmExecPath = path.join(context.extensionPath, "out", "wasm_exec.js");

			const wasmExecPathTinygo = path.join(context.extensionPath, "out", "wasm_exec.tinygo.js");

			let useTinygo = false;

			// check if wasm_exec.golang.js exists
			if (fs.existsSync(wasmExecPathTinygo)) {
				wasmExecPath = wasmExecPathTinygo;
				useTinygo = true;
			}

			const wasmExecContent = await vscode.workspace.fs.readFile(vscode.Uri.file(wasmExecPath));
			let wasmExecContentString = wasmExecContent.toString();

			if (useTinygo) {
				// prevents an error when .String() is called, but does not fully solve the memory leak issue
				// - however, the memory leak is not that bad - https://github.com/tinygo-org/tinygo/issues/1140#issuecomment-1314608377
				wasmExecContentString = wasmExecContentString.replace(
					'"syscall/js.finalizeRef":',
					`"syscall/js.finalizeRef": (v_ref) => {
				 	const id = mem().getUint32(unboxValue(v_ref), true);
				 	this._goRefCounts[id]--;
				 	if (this._goRefCounts[id] === 0) {
				 		const v = this._values[id];
				 		this._values[id] = null;
				 		this._ids.delete(v);
				 		this._idPool.push(id);
				 	}
				 },
				 "syscall/js.finalizeRef-tinygo":`
				);
			}

			// Create a new context for the WASM execution
			this.log("Creating Go runtime...");
			this.go = new (Function(`
				${wasmExecContentString}
				return Go;
			`)())();

			if (!this.go) {
				throw new Error("Failed to create Go runtime");
			}

			// Load and instantiate the WASM module
			const wasmPath = path.join(context.extensionPath, "out", "retab.wasm");
			this.log(`Loading WASM module from ${wasmPath}`);

			const wasmBuffer = await vscode.workspace.fs.readFile(vscode.Uri.file(wasmPath));
			this.log(`WASM module loaded, size: ${wasmBuffer.length} bytes`);

			const wasmModule = await WebAssembly.compile(wasmBuffer);
			this.log("WASM module compiled");

			const instance = await WebAssembly.instantiate(wasmModule, this.go.importObject);
			this.log("WASM module instantiated");

			this.go.run(instance);
			this.log("WASM module started");

			// Wait for initialization to complete
			await this.waitForInit();
			this.log("WASM module fully initialized");
			this.initialized = true;
		} catch (err) {
			this.log(`Error initializing WASM: ${err}`);
			throw err;
		}
	}

	async format(content: string, filePath: string, formatType: RetabFormat): Promise<string> {
		this.log(`Formatting ${filePath} (${formatType})`);
		this.log(`Content length: ${content.length}`);

		// Ensure WASM is initialized
		if (!this.initialized || !globalThis.retab_initialized) {
			const error = "WASM module not fully initialized";
			this.log(error);
			throw new Error(error);
		}

		if (!this.go || !globalThis.retab?.fmt) {
			const error = "WASM module not properly initialized or retab.fmt not available";
			this.log(error);
			throw new Error(error);
		}

		try {
			// Use editorconfig-core-js to parse the file
			this.log(`Looking for editorconfig settings for ${filePath}`);
			let files: Visited[] = [];
			await editorconfig.parse(filePath, { files: files });
			this.log(`Found editorconfig settings: ${JSON.stringify(files)}`);

			let editorconfigContent = "";
			if (files.length > 0) {
				const content = await vscode.workspace.fs.readFile(vscode.Uri.file(files[0].fileName));
				editorconfigContent = content.toString();
			} else {
				this.log("No editorconfig settings found");
			}

			this.log("Calling retab.fmt...");
			const response = globalThis.retab.fmt(formatType, filePath, content, editorconfigContent);
			this.log("retab.fmt response received");

			if (!response) {
				this.log("retab.fmt response is undefined");
				throw new Error("retab.fmt response is undefined");
			}

			if (response.error) {
				this.log(`Error: ${response.error}`);
				throw new Error(response.error);
			}

			this.log(`Formatting complete. Result length: ${(response.result as string).length}`);
			return response.result as string;
		} catch (err) {
			const error = `WASM formatting error: ${err}`;
			this.log(error);
			throw new Error(error);
		}
	}

	async getVersion(context: vscode.ExtensionContext): Promise<string> {
		// WASM version is tied to extension version
		const extension = vscode.extensions.getExtension("walteh.retab-vscode");
		return extension?.packageJSON.version || "unknown";
	}
}
