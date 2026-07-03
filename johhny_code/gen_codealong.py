# -*- coding: utf-8 -*-
"""
Generator for code_along.html — a COMPLETE, verbatim, build-from-scratch runbook.
Every functional .go file is embedded in full, pulled straight from disk and
HTML-escaped, so a learner can copy each snippet from the page and end up with a
byte-for-byte working backend. Every import in every file is explained from a
curated dictionary; every chapter has an authored "why" box.
"""
import os, re, html, sys

SRC = r"G:/GitSync/Johnny1110-crypto-exchange/crypto-exchange"
OUT = r"G:/GitSync/Johnny1110-crypto-exchange/johhny_code/code_along.html"

# ---------------------------------------------------------------------------
# 1. IMPORT EXPLANATIONS — why a developer adds each import.
# ---------------------------------------------------------------------------
IMPORTS = {
    # --- standard library ---
    "context": "Go's request-scoped handle, threaded through calls for cancellation/deadlines; by convention the first argument of DB and service methods.",
    "time": "durations, timestamps, and <code>time.Ticker</code> to drive scheduled loops.",
    "fmt": "formatted strings and error wrapping with <code>%w</code>.",
    "net/http": "HTTP status codes and, where an outbound client is needed, <code>*http.Client</code>.",
    "database/sql": "Go's driver-agnostic database interface (<code>*sql.DB</code> / <code>*sql.Tx</code>).",
    "sync": "mutexes / <code>RWMutex</code> to guard state shared across goroutines.",
    "sync/atomic": "lock-free atomic counters (orders/sec, trades) bumped on the hot path.",
    "errors": "create and compare sentinel error values.",
    "encoding/json": "marshal / unmarshal JSON for API payloads and WebSocket frames.",
    "strings": "string manipulation helpers.",
    "strconv": "convert between strings and numbers.",
    "math": "numeric helpers (rounding, min/max, <code>Sin</code> for wave pricing).",
    "math/big": "arbitrary-precision integers (Ethereum wei amounts).",
    "io": "read request/response bodies as a byte stream.",
    "io/ioutil": "read an entire response body in one call (legacy helper).",
    "os": "process environment, files and OS signals.",
    "log": "standard-library logging (used in a couple of low-level spots).",
    "path/filepath": "build filesystem paths portably across OSes.",
    "crypto/rand": "cryptographically-secure randomness.",
    "crypto/ecdsa": "elliptic-curve key pairs (Ethereum wallet keys).",
    # --- third party ---
    "github.com/labstack/gommon/log": "the leveled logger used across the app.",
    "github.com/gin-gonic/gin": "the HTTP web framework; handlers take a <code>*gin.Context</code>.",
    "github.com/gin-contrib/cors": "ready-made CORS middleware so the browser front-end may call the API.",
    "github.com/gorilla/websocket": "the WebSocket implementation: upgrades an HTTP request to a live socket.",
    "github.com/google/uuid": "generate unique IDs and login tokens.",
    "golang.org/x/crypto/bcrypt": "one-way password hashing and verification (never store plaintext).",
    "github.com/prometheus/client_golang/prometheus": "define Prometheus metrics (counters/gauges).",
    "github.com/prometheus/client_golang/prometheus/promhttp": "serve those metrics at <code>/metrics</code> for scraping.",
    "github.com/ncruces/go-sqlite3": "the pure-Go (CGO-free) SQLite package.",
    "github.com/ncruces/go-sqlite3/driver": "blank-imported: its <code>init()</code> registers the <code>sqlite3</code> driver with <code>database/sql</code>.",
    "github.com/ncruces/go-sqlite3/embed": "blank-imported: embeds the SQLite WASM build so no external C library is needed.",
    "github.com/emirpasic/gods/maps/treemap": "a sorted map (red-black tree) giving ordered price levels for O(log n) best-price lookup.",
    "github.com/emirpasic/gods/utils": "comparator helpers for the treemap (so we can sort prices ascending/descending).",
    "github.com/ethereum/go-ethereum/ethclient": "an Ethereum JSON-RPC client (on-chain address utilities).",
    "github.com/ethereum/go-ethereum/crypto": "Ethereum key/hash primitives.",
    "github.com/ethereum/go-ethereum/core/types": "Ethereum transaction/type definitions.",
    "github.com/ethereum/go-ethereum/common": "Ethereum common types (addresses, hashes).",
    # --- internal project packages ---
    "github.com/johnny1110/crypto-exchange/dto": "this project's neutral data shapes carried across layers (with their JSON tags).",
    "github.com/johnny1110/crypto-exchange/service": "the service-layer interfaces (business-logic contracts).",
    "github.com/johnny1110/crypto-exchange/engine-v2/model": "the engine's lean order and enum types.",
    "github.com/johnny1110/crypto-exchange/repository": "the storage interfaces plus the <code>DBExecutor</code> abstraction.",
    "github.com/johnny1110/crypto-exchange/engine-v2/book": "the order-book data structures and matching logic.",
    "github.com/johnny1110/crypto-exchange/utils": "shared helpers: the atomic system counters and math rounding.",
    "github.com/johnny1110/crypto-exchange/settings": "static configuration: the market and asset lists.",
    "github.com/johnny1110/crypto-exchange/ohlcv": "candlestick aggregation fed by the trade stream.",
    "github.com/johnny1110/crypto-exchange/engine-v2/core": "the matching-engine facade over the per-market books.",
    "github.com/johnny1110/crypto-exchange/engine-v2/market": "market metadata (base/quote asset, precision).",
    "github.com/johnny1110/crypto-exchange/security": "the token&rarr;user credential cache.",
    "github.com/johnny1110/crypto-exchange/ws": "the WebSocket hub, client and channel model.",
    "github.com/johnny1110/crypto-exchange/container": "the dependency-injection composition root.",
    "github.com/johnny1110/crypto-exchange/service/impl/amm": "the automated-market-maker bot.",
    "github.com/johnny1110/crypto-exchange/scheduler": "the timed background jobs.",
    "github.com/johnny1110/crypto-exchange/engine-v2/util": "the per-price-level doubly-linked deque.",
    "github.com/johnny1110/crypto-exchange/controller": "the thin HTTP handlers.",
    "github.com/johnny1110/crypto-exchange/service/serviceHelper": "settlement math and order helpers shared by the services.",
    "github.com/johnny1110/crypto-exchange/service/impl/metrics": "Prometheus metric collectors.",
    "github.com/johnny1110/crypto-exchange/service/impl": "the concrete service implementations (for the container to wire).",
    "github.com/johnny1110/crypto-exchange/repository/impl": "the concrete SQL repository implementations.",
    "github.com/johnny1110/crypto-exchange/middleware": "the auth / CORS / error middleware.",
    "github.com/johnny1110/crypto-exchange/external": "outbound calls to the external index-price API.",
}

