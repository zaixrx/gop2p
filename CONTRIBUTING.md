# Contributing to gop2p

I'm the sole maintainer of `gop2p`, a hybrid peer-to-peer networking library written in Go.

If you're curious about how peer-to-peer systems work or want to test your knowledge in networking, cryptography, or systems design, you're very welcome to contribute.

---

## Project Philosophy

This project favors clarity, full control, and extensibility. Since it's being used in my personal tooling, I try to avoid over-designing things.

Please:
- Stick to existing definitions and naming conventions.
- Follow the design decisions already made, unless you think there's a real improvement.
- If you're planning to **add a new system or major abstraction**, please **open a Discussion first** so we can talk through it.

---

## What You Can Contribute

### 1. Add More Transports
Path: `core/p2p/`

Right now only TCP is supported. I'd appreciate:
- UDP (stateful, not just for broadcasting)
- WebRTC (via pion)
- Any other connection method that fits the `t_transport` and `t_conn` interfaces in `transport.go`

### 3. Add End-to-End Encryption
Path: `core/secure/`

The `Packet` in type shared/packet.go already handles binary encoding/decoding. Encryption should act as a middleware layer that transforms the raw bytes before transmission — not interfere with packet structure.

#### Design Guideline

Encryption should follow this pipeline:

Packet ➜ Packet.GetBytes()[]byte ➜ Encrypt([]byte) ➜ Send ... Recv ➜ Decrypt([]byte) ➜ Packet.Load([]byte) ➜ Packet

This keeps encryption modular, pluggable, and easy to test.

A recommended interface:

```go
type Encryptor interface {
    Encrypt([]byte) ([]byte, error)
    Decrypt([]byte) ([]byte, error)
}
```

### 3. Abstract the Broadcaster
Path: `broadcasting/`

Currently, broadcasting is done over UDP. I'd like to turn it into a pluggable abstraction, so other methods (HTTP, WebSocket, WebRTC, etc.) can be added.

I'm planning to do this myself, but feel free to start a PR or a Discussion if you have ideas.

### 4. Extend the Logging System
Path: `logging/`

The logger system is interface-based and lives in `logging/def.go`. It already includes multiple log levels that should be respected.

### 5. Add Tests
Path: `tests/` and inline `*_test.go`

Testing network systems is hard. But I welcome tests for:
- Packet serialization/deserialization
- Mock transports
- Broadcast pool behavior
- Encoding/decoding messages

If you're unsure how to test something, open a Discussion and I’ll help you figure it out.

---

## How to Contribute

### Step 1: Fork the Repo
Click “Fork” on the GitHub page to create your own copy.

### Step 2: Clone and Branch
```bash
git clone https://github.com/YOUR_USERNAME/gop2p
cd gop2p
git checkout -b feature/my-cool-thing
```

### Step 3: Make Your Changes and Commit
```bash
git add .
git commit -m "Describe your change"
```

### Step 4: Push and Open a Pull Request
```bash
git push origin feature/my-cool-thing
```
Then open a PR from your branch into main on the original repo via GitHub.
