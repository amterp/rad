import * as path from "path";
import * as vscode from "vscode";
import { LanguageClient, LanguageClientOptions, ServerOptions } from "vscode-languageclient/node";
import * as fs from "fs";

function getLspPath(): string {
    const lspPath = path.join(__dirname, "..", "..", "bin", "rls");

    if (!fs.existsSync(lspPath)) {
        vscode.window.showErrorMessage(`LSP binary not found: ${lspPath}`);
        throw new Error(`LSP binary missing at ${lspPath}`);
    }

    return lspPath;
}

export function activate(context: vscode.ExtensionContext) {
    const serverExecutable = getLspPath();

    const serverOptions: ServerOptions = {
        run: { command: serverExecutable },
        debug: { command: serverExecutable }
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: "file", language: "rad" }]
    };

    const client = new LanguageClient("rls", "Rad Language Server", serverOptions, clientOptions);
    client.start();

    context.subscriptions.push(client);
}