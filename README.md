# gossip (ﾉ◕ヮ◕)ﾉ*:･ﾟ✧

> A lightweight, type‑safe event bus for Go — because sometimes your code just needs to talk behind your back.

---

### What is this?
`gossip` is an in‑process pub/sub library for Go. It lets your code publish events when something interesting happens (user created, login succeeded, token revoked, you name it) and lets other parts of your program subscribe without tight coupling. Think of it as a rumor mill for your application: one function whispers, the rest of the system listens in.

### Why?
- **Decoupling**: Core logic stays clean — side effects live in listeners.  
- **Extensibility**: Add logging, auditing, metrics, or notifications without touching the original code.  
- **Type‑safety**: Events are registered with identifiers, not raw strings, so you don’t get bitten by typos.  
- **Go‑native**: No Kafka, no brokers, no infra. Just a simple, fast, in‑memory bus.  

### Example use cases
- Fire an audit log entry every time a user logs in.  
- Trigger a notification when a resource is updated.  
- Collect metrics when tokens are issued or revoked.  
- Or just… spread some gossip.  

## Getting Started

```bash
go get github.com/seyedali-dev/gossip
```
