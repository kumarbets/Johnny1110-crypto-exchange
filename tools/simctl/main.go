package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const apiBase = "http://localhost:8080"
const pass = "Entamma!23"

var users = []string{"user1@gmail.com", "user2@gmail.com", "user3@gmail.com", "user4@gmail.com", "user5@gmail.com"}

// simulation run-time tracking (for the "Duration" field)
var (
	durMu     sync.Mutex
	startTime time.Time // zero when stopped
	frozen    float64   // elapsed seconds captured at the last stop
)

// Duration accumulates ALL time the simulation is actually running: Start resumes
// counting from the accumulated total, Stop banks the interval, only Reset zeroes it.
func markStarted() { durMu.Lock(); startTime = time.Now(); durMu.Unlock() } // keep frozen (accumulate)
func markStopped() {
	durMu.Lock()
	if !startTime.IsZero() {
		frozen += time.Since(startTime).Seconds() // bank the running interval
	}
	startTime = time.Time{}
	durMu.Unlock()
}
func markReset() { durMu.Lock(); startTime = time.Time{}; frozen = 0; durMu.Unlock() }
func elapsedSecs() int {
	durMu.Lock()
	defer durMu.Unlock()
	if !startTime.IsZero() {
		return int(frozen + time.Since(startTime).Seconds()) // banked + current interval
	}
	return int(frozen)
}

func login(u string) string {
	body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, u, pass)
	resp, err := http.Post(apiBase+"/api/v1/users/login", "application/json", strings.NewReader(body))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	s := string(b)
	i := strings.Index(s, `"token":"`)
	if i < 0 {
		return ""
	}
	s = s[i+9:]
	j := strings.Index(s, `"`)
	if j < 0 {
		return ""
	}
	return s[:j]
}

func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")
}

func unit(i int) string { return fmt.Sprintf("gen%d", i+1) }

func startSim(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}
	started := 0
	for i, u := range users {
		exec.Command("systemctl", "stop", unit(i)).Run()
		exec.Command("systemctl", "reset-failed", unit(i)).Run()
		tok := login(u)
		if tok == "" {
			continue
		}
		// max capacity: high concurrency, continuous (Restart=always), limit+market mix
		err := exec.Command("systemd-run", "--unit="+unit(i), "--collect",
			"-p", "Restart=always", "-p", "RestartSec=1",
			"/app/loadtest", "-base", apiBase, "-market", "BTC-USDT",
			"-token", tok, "-n", "100000000", "-c", "20", "-mid", "65000", "-mktpct", "20", "-band", "25").Run()
		if err == nil {
			started++
		}
	}
	if started > 0 {
		markStarted()
	}
	fmt.Fprintf(w, `{"status":"started","generators":%d}`, started)
}

func stopSim(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}
	for i := range users {
		exec.Command("systemctl", "stop", unit(i)).Run()
		exec.Command("systemctl", "reset-failed", unit(i)).Run()
	}
	markStopped()
	fmt.Fprint(w, `{"status":"stopped"}`)
}

func resetSim(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}
	// stop the load first, then wipe + re-fund via the backend admin reset
	for i := range users {
		exec.Command("systemctl", "stop", unit(i)).Run()
		exec.Command("systemctl", "reset-failed", unit(i)).Run()
	}
	req, _ := http.NewRequest("POST", apiBase+"/admin/api/v1/reset", nil)
	req.Header.Set("Admin-Token", "frizo")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(w, `{"status":"error","error":%q}`, err.Error())
		return
	}
	resp.Body.Close()
	markReset()
	fmt.Fprint(w, `{"status":"reset"}`)
}

func statusSim(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}
	running := 0
	for i := range users {
		out, _ := exec.Command("systemctl", "is-active", unit(i)).Output()
		if strings.TrimSpace(string(out)) == "active" {
			running++
		}
	}
	fmt.Fprintf(w, `{"running":%v,"generators":%d,"duration":%d}`, running > 0, running, elapsedSecs())
}

func main() {
	http.HandleFunc("/start", startSim)
	http.HandleFunc("/stop", stopSim)
	http.HandleFunc("/reset", resetSim)
	http.HandleFunc("/status", statusSim)
	fmt.Println("sim control on :8091")
	http.ListenAndServe(":8091", nil)
}
