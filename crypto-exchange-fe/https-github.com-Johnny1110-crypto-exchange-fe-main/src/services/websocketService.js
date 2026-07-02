// websocketService.js
class WebSocketService {
    constructor() {
        this.ws = null
        this.isConnected = false
        this.reconnectAttempts = 0
        this.maxReconnectAttempts = 5
        this.reconnectDelay = 3000 // 3 seconds
        this.subscribers = new Map() // 儲存訂閱回調函數
        this.messageQueue = [] // 連線前的訊息佇列

        // WebSocket 端點配置
        this.endpoints = {
            local: 'ws://localhost:8081/ws',
            remote: 'ws://34.80.224.23:8081/ws'
        }
        this.currentEndpoint = this.endpoints.remote // 預設使用 remote
    }

    /**
     * 建立 WebSocket 連線
     * @param {string} endpoint - 'local' 或 'remote'
     */
    connect(endpoint = 'remote') {
        if (this.ws && this.isConnected) {
            console.log('WebSocket 已連線')
            return Promise.resolve()
        }

        this.currentEndpoint = this.endpoints[endpoint] || this.endpoints.remote

        return new Promise((resolve, reject) => {
            try {
                this.ws = new WebSocket(this.currentEndpoint)

                this.ws.onopen = () => {
                    console.log('WebSocket 連線成功:', this.currentEndpoint)
                    this.isConnected = true
                    this.reconnectAttempts = 0

                    // 處理佇列中的訊息
                    this.processMessageQueue()

                    resolve()
                }

                this.ws.onmessage = (event) => {
                    try {
                        const message = JSON.parse(event.data)
                        this.handleMessage(message)
                    } catch (error) {
                        console.error('解析 WebSocket 訊息失敗:', error, event.data)
                    }
                }

                this.ws.onclose = (event) => {
                    console.log('WebSocket 連線關閉:', event.code, event.reason)
                    this.isConnected = false

                    // 如果不是主動關閉，嘗試重連
                    if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
                        this.attemptReconnect()
                    }
                }

                this.ws.onerror = (error) => {
                    console.error('WebSocket 連線錯誤:', error)
                    this.isConnected = false
                    reject(error)
                }

                // 連線超時處理
                setTimeout(() => {
                    if (!this.isConnected) {
                        reject(new Error('WebSocket 連線超時'))
                    }
                }, 10000) // 10 秒超時

            } catch (error) {
                console.error('建立 WebSocket 連線失敗:', error)
                reject(error)
            }
        })
    }

    /**
     * 嘗試重新連線
     */
    attemptReconnect() {
        this.reconnectAttempts++
        console.log(`嘗試重新連線 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`)

        setTimeout(() => {
            this.connect().catch(error => {
                console.error('重新連線失敗:', error)
                if (this.reconnectAttempts >= this.maxReconnectAttempts) {
                    console.error('已達到最大重連次數，停止重連')
                }
            })
        }, this.reconnectDelay * this.reconnectAttempts)
    }

    /**
     * 處理訊息佇列
     */
    processMessageQueue() {
        while (this.messageQueue.length > 0) {
            const message = this.messageQueue.shift()
            this.sendMessage(message)
        }
    }

    /**
     * 發送訊息
     * @param {Object} message - 要發送的訊息
     */
    sendMessage(message) {
        if (!this.isConnected || !this.ws) {
            console.log('WebSocket 未連線，訊息加入佇列:', message)
            this.messageQueue.push(message)
            return
        }

        try {
            this.ws.send(JSON.stringify(message))
            console.log('發送 WebSocket 訊息:', message)
        } catch (error) {
            console.error('發送 WebSocket 訊息失敗:', error)
        }
    }

    /**
     * 處理接收到的訊息
     * @param {Object} message - 接收到的訊息
     */
    handleMessage(message) {
        console.log('收到 WebSocket 訊息:', message)

        const { channel, data, timestamp } = message

        if (channel === 'ohlcv' && data) {
            this.handleOhlcvMessage(data, timestamp)
        }
    }

    /**
     * 處理 OHLCV 訊息
     * @param {Object} data - OHLCV 數據
     * @param {number} timestamp - 時間戳
     */
    handleOhlcvMessage(data, timestamp) {
        if (data.s !== 'ok') {
            console.warn('OHLCV 數據狀態異常:', data.s)
            return
        }

        // 轉換數據格式
        const transformedData = this.transformOhlcvData(data)

        // 通知所有 OHLCV 訂閱者
        const ohlcvSubscribers = this.subscribers.get('ohlcv') || []
        ohlcvSubscribers.forEach(callback => {
            try {
                callback(transformedData, timestamp)
            } catch (error) {
                console.error('OHLCV 回調函數執行失敗:', error)
            }
        })
    }

    /**
     * 轉換 OHLCV 數據格式
     * @param {Object} data - 原始 OHLCV 數據
     * @returns {Object} 轉換後的數據
     */
    transformOhlcvData(data) {
        const candleData = []
        const volumeData = []

        for (let i = 0; i < data.t.length; i++) {
            const time = data.t[i]
            const open = data.o[i]
            const high = data.h[i]
            const low = data.l[i]
            const close = data.c[i]
            const volume = Math.abs(data.v[i])

            candleData.push({ time, open, high, low, close })
            volumeData.push({
                time,
                value: volume,
                color: close >= open ? '#26a69a' : '#ef5350'
            })
        }

        return { candleData, volumeData }
    }

    /**
     * 訂閱 OHLCV 數據
     * @param {string} symbol - 交易對符號 (e.g., 'ETH-USDT')
     * @param {string} interval - 時間間隔 ('15m', '1h', '1d', '1w')
     * @param {Function} callback - 接收數據的回調函數
     */
    subscribeOhlcv(symbol, interval, callback) {
        // 註冊回調函數
        if (!this.subscribers.has('ohlcv')) {
            this.subscribers.set('ohlcv', [])
        }
        this.subscribers.get('ohlcv').push(callback)

        // 發送訂閱訊息
        const subscribeMessage = {
            action: 'subscribe',
            channel: 'ohlcv',
            params: {
                symbol,
                interval
            }
        }

        this.sendMessage(subscribeMessage)
        console.log(`訂閱 OHLCV: ${symbol} ${interval}`)

        // 返回取消訂閱函數
        return () => this.unsubscribeOhlcv(symbol, interval, callback)
    }

    /**
     * 取消訂閱 OHLCV 數據
     * @param {string} symbol - 交易對符號
     * @param {string} interval - 時間間隔
     * @param {Function} callback - 要移除的回調函數
     */
    unsubscribeOhlcv(symbol, interval, callback) {
        // 移除回調函數
        const ohlcvSubscribers = this.subscribers.get('ohlcv') || []
        const index = ohlcvSubscribers.indexOf(callback)
        if (index > -1) {
            ohlcvSubscribers.splice(index, 1)
        }

        // 發送取消訂閱訊息
        const unsubscribeMessage = {
            action: 'unsubscribe',
            channel: 'ohlcv',
            params: {
                symbol,
                interval
            }
        }

        this.sendMessage(unsubscribeMessage)
        console.log(`取消訂閱 OHLCV: ${symbol} ${interval}`)
    }

    /**
     * 訂閱 Orderbook 數據
     * @param {string} symbol - 交易對符號
     * @param {Function} callback - 接收數據的回調函數
     */
    subscribeOrderbook(symbol, callback) {
        // 註冊回調函數
        if (!this.subscribers.has('orderbook')) {
            this.subscribers.set('orderbook', [])
        }
        this.subscribers.get('orderbook').push(callback)

        // 發送訂閱訊息
        const subscribeMessage = {
            action: 'subscribe',
            channel: 'orderbook',
            params: { symbol }
        }

        this.sendMessage(subscribeMessage)
        console.log(`訂閱 Orderbook: ${symbol}`)

        // 返回取消訂閱函數
        return () => this.unsubscribeOrderbook(symbol, callback)
    }

    /**
     * 取消訂閱 Orderbook 數據
     * @param {string} symbol - 交易對符號
     * @param {Function} callback - 要移除的回調函數
     */
    unsubscribeOrderbook(symbol, callback) {
        const orderbookSubscribers = this.subscribers.get('orderbook') || []
        const index = orderbookSubscribers.indexOf(callback)
        if (index > -1) {
            orderbookSubscribers.splice(index, 1)
        }

        const unsubscribeMessage = {
            action: 'unsubscribe',
            channel: 'orderbook',
            params: { symbol }
        }

        this.sendMessage(unsubscribeMessage)
        console.log(`取消訂閱 Orderbook: ${symbol}`)
    }

    /**
     * 訂閱市場數據
     * @param {Function} callback - 接收數據的回調函數
     */
    subscribeMarkets(callback) {
        // 註冊回調函數
        if (!this.subscribers.has('markets')) {
            this.subscribers.set('markets', [])
        }
        this.subscribers.get('markets').push(callback)

        // 發送訂閱訊息
        const subscribeMessage = {
            action: 'subscribe',
            channel: 'markets'
        }

        this.sendMessage(subscribeMessage)
        console.log('訂閱 Markets')

        // 返回取消訂閱函數
        return () => this.unsubscribeMarkets(callback)
    }

    /**
     * 取消訂閱市場數據
     * @param {Function} callback - 要移除的回調函數
     */
    unsubscribeMarkets(callback) {
        const marketsSubscribers = this.subscribers.get('markets') || []
        const index = marketsSubscribers.indexOf(callback)
        if (index > -1) {
            marketsSubscribers.splice(index, 1)
        }

        const unsubscribeMessage = {
            action: 'unsubscribe',
            channel: 'markets'
        }

        this.sendMessage(unsubscribeMessage)
        console.log('取消訂閱 Markets')
    }

    /**
     * 關閉 WebSocket 連線
     */
    disconnect() {
        if (this.ws) {
            console.log('主動關閉 WebSocket 連線')
            this.ws.close(1000, '主動斷線')
            this.ws = null
            this.isConnected = false
            this.subscribers.clear()
            this.messageQueue = []
        }
    }

    /**
     * 獲取連線狀態
     * @returns {boolean} 是否已連線
     */
    getConnectionState() {
        return {
            isConnected: this.isConnected,
            reconnectAttempts: this.reconnectAttempts,
            endpoint: this.currentEndpoint,
            subscribersCount: Array.from(this.subscribers.values()).reduce((sum, arr) => sum + arr.length, 0)
        }
    }
}

// 建立單例實例
const websocketService = new WebSocketService()

// 導出服務實例和類別
export default websocketService
export { WebSocketService }