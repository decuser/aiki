"""Thin Xed client for ``aiki lsp``.

The plugin owns transport and presentation only. It contains no Aiki syntax,
scope, formatting, or diagnostic logic.
"""

import json
import os
import subprocess
import threading

from gi.repository import GLib, GObject, Pango, Xed


class LSPClient:
    def __init__(self, on_ready, on_diagnostics):
        self._on_ready = on_ready
        self._on_diagnostics = on_diagnostics
        self._lock = threading.Lock()
        self._next_id = 2
        self._running = True
        self._process = subprocess.Popen(
            ["aiki", "lsp"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            bufsize=0,
        )
        self._reader = threading.Thread(target=self._read_loop, daemon=True)
        self._reader.start()
        self._send_request(
            1,
            "initialize",
            {
                "processId": os.getpid(),
                "capabilities": {"general": {"positionEncodings": ["utf-16"]}},
                "clientInfo": {"name": "xed-aiki"},
            },
        )

    def close(self):
        if not self._running:
            return
        self._running = False
        try:
            request_id = self._next_request_id()
            self._send_request(request_id, "shutdown", None)
            self._send_notification("exit", None)
        except (BrokenPipeError, OSError):
            pass
        try:
            self._process.terminate()
        except OSError:
            pass

    def did_open(self, uri, text, version):
        self._send_notification(
            "textDocument/didOpen",
            {
                "textDocument": {
                    "uri": uri,
                    "languageId": "aiki",
                    "version": version,
                    "text": text,
                }
            },
        )

    def did_change(self, uri, text, version):
        self._send_notification(
            "textDocument/didChange",
            {
                "textDocument": {"uri": uri, "version": version},
                "contentChanges": [{"text": text}],
            },
        )

    def did_close(self, uri):
        self._send_notification(
            "textDocument/didClose", {"textDocument": {"uri": uri}}
        )

    def _next_request_id(self):
        request_id = self._next_id
        self._next_id += 1
        return request_id

    def _send_request(self, request_id, method, params):
        self._send({"jsonrpc": "2.0", "id": request_id, "method": method, "params": params})

    def _send_notification(self, method, params):
        self._send({"jsonrpc": "2.0", "method": method, "params": params})

    def _send(self, message):
        if not self._running:
            return
        body = json.dumps(message, separators=(",", ":")).encode("utf-8")
        frame = b"Content-Length: %d\r\n\r\n" % len(body) + body
        with self._lock:
            self._process.stdin.write(frame)
            self._process.stdin.flush()

    def _read_loop(self):
        stream = self._process.stdout
        while self._running:
            headers = {}
            while True:
                line = stream.readline()
                if not line:
                    return
                if line in (b"\r\n", b"\n"):
                    break
                name, _, value = line.decode("ascii").partition(":")
                headers[name.lower()] = value.strip()
            length = int(headers.get("content-length", "0"))
            if length <= 0:
                continue
            body = stream.read(length)
            if len(body) != length:
                return
            message = json.loads(body.decode("utf-8"))
            if message.get("id") == 1 and "result" in message:
                self._send_notification("initialized", {})
                GLib.idle_add(self._on_ready)
            elif message.get("method") == "textDocument/publishDiagnostics":
                params = message.get("params", {})
                GLib.idle_add(
                    self._on_diagnostics,
                    params.get("uri", ""),
                    params.get("diagnostics", []),
                )


class AikiLanguageServicesPlugin(GObject.Object, Xed.WindowActivatable):
    __gtype_name__ = "AikiLanguageServicesPlugin"

    window = GObject.Property(type=Xed.Window)

    def __init__(self):
        GObject.Object.__init__(self)
        self._client = None
        self._ready = False
        self._documents = {}
        self._diagnostics = {}

    def do_activate(self):
        self._client = LSPClient(self._on_ready, self._on_diagnostics)
        self._sync_documents()

    def do_deactivate(self):
        for document in list(self._documents):
            self._detach(document)
        if self._client is not None:
            self._client.close()
            self._client = None

    def do_update_state(self):
        # WindowActivatable update_state is called as tabs/documents change;
        # use it as the reconciliation point instead of depending on private
        # Xed tab signals.
        self._sync_documents()

    def _on_ready(self):
        self._ready = True
        for document in list(self._documents):
            self._open(document)
        return False

    def _sync_documents(self):
        current = set(self.window.get_documents())
        known = set(self._documents)
        for document in current - known:
            self._attach(document)
        for document in current & known:
            self._reconcile_document(document)
        for document in known - current:
            self._detach(document)

    def _attach(self, document):
        state = {"version": 1, "timeout": 0, "opened_uri": ""}
        state["changed"] = document.connect("changed", self._on_changed)
        self._documents[document] = state
        self._ensure_tags(document)
        if self._ready:
            self._open(document)

    def _detach(self, document):
        state = self._documents.pop(document, None)
        if state is None:
            return
        if state.get("timeout"):
            GLib.source_remove(state["timeout"])
        try:
            document.disconnect(state["changed"])
        except (TypeError, RuntimeError):
            pass
        opened_uri = state.get("opened_uri", "")
        if self._ready and opened_uri:
            self._client.did_close(opened_uri)
        self._clear_tags(document)

    def _on_changed(self, document):
        state = self._documents.get(document)
        if state is None:
            return
        state["version"] += 1
        if state["timeout"]:
            GLib.source_remove(state["timeout"])
        state["timeout"] = GLib.timeout_add(150, self._flush_change, document)

    def _flush_change(self, document):
        state = self._documents.get(document)
        if state is None:
            return False
        state["timeout"] = 0
        if not self._ready:
            return False
        self._reconcile_document(document)
        opened_uri = state.get("opened_uri", "")
        if opened_uri:
            self._client.did_change(
                opened_uri, self._text(document), state["version"]
            )
        return False

    def _reconcile_document(self, document):
        if not self._ready:
            return
        state = self._documents.get(document)
        if state is None:
            return
        current_uri = self._uri(document) if self._is_aiki(document) else ""
        opened_uri = state.get("opened_uri", "")
        if opened_uri == current_uri:
            return
        if opened_uri:
            self._client.did_close(opened_uri)
            self._diagnostics.pop(opened_uri, None)
            self._clear_tags(document)
            state["opened_uri"] = ""
        if current_uri:
            self._client.did_open(
                current_uri, self._text(document), state["version"]
            )
            state["opened_uri"] = current_uri

    def _open(self, document):
        self._reconcile_document(document)

    def _on_diagnostics(self, uri, diagnostics):
        self._diagnostics[uri] = diagnostics
        for document in self._documents:
            if self._uri(document) == uri:
                self._apply_diagnostics(document, diagnostics)
                break
        return False

    def _apply_diagnostics(self, document, diagnostics):
        self._clear_tags(document)
        error_tag, warning_tag = self._ensure_tags(document)
        for diagnostic in diagnostics:
            start = diagnostic.get("range", {}).get("start", {})
            end = diagnostic.get("range", {}).get("end", start)
            start_iter = self._iter_at_lsp(document, start)
            end_iter = self._iter_at_lsp(document, end)
            start_iter, end_iter = self._visible_range(
                document, start_iter, end_iter
            )
            tag = warning_tag if diagnostic.get("severity") == 2 else error_tag
            document.apply_tag(tag, start_iter, end_iter)

    @staticmethod
    def _visible_range(document, start_iter, end_iter):
        # LSP permits zero-width diagnostics (for example an error at EOF).
        # GtkTextBuffer tags need a non-empty range to render an underline.
        if start_iter.get_offset() != end_iter.get_offset():
            return start_iter, end_iter

        forward = end_iter.copy()
        if forward.forward_char():
            return start_iter, forward

        backward = start_iter.copy()
        if backward.backward_char():
            return backward, end_iter

        return start_iter, end_iter

    def _ensure_tags(self, document):
        table = document.get_tag_table()
        error_tag = table.lookup("aiki-lsp-error")
        warning_tag = table.lookup("aiki-lsp-warning")
        if error_tag is None:
            error_tag = document.create_tag(
                "aiki-lsp-error", underline=Pango.Underline.ERROR
            )
        if warning_tag is None:
            warning_tag = document.create_tag(
                "aiki-lsp-warning", underline=Pango.Underline.SINGLE
            )
        return error_tag, warning_tag

    def _clear_tags(self, document):
        table = document.get_tag_table()
        start, end = document.get_bounds()
        for name in ("aiki-lsp-error", "aiki-lsp-warning"):
            tag = table.lookup(name)
            if tag is not None:
                document.remove_tag(tag, start, end)

    @staticmethod
    def _uri(document):
        location = document.get_location()
        return location.get_uri() if location is not None else ""

    @staticmethod
    def _is_aiki(document):
        location = document.get_location()
        return location is not None and location.get_basename().endswith(".ai")

    @staticmethod
    def _text(document):
        start, end = document.get_bounds()
        return document.get_text(start, end, True)

    @staticmethod
    def _iter_at_lsp(document, position):
        line = max(0, int(position.get("line", 0)))
        utf16_offset = max(0, int(position.get("character", 0)))
        line_start = document.get_iter_at_line(line)
        line_end = line_start.copy()
        line_end.forward_to_line_end()
        text = document.get_text(line_start, line_end, True)
        chars = 0
        units = 0
        for char in text:
            width = 2 if ord(char) > 0xFFFF else 1
            if units + width > utf16_offset:
                break
            units += width
            chars += 1
        result = line_start.copy()
        result.forward_chars(chars)
        return result
