import * as path from "path";
import * as vscode from "vscode";
import * as fs from "fs";
import { RetabEngine, RetabFormat, RetabFormatter } from "./extension";
import * as editorconfig from "editorconfig";
import { Visited } from "editorconfig";
import { exec, execSync } from "child_process";
import { tmpdir } from "os";
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
	var wasm_log: (message: string) => void;

	var retab_exec: (cmd: string, data: string, tempFiles: string) => string;
}

type Success<T> = {
	data: T;
	error: null;
};

type Failure<E> = {
	data: null;
	error: E;
};

type Result<T, E = Error> = Success<T> | Failure<E>;

export async function tryCatch<T, E = Error>(promise: Promise<T>): Promise<Result<T, E>> {
	try {
		const data = await promise;
		return { data, error: null };
	} catch (error) {
		return { data: null, error: error as E };
	}
}

export class WasmFormatter implements RetabFormatter {
	private go: Go | null = null;
	private initialized = false;
	private outputChannel: vscode.OutputChannel;

	constructor(outputChannel: vscode.OutputChannel) {
		this.outputChannel = outputChannel;
	}

	private log(message: string) {
		this.outputChannel.appendLine(`[engine:${RetabEngine.WASM}] ${message}`);
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

			globalThis.wasm_log = (message: string) => {
				this.outputChannel.appendLine(`[wasm:console.log] ${message}`);
			};

			globalThis.retab_exec = (cmd: string, data: string, tempFiles: string): string => {
				const tempFilesJson = JSON.parse(tempFiles);
				const tempFilesMap = new Map(Object.entries(tempFilesJson));

				const tmpDir = fs.mkdtempSync(path.join(tmpdir(), "retab-"));

				this.log(`Created temp directory: ${tmpDir}`);

				try {
					// Create the temporary directory

					for (const [key, value] of tempFilesMap.entries()) {
						const filePath = path.join(tmpDir, key);
						fs.writeFileSync(filePath, value as string);
						this.log(`Created temp file: ${filePath}`);
					}

					this.log(`Executing command: ${cmd}`);

					// Execute the modified command
					const result = execSync(cmd, {
						input: data,
						cwd: tmpDir,
					});

					this.log(`Command executed, result: ${result.toString()}`);
					return result.toString();
				} catch (error) {
					this.log(`Error executing command: ${error}`);
					throw error;
				} finally {
					// Clean up temporary files
					try {
						fs.rmSync(tmpDir, { recursive: true });

						this.log(`Cleaned up temp directory: ${tmpDir}`);
					} catch (cleanupError) {
						this.log(`Warning: Failed to clean up temp directory: ${cleanupError}`);
					}
				}
			};

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
			const { data: editorconfigContent, error } = await tryCatch(this.parseEditorconfig(filePath));
			if (error) {
				throw error;
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

	private async parseEditorconfig(filePath: string): Promise<string> {
		let files: Visited[] = [];
		let editorconfigContent = "";
		this.log(`Looking for editorconfig settings for ${filePath}`);

		const result = await tryCatch(editorconfig.parse(filePath, { files: files }));
		if (result.error) {
			this.log(`Error parsing editorconfig: ${result.error}`);
			return "";
		}

		this.log(`Found editorconfig settings: ${JSON.stringify(files)}`);

		if (files.length > 0) {
			const content = await vscode.workspace.fs.readFile(vscode.Uri.file(files[0].fileName));
			editorconfigContent = content.toString();
		} else {
			this.log("No editorconfig settings found");
		}

		return editorconfigContent;
	}
}
