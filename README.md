# ⚡ Fixy
> CLI Error Hunter — No logins. No BS.

Fixy is a developer tool that helps you debug faster. Give it your file and your error — it scans your code for the risky line, shows you the context, and searches StackOverflow, GitHub Issues, and Reddit simultaneously.

No browser. No login. Just answers.

---

## Install

```bash
go install github.com/Bearcry55/fixy@latest
```

---

## Usage

**Just an error message:**
```bash
fixy "panic: index out of range"
```

**File + error (recommended):**
```bash
fixy main.go "panic: index out of range"
fixy server.py "list index out of range"
fixy app.rs "attempt to subtract with overflow"
```

Fixy will:
1. Read your file and find the risky line
2. Show the function name, line number, and surrounding context
3. Search online sources concurrently and show clickable links

---

## Demo

```
📄 File:   main.go
🔤 Lang:   go
──────────────────────────────────────────────────────────────────────
  🔍 FOUND IN FILE
──────────────────────────────────────────────────────────────────────

  ⚙ Function: processItems
  ⚠ Risk:    index access (potential out of range)
  Line 12:   fmt.Println(items[8])
    Context:
    10: func processItems(data []string) {
    11:         items := []string{"a", "b"}
    13: }

⚠  Error:  panic: index out of range

──────────────────────────────────────────────────────────────────────
  ⚡ ONLINE RESULTS
──────────────────────────────────────────────────────────────────────

  [StackOverflow] panic: runtime error: index out of range in Go
  Votes: 38 | Answers: 3 | ✓ Answered
  🔗 https://stackoverflow.com/questions/26126235/...
```

---

## What Fixy Detects

| Risk | Example |
|------|---------|
| Index out of range | `items[8]` on a 2-element slice |
| Nil dereference | `user.Name` when user could be nil |
| Unchecked error | `val, _ := someFunc()` |
| Type assertion panic | `x.(string)` without ok check |
| Possible deadlock | Unguarded `<-` channel ops |
| Divide by zero | `a / b` with no zero check |

---

## Supported Languages

| Extension | Language |
|-----------|----------|
| `.go` | Go |
| `.py` | Python |
| `.js` `.ts` `.jsx` `.tsx` | JavaScript / TypeScript |
| `.rs` | Rust |
| `.java` | Java |
| `.rb` | Ruby |
| `.cpp` `.cc` `.cxx` | C++ |
| `.cs` | C# |
| `.php` | PHP |

---

## Debug Mode

If results are missing or sources fail silently:

```bash
FIXY_DEBUG=1 fixy main.go "your error"
```

---

## Sources

- [StackOverflow](https://stackoverflow.com) — public API, no key needed
- [GitHub Issues](https://github.com) — public search API
- [Reddit](https://reddit.com) — language-specific subreddits

---

## License

MIT