def explain_import(path):
    if path in IMPORTS:
        return IMPORTS[path]
    # internal subpackage fallback
    if path.startswith("github.com/johnny1110/crypto-exchange/"):
        return "an internal project package (" + path.split("/")[-1] + ")."
    return "third-party dependency."

# ---------------------------------------------------------------------------
# 2. Parse a file's import block into (alias, path) pairs.
# ---------------------------------------------------------------------------
def parse_imports(text):
    out = []
    # block form: import ( ... )
    m = re.search(r'^import\s*\((.*?)^\)', text, re.S | re.M)
    lines = []
    if m:
        lines = m.group(1).splitlines()
    # single form: import "x" or import alias "x"
    for sm in re.finditer(r'^import\s+((?:[\w.]+\s+)?"[^"]+")\s*$', text, re.M):
        lines.append(sm.group(1))
    for ln in lines:
        ln = ln.strip()
        if not ln or ln.startswith("//"):
            continue
        mm = re.match(r'(?:(\S+)\s+)?"([^"]+)"', ln)
        if mm:
            out.append((mm.group(1), mm.group(2)))
    return out

def esc(s):
    return html.escape(s, quote=False)

def import_block_html(text):
    imps = parse_imports(text)
    if not imps:
        return '<p class="noimp">No imports yet &mdash; this piece depends on nothing outside its own package.</p>'
    items = []
    for alias, path in imps:
        label = esc(path)
        note = ""
        if alias == "_":
            note = ' <span class="alias">(blank import <code>_</code> &mdash; added only for its <code>init()</code> side-effect, never named)</span>'
        elif alias and alias not in (".",):
            note = ' <span class="alias">(aliased as <code>%s</code>)</span>' % esc(alias)
        items.append('<li><code>"%s"</code>%s &mdash; %s</li>' % (label, note, explain_import(path)))
    return '<p class="imphdr">The imports we add now, and why each one:</p>\n<ul class="imp">\n%s\n</ul>' % "\n".join(items)

# --- Split a Go file into the pieces a developer adds one at a time. -------
DECL_RE = re.compile(r'^(func|type|var|const)\b')

def split_go(text):
    lines = text.split("\n")
    # header = package line + import block + any leading doc, up to the first top-level decl
    first = None
    in_import = False
    for idx, ln in enumerate(lines):
        if re.match(r'^import\s*\(', ln):
            in_import = True; continue
        if in_import:
            if ln.startswith(")"): in_import = False
            continue
        if DECL_RE.match(ln):
            first = idx; break
    if first is None:
        return [("header", lines)]
    header = lines[:first]
    body = lines[first:]
    segs, buf = [], []
    def has_code(b):
        return any(l.strip() and not l.lstrip().startswith("//") for l in b)
    for ln in body:
        if DECL_RE.match(ln) and has_code(buf):
            lead = []
            while buf and (buf[-1].strip() == "" or buf[-1].lstrip().startswith("//")):
                lead.insert(0, buf.pop())
            segs.append(buf)
            buf = lead + [ln]
        else:
            buf.append(ln)
    if buf:
        segs.append(buf)
    result = [("header", header)] + [("decl", s) for s in segs]
    # SAFETY: concatenation MUST reproduce the file exactly — no line lost or added.
    assert "\n".join(sum((s for _, s in result), [])) == text, "reassembly mismatch"
    return result

def label_for(seg_lines):
    decl = next((l for l in seg_lines if l.strip() and not l.lstrip().startswith("//")), seg_lines[0])
    for rx, kind, gi in [
        (r'func\s+\(\s*\w+\s+\*?(\w+)\)\s+(\w+)', "method", None),
        (r'func\s+(\w+)', "function", 1),
        (r'type\s+(\w+)\s+struct', "struct", 1),
        (r'type\s+(\w+)\s+interface', "interface", 1),
        (r'type\s+(\w+)', "type", 1),
        (r'const\s*\(', "const-block", 0),
        (r'var\s*\(', "var-block", 0),
        (r'const\s+(\w+)', "const", 1),
        (r'var\s+(\w+)', "var", 1),
    ]:
        m = re.match(rx, decl)
        if m:
            if kind == "method":
                return kind, "%s.%s" % (m.group(1), m.group(2))
            if gi == 0:
                return kind, ""
            return kind, m.group(1)
    return "code", ""

KINDWORD = {
    "method": "the method", "function": "the function", "struct": "the struct",
    "interface": "the interface", "type": "the type", "const": "the constant",
    "var": "the variable", "const-block": "the constants block",
    "var-block": "the variables block", "code": "the next piece",
}

# Hand-authored developer-voice notes for the pivotal pieces (keyed relpath::name).
AUTHORED = {
    "engine-v2/book/orderbook.go::OrderBook.PlaceOrder": "The one entrance to matching. Lock the book, branch on limit vs market, match against the opposite side, then rest any remainder as a maker. One <code>Mutex</code> means one order at a time per market &mdash; the invariant that keeps the book consistent.",
    "engine-v2/book/orderbook.go::OrderBook.canMatch": "The whole idea of a trade in one function: a buy fills against a sell only when the bid price is at least the ask price. Everything else is bookkeeping around this rule.",
    "service/impl/txn.go::WithTx": "Write begin/commit/rollback exactly once, with rollback in a <code>defer</code>, so no caller can ever forget it. Every money write flows through here.",
    "ws/hub.go::Hub.Run": "The trick that avoids locks: ONE goroutine owns the client map; everyone else sends it a message on a channel. Concurrency by communication, not shared memory.",
    "engine-v2/core/engine.go::MatchingEngine.PlaceOrder": "The facade: look up the right market's book and delegate. Callers above never touch an individual book.",
    "security/credential.go::CredentialCache.Get": "Read-hot path: hit the in-memory map first; only on a miss (e.g. after a restart) rebuild the user from the persisted token. That rebuild is why sessions survive a deploy.",
}

