'use strict';

const vscode = require('vscode');
const { LanguageClient } = require('vscode-languageclient/node');

let client;

async function activate() {
    const configuredPath = vscode.workspace.getConfiguration('aiki').get('server.path', 'aiki');
    const serverOptions = {
        command: configuredPath,
        args: ['lsp'],
        options: { env: process.env }
    };
    const clientOptions = {
        documentSelector: [{ scheme: 'file', language: 'aiki' }]
    };

    client = new LanguageClient(
        'aiki',
        'Aiki Language Services',
        serverOptions,
        clientOptions
    );
    await client.start();
}

async function deactivate() {
    if (client) {
        await client.dispose();
        client = undefined;
    }
}

module.exports = { activate, deactivate };
