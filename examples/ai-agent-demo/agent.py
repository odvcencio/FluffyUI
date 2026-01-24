#!/usr/bin/env python3
import json
import os
import socket
import sys
import time


def parse_addr(raw):
    if raw.startswith("tcp:"):
        raw = raw[4:]
    if raw.startswith("unix:"):
        raise ValueError("unix sockets are not supported in this demo")
    if ":" not in raw:
        raise ValueError("expected host:port")
    host, port = raw.rsplit(":", 1)
    return host, int(port)


def send(sock_file, msg):
    payload = json.dumps(msg).encode("utf-8") + b"\n"
    sock_file.write(payload)
    sock_file.flush()
    line = sock_file.readline()
    if not line:
        raise RuntimeError("connection closed")
    return json.loads(line.decode("utf-8"))


def main():
    raw_addr = os.environ.get("FLUFFYUI_AGENT_ADDR", "tcp:127.0.0.1:7777")
    try:
        host, port = parse_addr(raw_addr)
    except ValueError as exc:
        print("invalid address:", exc, file=sys.stderr)
        return 1

    sock = socket.create_connection((host, port))
    with sock:
        f = sock.makefile("rwb")
        print(send(f, {"id": 1, "type": "hello"}))
        time.sleep(0.2)
        print(send(f, {"id": 2, "type": "snapshot", "include_text": True}))
        time.sleep(0.2)
        print(send(f, {"id": 3, "type": "text", "text": "hello from agent"}))
        time.sleep(0.2)
        print(send(f, {"id": 4, "type": "key", "key": "enter"}))
        time.sleep(0.2)
        print(send(f, {"id": 5, "type": "snapshot", "include_text": True}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