# Per-file "why this file exists" and extra per-piece notes are loaded from a
# JSON file (authored by reading every file, so nothing here is guessed).
FILE_WHY = {}
NOTES_PATH = os.path.join(os.path.dirname(OUT), "file_notes.json")
if os.path.exists(NOTES_PATH):
    import json
    for o in json.load(open(NOTES_PATH, encoding="utf-8")):
        if o.get("why"):
            FILE_WHY[o["path"]] = o["why"]
        for kp in (o.get("key_pieces") or []):
            AUTHORED.setdefault("%s::%s" % (o["path"], kp["name"]), kp["note"])

# Match a piece note by the final identifier (method/func name), so a note keyed
# "Apply" or "CacheKeyPrefix.Apply" both resolve for the Apply method.
KP_BY_IDENT = {}
for key, note in AUTHORED.items():
    path, name = key.split("::", 1)
    KP_BY_IDENT.setdefault((path, name.split(".")[-1]), note)

def file_block(relpath, note=""):
    full = os.path.join(SRC, relpath)
    with open(full, "r", encoding="utf-8") as f:
        text = f.read()
    trailing_nl = text.endswith("\n")
    body = text[:-1] if trailing_nl else text
    n = body.count("\n") + 1
    segs = split_go(body)
    parts = ["<div class=\"filecard\">"]
    parts.append('<div class="file-path">%s <span class="lc">%d lines &middot; built in %d pieces</span></div>' % (esc(relpath), n, len(segs)))
    why = FILE_WHY.get(relpath) or note
    if why:
        parts.append('<div class="filewhy">%s</div>' % why)
    parts.append('<p class="buildnote">We grow this file the way a developer does &mdash; one declaration at a time. Type each piece in order; together they <em>are</em> the complete file.</p>')
    piece = 0
    for kind, lines in segs:
        piece += 1
        code = "\n".join(lines)
        if kind == "header":
            parts.append('<div class="piece"><div class="piece-h"><span class="pn">%d</span> Start the file &mdash; <code>package</code> &amp; imports</div>' % piece)
            parts.append('<pre><code>%s</code></pre>' % esc(code))
            parts.append(import_block_html(text))
            parts.append('</div>')
        else:
            k, name = label_for(lines)
            title = "Add %s%s" % (KINDWORD.get(k, "the next piece"), (" <code>%s</code>" % esc(name)) if name else "")
            parts.append('<div class="piece"><div class="piece-h"><span class="pn">%d</span> %s</div>' % (piece, title))
            pnote = KP_BY_IDENT.get((relpath, name.split(".")[-1])) if name else None
            if pnote:
                parts.append('<p class="pnote">%s</p>' % pnote)
            parts.append('<pre><code>%s</code></pre>' % esc(code))
            parts.append('</div>')
    parts.append('</div>')
    return "\n".join(parts)

# ---------------------------------------------------------------------------
# 3. The chapter manifest — build order. Each: (anchor, title, why_html, [ (relpath, note), ... ], checkpoint_html)
# ---------------------------------------------------------------------------
# Notes are optional per-file one-liners. Prose lives in why boxes / checkpoints.
CH = []
def chapter(anchor, title, why, files, chk, extra_top=""):
    CH.append((anchor, title, why, files, chk, extra_top))

chapter("c0", "Chapter 0 — Empty folder to skeleton",
  "Every Go project starts the same way: one command that turns a directory into a <em>module</em> (a versioned unit of code with a name and a dependency list). The module path is also the import prefix every internal package uses, which is why we set it first &mdash; everything you build later imports <code>github.com/johnny1110/crypto-exchange/...</code>.",
  [],
  "You have a module and an empty folder tree. Nothing compiles yet &mdash; that's expected; we build inward-out so each file's dependencies already exist when you type it.",
  extra_top="""
<div class="step"><span class="step-tag">STEP 0.1</span> Create the project and initialise the module.
<pre><code>mkdir crypto-exchange &amp;&amp; cd crypto-exchange
go mod init github.com/johnny1110/crypto-exchange</code></pre>
<p>This writes <code>go.mod</code> with your module path and Go version. From now on, every <code>go get</code> you run (each introduced the first time a file needs it) appends a line here; running <code>go mod tidy</code> at the end reconciles the full list. Below is the complete dependency set this project ends up with &mdash; you do <em>not</em> add them now; each chapter tells you the exact <code>go get</code> when its first importing file appears.</p>
<div class="why">Why a module and not loose files? Go compiles <em>packages</em>, and packages are located by the module path. Without <code>go mod init</code> there's no name to import your own code under, and no manifest to pin dependency versions &mdash; builds wouldn't be reproducible.</div>
</div>
<div class="step"><span class="step-tag">STEP 0.2</span> Create the folder skeleton (one folder = one package = one responsibility).
<pre><code>mkdir -p utils settings engine-v2/model engine-v2/market engine-v2/util \\
  engine-v2/book engine-v2/core dto repository repository/impl \\
  service service/impl service/impl/amm service/impl/metrics service/serviceHelper \\
  security middleware controller ohlcv ws scheduler container external chainUtil</code></pre>
<div class="why">Why split into so many folders now? Because in Go the <em>folder is the unit of dependency</em>: package A importing package B is a compiler-enforced arrow. Laying the folders out first lets us build strictly inward (utils &rarr; engine &rarr; services &rarr; web) so dependencies only ever point one way &mdash; the structure that keeps the fast core unaware of the slow edges.</div>
</div>
""")

