import * as path from "path";
import * as vscode from "vscode";
import * as fs from "fs";
import { RetabEngine, RetabFormat, RetabFormatter } from "./extension";
import { exec, spawn } from "child_process";
import { promisify } from "util";
import { WasmFormatter } from "./wasm";

const execAsync = promisify(exec);

// Base class for CLI-based formatters
abstract class CliFormatter implements RetabFormatter {
	protected version: string | undefined;

	abstract getCommand(): Promise<{ path: string; args: string[] }>;

	async initialize(context: vscode.ExtensionContext): Promise<void> {
		this.version = await this.getVersion(context);
	}

	async getVersion(context: vscode.ExtensionContext): Promise<string> {
		const { path: cmd, args } = await this.getCommand();
		try {
			const { stdout } = await execAsync(`${cmd} ${args.join(" ")} raw-version`);
			return stdout.trim();
		} catch (err) {
			throw new Error(`Failed to get retab version: ${err}`);
		}
	}

	async format(content: string, filePath: string, formatType: RetabFormat): Promise<string> {
		const { path: cmd, args } = await this.getCommand();
		return new Promise((resolve, reject) => {
			const allArgs = [...args, "fmt", "--stdin", "--format", formatType, filePath];
			const child = spawn(cmd, allArgs);
			let stdout = "";
			let stderr = "";

			child.stdout.on("data", (data) => {
				stdout += data;
			});

			child.stderr.on("data", (data) => {
				stderr += data;
			});

			child.on("error", (err) => {
				reject(new Error(`Failed to spawn retab: ${err.message}`));
			});

			child.on("close", (code) => {
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
}

// Go tool engine implementation
export class GoToolFormatter extends CliFormatter {
	async getCommand(): Promise<{ path: string; args: string[] }> {
		return {
			path: "go",
			args: ["tool", "github.com/walteh/retab/v2/cmd/retab"],
		};
	}
}

// Go run engine implementation
export class GoRunFormatter extends CliFormatter {
	async getCommand(): Promise<{ path: string; args: string[] }> {
		const extension = vscode.extensions.getExtension("walteh.retab-vscode");
		const version = extension?.packageJSON.version || "latest";
		return {
			path: "go",
			args: ["run", `github.com/walteh/retab/v2/cmd/retab@${version}`],
		};
	}
}

// Path engine implementation
export class PathFormatter extends CliFormatter {
	async getCommand(): Promise<{ path: string; args: string[] }> {
		return {
			path: "retab",
			args: [],
		};
	}
}

// Local engine implementation
export class LocalFormatter extends CliFormatter {
	async getCommand(): Promise<{ path: string; args: string[] }> {
		const config = vscode.workspace.getConfiguration("retab");
		const execPath = config.get<string>("executable");
		if (!execPath) {
			throw new Error("retab.executable not configured");
		}
		return {
			path: execPath,
			args: [],
		};
	}
}
