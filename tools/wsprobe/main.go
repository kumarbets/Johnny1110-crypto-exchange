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
	c.WriteJSON(map[string]interface{}{"action": "subscribe", "channel": "ohlcv", "params": map[string]string{"symbol": "BTC-USDT", "interval": "1m"}})
	gotOB, gotUser, gotSys, gotOhlcv := false, false, false, false
	c.SetReadDeadline(time.Now().Add(12 * time.Second))
	for i := 0; i < 60; i++ {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}
		var m map[string]interface{}
		json.Unmarshal(msg, &m)
		ch, _ := m["channel"].(string)
		d, _ := json.Marshal(m["data"])
		if ch == "orderbook" && !gotOB {
			gotOB = true
			fmt.Println("ORDERBOOK OK:", string(d)[:min(120, len(string(d)))])
		}
		if ch == "user_data" && !gotUser {
			gotUser = true
			fmt.Println("USER_DATA OK:", string(d)[:min(260, len(string(d)))])
		}
		if ch == "sysstats" && !gotSys {
			gotSys = true
			fmt.Println("SYSSTATS OK:", string(d))
		}
		if ch == "ohlcv" && !gotOhlcv {
			gotOhlcv = true
			fmt.Println("OHLCV OK:", string(d)[:min(200, len(string(d)))])
		}
		if gotOB && gotUser && gotSys && gotOhlcv {
			break
		}
	}
	fmt.Printf("RESULT orderbook=%v user_data=%v sysstats=%v ohlcv=%v\n", gotOB, gotUser, gotSys, gotOhlcv)
}
func min(a, b int) int { if a < b { return a }; return b }