chapter("c1", "Chapter 1 — Foundations: counters, math, settings",
  "Before any business logic, we lay down the tiny leaf packages that <em>everything</em> depends on and that depend on nothing themselves. <code>utils</code> holds the atomic system counters (orders/sec, trades) and rounding math; <code>settings</code> holds the static list of markets and assets. Leaf-first means when we build the engine next, its imports already exist.",
  [("utils/counter.go","The live system counters. Uses <code>sync/atomic</code> so the hot path can bump orders/trades with no lock, plus a running-duration clock the UI reads."),
   ("utils/math.go","Rounding helpers so prices/sizes land on clean increments."),
   ("settings/settings.go","The single source of truth for which markets and assets exist. The engine, container and OHLCV all read this."),
   ("settings/cache_key.go","Constant cache-key names, kept in one place so producers and consumers never disagree on a string.")],
  "<code>go build ./utils ./settings</code> compiles. The bedrock is down.")

chapter("c2", "Chapter 2 — The engine's vocabulary: orders &amp; markets",
  "The matching engine needs its own <em>lean</em> types &mdash; just what matching requires, no JSON tags, no DB columns. <code>model.Order</code> plus the enums (Side, Mode, OrderType, OrderStatus) are that vocabulary; <code>market.MarketInfo</code> describes a tradable pair. Keeping these separate from the API/DB shapes (the DTOs, later) is what lets the engine stay fast and dependency-free.",
  [("engine-v2/model/order.go","The heart of the engine's data: the Side/Mode/OrderType/OrderStatus enums (via <code>iota</code>), the <code>Order</code> struct the book matches on, and the <code>OrderNode</code> wrapper used inside price levels."),
   ("engine-v2/market/market.go","A tradable pair's metadata: base/quote asset and precision. No imports &mdash; pure data.")],
  "<code>go build ./engine-v2/model ./engine-v2/market</code> compiles. The engine now has words to think in.")

chapter("c3", "Chapter 3 — Engine data structures: deque, book side, index",
  "A fast order book is really three cooperating structures: a <strong>deque</strong> per price level (O(1) FIFO for price-time priority), a <strong>book side</strong> (a sorted map of price&rarr;deque, so best price is O(log n)), and an <strong>order index</strong> (a hash map order-id&rarr;node, so cancel is O(1)). We build them bottom-up before the book that combines them.",
  [("engine-v2/util/deque.go","A doubly-linked deque of orders for one price level: push-back for new makers, pop-front for the oldest &mdash; that ordering <em>is</em> price-time priority."),
   ("engine-v2/book/book_side.go","One side (bids or asks) as a <code>treemap</code> of price&rarr;deque. The comparator direction decides whether best is highest (bids) or lowest (asks)."),
   ("engine-v2/book/order_index.go","A map from order-id to its node so an incoming cancel finds and unlinks the order in O(1) instead of scanning.")],
  "<code>go build ./engine-v2/util ./engine-v2/book</code> (book_side/order_index) compiles. The scaffolding for matching is ready.")

chapter("c4", "Chapter 4 — The order book &amp; the matching loop",
  "This is the crown jewel. <code>orderbook.go</code> ties the two sides + index together and implements <code>PlaceOrder</code>: a taker walks the opposite side, matching against the best price while it crosses, generating trades, and any unfilled remainder rests as a maker. Every maker/taker, limit/market rule lives here, under one mutex.",
  [("engine-v2/book/orderbook.go","The full matching engine for one market: PlaceOrder, the limit/market paths, <code>canMatch</code> (the price-cross rule), the taker loop that drains liquidity, and the Trade type. Guarded by one <code>sync.Mutex</code> so a market matches one order at a time &mdash; the invariant that keeps the book consistent.")],
  "<code>go build ./engine-v2/book</code> compiles. A single market can now match orders entirely in memory.")

chapter("c5", "Chapter 5 — The engine facade",
  "A real exchange has many markets. <code>core.MatchingEngine</code> is the facade that owns one <code>OrderBook</code> per market and routes an order to the right one, behind an <code>RWMutex</code> so adding a market never races with matching. Everything above the engine talks to <em>this</em>, not to individual books.",
  [("engine-v2/core/engine.go","Owns the map of market&rarr;OrderBook, exposes AddMarket / PlaceOrder / CancelOrder / snapshots, and the Reset used by the demo's reset button. The single door into the whole engine.")],
  "<code>go build ./engine-v2/...</code> compiles. The entire matching engine is done and has no idea a database or web server exists.")

chapter("c6", "Chapter 6 — DTOs: the shapes that cross layers",
  "The engine's types are lean; the API and DB need richer shapes with JSON tags, status, fees and timestamps. The <code>dto</code> package holds those neutral shapes, imported by both the web edge and the services, so neither the engine nor the storage layer leaks into the other. It's also where request/response envelopes live.",
  [("dto/users.go","The User shape; note <code>json:\"-\"</code> hides the password hash from every response."),
   ("dto/balances.go","Available / Locked / Total per asset &mdash; the two-column money model surfaced to the API."),
   ("dto/orders.go","The rich Order (market, status, fees, timestamps) plus <code>ToEngineOrder()</code>, the one translation point to the engine's lean order."),
   ("dto/trade.go","A trade as the API/DB sees it."),
   ("dto/markets.go","Market summary shape for the markets endpoint."),
   ("dto/req_struct.go","Inbound request bodies/queries, with <code>form</code>/<code>json</code> binding tags."),
   ("dto/resp_struct.go","The uniform response envelope + pagination wrapper.")],
  "<code>go build ./dto</code> compiles. Both edges now share one neutral vocabulary.")

chapter("c7", "Chapter 7 — The database &amp; startup",
  "In-memory is fast but volatile; money must be durable. <code>db.go</code> opens SQLite tuned for throughput (WAL + <code>synchronous=NORMAL</code>, single writer), ensures the extra tables this build added (1-minute candles, persisted credentials) and seeds counters. <code>startup_helper.go</code> loads schema/mock data and recovers the order book from disk so a restart loses nothing.",
  [("db.go","Opens and tunes SQLite (the WAL pragmas are the ~10&times; throughput win), ensures the <code>ohlcv_1min</code> and <code>credentials</code> tables, purges garbage candles, and seeds the live counters from history."),
   ("startup_helper.go","Runs the SQL seed files in a transaction (test mode) and rebuilds the in-memory order book from still-open orders on boot &mdash; the reconciliation that makes the in-memory engine restart-safe.")],
  "<code>go build .</code> (root package) compiles the DB layer. There is now a durable home for every order, trade and balance.")

