import * as vscode from "vscode";
import { LanguageClient, LanguageClientOptions, ServerOptions } from "vscode-languageclient/node";
import { execSync } from "child_process";

const INSTALL_URL = "https://amterp.github.io/rad/guide/getting-started/#installation";
const INSTALL_HINT = "Install Rad to enable language features: brew install amterp/rad/rad";

function findRadls(): string | null {
    // Check RAD_LSP_PATH environment variable first (for development/custom builds)
    const envPath = process.env.RAD_LSP_PATH;
    if (envPath) {
        return envPath;
    }

    // Check if radls is on PATH
    try {
        const which = process.platform === "win32" ? "where" : "which";
        execSync(`${which} radls`, { stdio: "ignore" });
        return "radls";
    } catch {
        return null;
    }
}

function showInstallError(): void {
    vscode.window.showErrorMessage(
        `Rad Language Server (radls) not found. ${INSTALL_HINT}`,
        "Open Install Guide"
    ).then(selection => {
        if (selection === "Open Install Guide") {
            vscode.env.openExternal(vscode.Uri.parse(INSTALL_URL));
        }
    });
}

export function activate(context: vscode.ExtensionContext) {
    const radlsCommand = findRadls();

    if (!radlsCommand) {
        showInstallError();
        return;
    }

    const serverOptions: ServerOptions = {
        run: { command: radlsCommand },
        debug: { command: radlsCommand }
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: "file", language: "rad" }]
    };

    const client = new LanguageClient("radls", "Rad Language Server", serverOptions, clientOptions);

    client.start().catch((error: Error) => {
        vscode.window.showErrorMessage(`Failed to start Rad Language Server: ${error.message}`);
    });

    context.subscriptions.push(client);
}
