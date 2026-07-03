package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	url := flag.String("url", "ws://localhost:8081/ws", "")
	token := flag.String("token", "", "")
	flag.Parse()
	c, _, err := websocket.DefaultDialer.Dial(*url, nil)
	if err != nil {
		fmt.Println("DIAL ERR:", err)
		return
	}
	defer c.Close()
	c.WriteJSON(map[string]interface{}{"action": "subscribe", "channel": "orderbook", "params": map[string]string{"market": "BTC-USDT"}})
	c.WriteJSON(map[string]interface{}{"action": "subscribe", "channel": "user_data", "params": map[string]string{"token": *token, "market": "BTC-USDT"}})
	c.WriteJSON(map[string]interface{}{"action": "subscribe", "channel": "sysstats"})
	gotOB, gotUser, gotSys := false, false, false
	c.SetReadDeadline(time.Now().Add(6 * time.Second))
	for i := 0; i < 20; i++ {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}
		var m map[string]interface{}
		json.Unmarshal(msg, &m)
		ch, _ := m["channel"].(string)
		if ch == "orderbook" && !gotOB {
			gotOB = true
			d, _ := json.Marshal(m["data"])
			fmt.Println("ORDERBOOK OK:", string(d)[:min(160, len(string(d)))])
		}
		if ch == "user_data" && !gotUser {
			gotUser = true
			d, _ := json.Marshal(m["data"])
			s := string(d)
			fmt.Println("USER_DATA OK:", s[:min(220, len(s))])
		}
		if ch == "sysstats" && !gotSys { gotSys = true; d,_ := json.Marshal(m["data"]); fmt.Println("SYSSTATS OK:", string(d)) }
		if gotOB && gotUser && gotSys {
			break
		}
	}
	fmt.Printf("RESULT orderbook=%v user_data=%v\n", gotOB, gotUser)
}
func min(a, b int) int { if a < b { return a }; return b }