chapter("c8", "Chapter 8 — Repository interfaces",
  "Services must not contain SQL &mdash; that would weld money logic to table shapes. The <code>repository</code> package declares interfaces (what storage can do) and the <code>DBExecutor</code> trick: one interface satisfied by <em>both</em> <code>*sql.DB</code> and <code>*sql.Tx</code>, so every repo method works inside a transaction or standalone with no duplicated variants.",
  [("repository/interface.go","All storage contracts (users, balances, orders, trades) plus <code>DBExecutor</code>. The services depend only on these, never on the SQL.")],
  "<code>go build ./repository</code> compiles. Storage is now an abstract contract.")

chapter("c9", "Chapter 9 — Repository implementations (the real SQL)",
  "Now the hand-written SQL behind those interfaces. No ORM: an exchange needs exact control over atomic balance moves and locking. The star is the balance repo's freeze &mdash; <code>UPDATE ... WHERE available &gt;= ?</code> makes 'check funds' and 'move funds' one atomic step, which is where double-spends are physically prevented.",
  [("repository/impl/users.go","Insert/find users; the login lookup reads the password hash for bcrypt verification."),
   ("repository/impl/balances.go","The money table: atomic lock/unlock/modify with the <code>available &gt;= ?</code> guard, plus batch-create zeroed balances for a new user."),
   ("repository/impl/orders.go","Persist and query orders: insert, status updates, a user's history, active orders per market, latest price."),
   ("repository/impl/trades.go","Batch-insert trades and read recent trades / 24h volume.")],
  "<code>go build ./repository/...</code> compiles. The money-safety rule now lives in SQL, not in guesswork.")

chapter("c10", "Chapter 10 — Transactions &amp; settlement math",
  "Placing an order writes several rows that must all-or-nothing together. <code>WithTx</code> writes the begin/commit/rollback envelope <em>once</em> (rollback in a <code>defer</code> so you can't forget it) and takes your writes as a function. The <code>serviceHelper</code> package holds the settlement arithmetic &mdash; fees, who-gets-what per fill &mdash; kept separate so it's unit-testable in isolation.",
  [("service/impl/txn.go","The <code>WithTx</code> helper: one place that guarantees commit-or-rollback around any block of writes."),
   ("service/serviceHelper/helpers.go","Order-construction helpers: turn a DTO order into an engine order, compute freeze asset/amount, etc."),
   ("service/serviceHelper/settlement.go","The per-trade settlement math: maker/taker fees, base/quote transfers, balance deltas both sides.")],
  "<code>go build ./service/impl ./service/serviceHelper</code> (these files) compiles. The safety net and the money arithmetic are ready.")

chapter("c11", "Chapter 11 — Service interfaces",
  "The service layer is the business brain. We declare its contracts first (as with repositories) so controllers, the AMM bot and schedulers can all depend on <em>interfaces</em> and be tested with fakes. One file names every capability the app offers.",
  [("service/interface.go","Every service contract: users, balances, orders, order-book, admin, cache, market-data. The public surface of the business logic.")],
  "<code>go build ./service</code> compiles. The business capabilities are now a contract.")

chapter("c12", "Chapter 12 — Service implementations (the money brain)",
  "Here the rules live. The order service's two-phase placement &mdash; freeze &rarr; insert &rarr; match &rarr; persist &rarr; settle, all inside one <code>WithTx</code> &mdash; is the most important code in the project. The other services (users with bcrypt, balances, markets, caches, admin reset, metrics) round out the brain.",
  [("service/impl/users.go","Register (bcrypt-hash + create zeroed balances) and Login (verify + issue token)."),
   ("service/impl/balances.go","Read a user's balances, valued via market data."),
   ("service/impl/orders.go","The two-phase order placement and cancellation &mdash; the atomic freeze/match/settle path, plus the live counters."),
   ("service/impl/orderbooks.go","Read-side access to the engine's order-book snapshots."),
   ("service/impl/markets.go","24h stats, last price, market summaries."),
   ("service/impl/caches.go","A tiny in-memory cache used by market data."),
   ("service/impl/admins.go","Manual balance adjust and the full ResetExchange (wipe + refund) behind the demo's reset button."),
   ("service/impl/metrics/metrics.go","Prometheus collectors exposing order-book/scheduler health.")],
  "<code>go build ./service/...</code> compiles. The brain is complete and testable.")

chapter("c13", "Chapter 13 — Security: the token store",
  "Login returns a token; every later request presents it. The credential cache maps token&rarr;user, backed by a DB table so a token survives a restart (a memory-only cache would log everyone out on every deploy). An <code>RWMutex</code> fits its read-heavy access.",
  [("security/credential.go","The DB-backed token&rarr;user cache: Put/Get/Delete with an in-memory map guarded by <code>RWMutex</code>, rehydrating from the <code>credentials</code> table on a miss.")],
  "<code>go build ./security</code> compiles. Auth has a durable home.")

chapter("c14", "Chapter 14 — Middleware: auth, CORS, errors",
  "Behaviour that must wrap <em>every</em> request &mdash; check the token, allow the browser, turn errors into clean JSON &mdash; belongs in middleware, not copied into each handler. A middleware is a closure that captures its dependency once and runs per request.",
  [("middleware/middleware.go","The auth middleware (resolve token via the credential cache, stash the user, abort on failure), CORS, and error handling.")],
  "<code>go build ./middleware</code> compiles.")

chapter("c15", "Chapter 15 — Controllers: the HTTP handlers",
  "Controllers are deliberately thin: read the user + inputs, call one service method, shape the JSON reply. No business logic here &mdash; that's the service's job &mdash; which is why the same logic is reachable from a test or a bot without duplication.",
  [("controller/resp.go","The shared success/error response shaping (uniform envelope + code mapping)."),
   ("controller/users.go","Register / Login / Profile / Logout."),
   ("controller/balances.go","Read balances for the authenticated user."),
   ("controller/orders.go","Place / cancel / list orders."),
   ("controller/orderbooks.go","Public order-book snapshot."),
   ("controller/markets.go","Public markets + stats."),
   ("controller/admins.go","Admin balance adjust and reset.")],
  "<code>go build ./controller</code> compiles.")

