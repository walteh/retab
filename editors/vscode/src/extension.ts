/**
 * Retab VS Code Extension
 *
 * This extension provides formatting capabilities for various file types using different formatting engines.
 *
 * Engine Types:
 * - wasm (default): Uses WebAssembly for fastest performance, runs in the background
 * - go-tool: Uses 'go tool github.com/walteh/retab/v2/cmd/retab' (requires Go 1.24+)
 * - go-run: Uses 'go run github.com/walteh/retab/v2/cmd/retab@version'
 * - path: Uses 'retab' from system PATH
 * - local: Uses specified executable path from 'retab.executable'
 *
 * Configuration:
 * - retab.engine: The formatting engine to use (default: "wasm")
 * - retab.executable: Path to retab executable when using "local" engine
 * - retab.disable_wasm_fallback: Disable fallback to WASM when other engines fail
 * - retab.format_tf_as_hcl: Format Terraform files using HCL formatter
 *
 */

import * as vscode from "vscode";
import { WasmFormatter } from "./wasm";
import { GoRunFormatter, GoToolFormatter, LocalFormatter, PathFormatter } from "./cli";

// Create output channel
const outputChannel = vscode.window.createOutputChannel("retab");

// Engine type enum
export enum RetabEngine {
	WASM = "wasm",
	GO_TOOL = "go-tool",
	GO_RUN = "go-run",
	PATH = "path",
	LOCAL = "local",
}

// Language IDs enum
export enum VSCodeLanguageID {
	PROTO = "proto",
	PROTO3 = "proto3",
	HCL = "hcl",
	HCL2 = "hcl2",
	TERRAFORM = "terraform",
	TF = "tf",
	TFVARS = "tfvars",
	DART = "dart",
	PROTOBUF = "protobuf",
	SWIFT = "swift",
	YAML = "yaml",
	YML = "yml",
}

// Format types enum
export enum RetabFormat {
	PROTO = "proto",
	HCL = "hcl",
	DART = "dart",
	TF = "tf",
	SWIFT = "swift",
	AUTO = "auto",
	YAML = "yaml",
}

// Supported languages array
export const SUPPORTED_LANGUAGES: VSCodeLanguageID[] = [
	VSCodeLanguageID.PROTO,
	VSCodeLanguageID.PROTO3,
	VSCodeLanguageID.HCL,
	VSCodeLanguageID.HCL2,
	VSCodeLanguageID.TERRAFORM,
	VSCodeLanguageID.TF,
	VSCodeLanguageID.TFVARS,
	VSCodeLanguageID.DART,
	VSCodeLanguageID.PROTOBUF,
	VSCodeLanguageID.SWIFT,
	VSCodeLanguageID.YAML,
	VSCodeLanguageID.YML,
];

// Interface for engine implementations
export interface RetabFormatter {
	initialize(extensionContext: vscode.ExtensionContext): Promise<void>;
	format(content: string, filePath: string, formatType: RetabFormat): Promise<string>;
	getVersion(context: vscode.ExtensionContext): Promise<string>;
}

// Map VS Code language IDs to retab format types
function getFormatType(languageId: string, formatTfAsHcl: boolean): RetabFormat {
	// First map the language to its base format type
	const formatTypeMap: { [key in VSCodeLanguageID]: RetabFormat } = {
		[VSCodeLanguageID.PROTO]: RetabFormat.PROTO,
		[VSCodeLanguageID.PROTO3]: RetabFormat.PROTO,
		[VSCodeLanguageID.HCL]: RetabFormat.HCL,
		[VSCodeLanguageID.HCL2]: RetabFormat.HCL,
		[VSCodeLanguageID.TERRAFORM]: RetabFormat.TF,
		[VSCodeLanguageID.TF]: RetabFormat.TF,
		[VSCodeLanguageID.TFVARS]: RetabFormat.TF,
		[VSCodeLanguageID.DART]: RetabFormat.DART,
		[VSCodeLanguageID.PROTOBUF]: RetabFormat.PROTO,
		[VSCodeLanguageID.SWIFT]: RetabFormat.SWIFT,
		[VSCodeLanguageID.YAML]: RetabFormat.YAML,
		[VSCodeLanguageID.YML]: RetabFormat.YAML,
	};

	// Get the base format type
	const baseFormat = formatTypeMap[languageId as VSCodeLanguageID] || RetabFormat.AUTO;

	// Handle special case for Terraform files
	if (formatTfAsHcl && baseFormat === RetabFormat.TF) {
		return RetabFormat.HCL;
	}

	outputChannel.appendLine(`[main] mapped language ${languageId} to format type ${baseFormat}`);
	return baseFormat;
}