chapter("c16", "Chapter 16 — The router: wiring URLs to handlers",
  "The router maps URL&rarr;handler and assembles the middleware chain. Public routes need no token; private ones sit behind the auth middleware as a <em>group</em>, so the security boundary is visible in the file's shape and impossible to forget on one endpoint.",
  [("router.go","Registers the public group, the private (authenticated) group, admin routes, the <code>/metrics</code> endpoint and the WebSocket upgrade route.")],
  "<code>go build .</code> &mdash; the HTTP half of the app is wired.")

chapter("c17", "Chapter 17 — OHLCV: trades into candlesticks",
  "A chart needs candles; trades arrive as a fast stream. The aggregator folds each trade into the current bucket for every interval and flushes closed bars to the DB in batches, decoupled from matching by a buffered channel so a trade burst never blocks an order. This is a whole subsystem &mdash; interface, models, per-symbol bars, the stream, workers, the aggregator and its SQLite repo.",
  [("ohlcv/interface.go","The OHLCV repository and stream contracts."),
   ("ohlcv/model.go","Candle/bar and trade shapes for the aggregator."),
   ("ohlcv/config.go","Batch size, flush interval, channel size, concurrency knobs."),
   ("ohlcv/symbol_bars.go","Per-symbol, per-interval current-bar state (open/high/low/close/volume)."),
   ("ohlcv/trade_stream.go","The buffered trade channel with a non-blocking send (drop-for-charting rather than stall the match)."),
   ("ohlcv/workers.go","The background workers that consume trades and fold them into bars."),
   ("ohlcv/aggregator.go","The orchestrator: add symbols, run the flush ticker, batch-write closed bars."),
   ("ohlcv/repository.go","The SQLite persistence for every interval table, plus the reads the chart endpoint uses."),
   ("ohlcv/utils.go","Bucket-boundary and rounding helpers.")],
  "<code>go build ./ohlcv/...</code> compiles. Trades now become chart data off the hot path.")

chapter("c18", "Chapter 18 — WebSocket: pushing live data",
  "REST is pull; WebSocket is push. The hub is one goroutine that owns the client set and fans messages out over per-client channels &mdash; no locks on the map, no races. Public channels broadcast to all; private <code>user_data</code> goes only to the owning token.",
  [("ws/model.go","The channel names (ORDERBOOK/OHLCV/MARKETS/SYSSTATS public, USER_DATA private) and the subscribe request params &mdash; the contract the front-end must match exactly."),
   ("ws/client.go","One connected browser: its socket, its buffered <code>Send</code> queue (raised to 1024 for bursty data), its subscriptions and token, and the read/write pumps."),
   ("ws/hub.go","The single owner-goroutine: register/unregister/broadcast via a <code>select</code> loop, delivering only what each client subscribed to."),
   ("ws/mock.go","Helpers that build sample frames (used for wiring/dev).")],
  "<code>go build ./ws/...</code> compiles. The server can push live state.")

chapter("c19", "Chapter 19 — Schedulers: the heartbeats",
  "Some work runs on a clock, not per-request: snapshot the book every 300ms, push WS frames every 400ms&ndash;1s, refresh 24h stats every 30s, top up AMM liquidity every 5 min. Each scheduler is a goroutine with a ticker; a reporter tracks their health.",
  [("scheduler/interfaces.go","The Scheduler contract (Start/Stop/health)."),
   ("scheduler/init.go","Package init glue."),
   ("scheduler/scheduler_reporter.go","Aggregates each scheduler's last-run/health for the metrics endpoint."),
   ("scheduler/market_job.go","Refreshes 24h market stats into the cache."),
   ("scheduler/orderbook_snapshot_job.go","Builds the cached order-book snapshot every 300ms so readers never touch the live book."),
   ("scheduler/lqdt_amm_job.go","Wakes the AMM to replenish liquidity on an interval."),
   ("scheduler/ws_data_feeder.go","The feeder that makes the UI feel alive: two tickers pushing orderbook (fast) and ohlcv/markets/sysstats/user_data (slow) to subscribers.")],
  "<code>go build ./scheduler/...</code> compiles. The system now has a pulse.")

chapter("c20", "Chapter 20 — AMM: a bot that provides liquidity",
  "An empty book is a dead market. The automated market maker continuously quotes both sides around a reference price so there's always something to trade against &mdash; and it does so through the <em>same</em> service interfaces a human uses, inheriting every money-safety guarantee for free. That reuse is the payoff of the layering.",
  [("service/impl/amm/amm_strategies.go","The strategy enum/selection: how the bot decides its quotes."),
   ("service/impl/amm/strategy_helper.go","Quote math: spread, sizing, reference price around which to post."),
   ("service/impl/amm/provide_liquidity_amm.go","The proxy that actually places the bot's orders via IOrderService &mdash; just another client of the services.")],
  "<code>go build ./service/impl/amm/...</code> compiles. The market can quote itself.")

chapter("c21", "Chapter 21 — External data &amp; chain utilities",
  "Two small support packages the wiring needs: <code>external</code> fetches an index price to seed the first candle sensibly (with a safe fallback when the datacenter IP is blocked), and <code>chainUtil</code> holds Ethereum address helpers used by the on-chain-flavoured parts of the demo.",
  [("external/index_price_api.go","The outbound index-price call with the 403/低-price fallback that avoids seeding a garbage first candle."),
   ("chainUtil/chainUtils.go","Ethereum key/address helpers (go-ethereum): generate/derive addresses for the wallet-flavoured demo flows.")],
  "<code>go build ./external ./chainUtil</code> compiles. The support packages are in place.")

chapter("c22", "Chapter 22 — The container: composition root",
  "Every layer depended only on interfaces; <em>someone</em> must create the real objects and hand each its dependencies. The container is that single place, constructing the graph bottom-up (repos &rarr; OHLCV &rarr; services &rarr; hub &rarr; AMM &rarr; schedulers). Change an implementation and you edit one file.",
  [("container/container.go","The whole object graph's birthplace: builds every repo, service, the OHLCV aggregator (seeded per market), the WS hub (and launches its goroutine), the AMM proxy and all schedulers &mdash; dependency injection with no framework, just passed arguments.")],
  "<code>go build ./container</code> compiles. The graph has one origin.")

chapter("c23", "Chapter 23 — main.go: ignition",
  "<code>main()</code> is the precise boot sequence: open the DB, build the engine, register markets, <strong>recover</strong> the book from disk, build the container, start schedulers, mount routes, listen &mdash; and catch signals for a graceful shutdown. Order matters: you can't recover into an engine you haven't built or serve routes before the container exists.",
  [("main.go","The entry point: initDB &rarr; NewMatchingEngine + AddMarket per market &rarr; recover book &rarr; NewContainer &rarr; start schedulers &rarr; register routes &rarr; listen, with <code>defer Cleanup()</code> and signal handling.")],
  "<code>go build -o exchange .</code> from the project root produces the binary. <strong>The entire backend is now written &mdash; every line, no gaps.</strong>")

# ---------------------------------------------------------------------------
# 4. Assemble.
# ---------------------------------------------------------------------------
CSS = """
:root{--bg:#0d1117;--card:#161b22;--card2:#1c2330;--ink:#e6edf3;--mut:#9aa7b4;--accent:#58a6ff;--accent2:#3fb950;--warn:#d29922;--line:#30363d;--code:#0b0f14;}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--ink);font:16px/1.65 -apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;}
.wrap{max-width:1000px;margin:0 auto;padding:28px 20px 120px;}
h1{font-size:2em;line-height:1.2;margin:.2em 0 .1em}
h2{font-size:1.5em;margin:2.4em 0 .6em;padding-top:.6em;border-top:2px solid var(--line);color:#fff}
h3{font-size:1.15em;margin:1.6em 0 .5em;color:var(--accent)}
a{color:var(--accent);text-decoration:none}a:hover{text-decoration:underline}
p{margin:.6em 0}
code{font-family:"SF Mono",Consolas,Menlo,monospace;font-size:.86em;background:#20262e;padding:.1em .35em;border-radius:4px;color:#e6edf3}
pre{background:var(--code);border:1px solid var(--line);border-radius:8px;padding:14px 16px;overflow-x:auto;margin:.7em 0}
pre code{background:none;padding:0;font-size:.82em;line-height:1.5;color:#d7e0ea;white-space:pre}
pre.full{max-height:none}
.intro{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:20px 22px;margin:18px 0}
.why{background:rgba(88,166,255,.08);border-left:4px solid var(--accent);border-radius:0 8px 8px 0;padding:12px 16px;margin:14px 0}
.why::before{content:"WHY WE DO THIS";display:block;font-size:.7em;letter-spacing:.12em;color:var(--accent);font-weight:700;margin-bottom:4px}
.step{background:var(--card2);border:1px solid var(--line);border-radius:8px;padding:12px 16px;margin:14px 0}
.step-tag{display:inline-block;background:var(--accent);color:#04101f;font-weight:700;font-size:.72em;letter-spacing:.05em;padding:2px 8px;border-radius:5px;margin-right:8px;vertical-align:middle}
.chk{background:rgba(63,185,80,.09);border-left:4px solid var(--accent2);border-radius:0 8px 8px 0;padding:12px 16px;margin:16px 0}
.chk::before{content:"\\2713 CHECKPOINT";display:block;font-size:.7em;letter-spacing:.1em;color:var(--accent2);font-weight:700;margin-bottom:4px}
.filecard{border:1px solid var(--line);border-radius:10px;margin:20px 0;padding:0 0 4px;background:var(--card)}
.file-path{font-family:"SF Mono",Consolas,monospace;font-size:.9em;font-weight:700;color:#fff;background:#11161d;border-bottom:1px solid var(--line);border-radius:10px 10px 0 0;padding:9px 14px}
.file-path .lc{float:right;font-weight:400;color:var(--mut);font-size:.85em}
.filenote{padding:2px 14px 0;color:var(--mut);font-style:italic}
.filewhy{margin:10px 14px 2px;padding:10px 14px;background:rgba(210,153,34,.09);border-left:4px solid var(--warn);border-radius:0 8px 8px 0;font-size:.95em}
.filewhy::before{content:"WHY THIS FILE EXISTS";display:block;font-size:.68em;letter-spacing:.11em;color:var(--warn);font-weight:700;margin-bottom:4px}
.imphdr{padding:6px 14px 0;margin:.4em 0 .1em;font-weight:600;color:var(--accent)}
p.noimp{padding:8px 14px;color:var(--mut)}
ul.imp{margin:.2em 0 .4em;padding:0 14px 0 34px}
ul.imp li{margin:3px 0;font-size:.93em}
ul.imp .alias{color:var(--warn);font-size:.9em}
.filecard pre{margin:8px 14px 12px;border-radius:8px}
.buildnote{padding:6px 14px 0;color:var(--accent2);font-size:.9em}
.piece{border-top:1px dashed var(--line);margin:0;padding:2px 0 0}
.piece:first-of-type{border-top:none}
.piece-h{font-weight:600;color:#fff;padding:10px 14px 2px;font-size:.96em}
.piece-h .pn{display:inline-block;min-width:22px;height:22px;line-height:22px;text-align:center;background:var(--accent);color:#04101f;border-radius:50%;font-size:.72em;font-weight:700;margin-right:8px}
.pnote{padding:2px 14px 0;color:var(--ink);font-size:.93em;background:rgba(88,166,255,.06);border-left:3px solid var(--accent);margin:4px 14px 0;border-radius:0 6px 6px 0;padding:8px 12px}
.toc{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:16px 22px;margin:18px 0;columns:2;column-gap:32px}
.toc a{display:block;padding:3px 0;color:var(--ink)}
.toc a:hover{color:var(--accent)}
.tag{display:inline-block;background:var(--accent2);color:#03130a;font-weight:700;font-size:.68em;padding:2px 7px;border-radius:5px;margin-left:6px}
@media(max-width:720px){.toc{columns:1}.wrap{padding:20px 14px 90px}h1{font-size:1.6em}h2{font-size:1.3em}pre code{font-size:.76em}}
"""