// Current engine instance

const wasmFormatter = new WasmFormatter(outputChannel);

function getFormatter(engine: RetabEngine): RetabFormatter {
	switch (engine) {
		case RetabEngine.WASM:
			return wasmFormatter;
		case RetabEngine.GO_TOOL:
			return new GoToolFormatter();
		case RetabEngine.GO_RUN:
			return new GoRunFormatter();
		case RetabEngine.PATH:
			return new PathFormatter();
		case RetabEngine.LOCAL:
			return new LocalFormatter();
		default:
			throw new Error(`unknown engine type: ${engine}`);
	}
}

export function activate(context: vscode.ExtensionContext) {
	outputChannel.appendLine("[main] retab formatter activated");

	// Initialize WASM engine by default
	wasmFormatter.initialize(context).catch((err) => {
		outputChannel.appendLine(`[main] failed to initialize WASM engine: ${err}`);
	});

	let currentEngine: RetabEngine = RetabEngine.WASM;

	let currentFormatter: RetabFormatter = wasmFormatter;

	// Register formatter for all supported languages
	let disposable = vscode.languages.registerDocumentFormattingEditProvider(SUPPORTED_LANGUAGES, {
		async provideDocumentFormattingEdits(document: vscode.TextDocument): Promise<vscode.TextEdit[]> {
			try {
				const config = vscode.workspace.getConfiguration("retab");
				const engine = config.get<RetabEngine>("engine") || RetabEngine.WASM;
				const disableWasmFallback = config.get<boolean>("disable_wasm_fallback") || false;
				const formatTfAsHcl = config.get<boolean>("format_tf_as_hcl") || false;

				// Create formatter instance based on engine type
				if (engine !== currentEngine) {
					currentEngine = engine;
					currentFormatter = getFormatter(engine);
					if (currentFormatter !== wasmFormatter) {
						await currentFormatter.initialize(context);
					}
				}

				const formatType = getFormatType(document.languageId, formatTfAsHcl);

				try {
					const formatted = await currentFormatter.format(document.getText(), document.fileName, formatType);

					return [vscode.TextEdit.replace(new vscode.Range(0, 0, document.lineCount, 0), formatted)];
				} catch (err) {
					outputChannel.appendLine(`[${engine}] error formatting ${document.fileName}: ${err}`);
					// Try WASM fallback if enabled
					if (!disableWasmFallback && engine !== RetabEngine.WASM) {
						outputChannel.appendLine(`[main] falling back to wasm: ${err}`);
						try {
							const formatted = await wasmFormatter.format(
								document.getText(),
								document.fileName,
								formatType
							);
							return [vscode.TextEdit.replace(new vscode.Range(0, 0, document.lineCount, 0), formatted)];
						} catch (err) {
							outputChannel.appendLine(
								`[main] failed to format ${document.fileName} with wasm fallback: ${err}`
							);
							return [];
						}
					}
					outputChannel.appendLine(`[main] wasm fallback disabled, returning empty edits`);
					return [];
				}
			} catch (err) {
				outputChannel.appendLine(`[main] error formatting ${document.fileName}: ${err}`);
				return [];
			}
		},
	});

	context.subscriptions.push(disposable);
}

export function deactivate() {
	outputChannel.dispose();
}