def build():
    total_files = sum(len(c[3]) for c in CH)
    out = []
    out.append("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">")
    out.append('<meta name="viewport" content="width=device-width, initial-scale=1">')
    out.append("<title>Code-Along: build the whole crypto-exchange backend, line by line</title>")
    out.append("<style>%s</style>\n</head>\n<body>\n<div class=\"wrap\">" % CSS)
    out.append("<h1>Code-Along &mdash; build the whole backend from scratch</h1>")
    total_pieces = 0
    for c in CH:
        for relpath, _ in c[3]:
            with open(os.path.join(SRC, relpath), encoding="utf-8") as fh:
                t = fh.read()
            total_pieces += len(split_go(t[:-1] if t.endswith("\n") else t))
    out.append('<div class="intro"><p>This is a <strong>shadow-a-developer runbook</strong>. You will not paste finished files &mdash; you will build each one the way a developer actually does: <strong>one declaration at a time</strong>, in the order they\'d add them, understanding each piece before the next. Every file is grown from its <code>package</code> line and first import through its last method. Nothing is elided: the pieces of a file, concatenated, <em>are</em> that file, byte for byte (the generator asserts this). By the end you\'ll have typed the entire functional backend &mdash; <strong>%d files</strong> in <strong>%d pieces</strong> &mdash; and compiled the binary that replaces the current one.</p>'
        '<p class="why" style="margin-top:10px">Two orders are in play. <strong>Between files</strong>: dependency order &mdash; leaf packages first (utils, engine), edges last (web, wiring), so every import already exists when you need it. <strong>Within a file</strong>: construction order &mdash; the type, then its constructor, then its methods &mdash; exactly how you\'d flesh it out in an editor. Each chapter ends with a <code>go build</code> that really passes.</p></div>' % (total_files, total_pieces))
    # TOC
    out.append('<div class="toc">')
    for anchor, title, *_ in CH:
        n = len([c for c in CH if c[0]==anchor][0][3])
        tag = ' <span class="tag">%d</span>' % n if n else ''
        out.append('<a href="#%s">%s%s</a>' % (anchor, esc(title), tag))
    out.append('<a href="#run">Chapter 24 &mdash; Build, run &amp; replace the binary</a>')
    out.append('</div>')
    # chapters
    for anchor, title, why, files, chk, extra_top in CH:
        out.append('<h2 id="%s">%s</h2>' % (anchor, esc(title)))
        if why:
            out.append('<div class="why">%s</div>' % why)
        if extra_top:
            out.append(extra_top)
        for relpath, note in files:
            out.append(file_block(relpath, note))
        if chk:
            out.append('<div class="chk">%s</div>' % chk)
    # final run chapter
    out.append(RUN_CH)
    out.append('<div style="text-align:center;margin:48px 0 12px;opacity:.7;font-size:.9em">&mdash; end of runbook: %d files, every line included &mdash;</div>' % total_files)
    out.append("</div>\n</body>\n</html>")
    return "\n".join(out)

RUN_CH = """
<h2 id="run">Chapter 24 — Build, run &amp; replace the binary</h2>
<div class="why">Code you can't run is a guess. This chapter turns the source you just typed into the binary that replaces the current backend.</div>
<div class="step"><span class="step-tag">STEP 24.1</span> Resolve dependencies and build.
<pre><code># from the project root (where go.mod lives)
go mod tidy                          # pulls every dependency the chapters introduced
CGO_ENABLED=0 go build -o exchange . # pure-Go build (the SQLite driver is CGO-free)</code></pre>
<div class="why">Why <code>CGO_ENABLED=0</code>? The <code>ncruces/go-sqlite3</code> driver is a pure-Go/WASM build, so the binary needs no C toolchain and is fully static &mdash; it drops onto any Linux box and just runs. That's the whole reason this exchange is one self-contained file.</div>
</div>
<div class="step"><span class="step-tag">STEP 24.2</span> Cross-compile for the server (from Windows/Mac) and swap it in.
<pre><code># build a linux/amd64 binary from any OS
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o exchange-linux .
# copy up, stop the old one, replace, restart (systemd example)
scp exchange-linux user@server:/opt/exchange/exchange.new
ssh user@server 'sudo systemctl stop exchange \\
   &amp;&amp; mv /opt/exchange/exchange.new /opt/exchange/exchange \\
   &amp;&amp; chmod +x /opt/exchange/exchange \\
   &amp;&amp; sudo systemctl start exchange'</code></pre>
<div class="why">Replacing the binary is exactly this: build the same source for the server's OS/arch, stop the service so the file isn't busy, move the new file over the old, restart. The database file (<code>exg.db</code>) is untouched, so balances/orders survive the swap &mdash; and <code>recoverOrderBook</code> (Chapter 7) rebuilds the in-memory book on boot.</div>
</div>
<div class="step"><span class="step-tag">STEP 24.3</span> Prove it works.
<pre><code>BASE=http://localhost:3000/api/v1
curl -s -XPOST $BASE/users/register -d '{"username":"alice","password":"pw"}'
TOKEN=$(curl -s -XPOST $BASE/users/login -d '{"username":"alice","password":"pw"}' | jq -r .data)
curl -s -XPOST $BASE/orders/BTC-USDT -H "Authorization: $TOKEN" \\
     -d '{"side":"BID","type":"LIMIT","price":65000,"size":0.01}'
curl -s $BASE/balances -H "Authorization: $TOKEN"          # available -> locked -> settled
curl -s $BASE/orderbooks/BTC-USDT/snapshot                 # public book, no token</code></pre>
</div>
<div class="chk">You built every one of the functional files from this page alone, compiled the binary, and swapped it in with the database intact. That is the whole backend &mdash; and you know why every folder, import and lock is there.</div>
"""

if __name__ == "__main__":
    result = build()
    with open(OUT, "w", encoding="utf-8", newline="\n") as f:
        f.write(result)
    print("wrote", OUT)
    print("bytes:", len(result))
    print("files embedded:", sum(len(c[3]) for c in CH))
