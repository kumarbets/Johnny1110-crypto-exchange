// src/services/apiService.js
import axios from 'axios'
import {authUtils} from "@/services/auth";

//const BASE_URL = 'http://localhost:8080'
const BASE_URL = 'http://34.80.224.23:8080'

const apiClient = axios.create({
    baseURL: BASE_URL,
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
})

export const marketAPI = {
    // 獲取所有市場數據
    getAllMarkets() {
        return apiClient.get('/api/v1/markets')
    },

    // 獲取特定市場數據
    getMarketData(market) {
        return apiClient.get(`/api/v1/market/${market}`)
    }
}

export const userAPI = {
    // 用戶註冊
    register(username, password) {
        return apiClient.post('/api/v1/users/register', {
            username,
            password
        })
    },

    // 用戶登入
    login(username, password) {
        return apiClient.post('/api/v1/users/login', {
            username,
            password
        })
    },

    // 用戶登出
    logout() {
        return apiClient.post('/api/v1/users/logout')
    },

    // 獲取用戶資料
    getProfile() {
        return apiClient.get('/api/v1/users/profile')
    }
}

// 新增：錢包 API
export const walletAPI = {
    // 獲取用戶餘額資訊
    getBalances() {
        return apiClient.get('/api/v1/balances')
    }
}

// orderBook
export const orderBooksAPI = {
    getOrderBook(market) {
        return apiClient.get(`/api/v1/orderbooks/${market}/snapshot`)
    }
}

//ohlcv
export const ohlcvAPI = {
    getOhlcvHistory(market, interval) {
        return apiClient.get(`/api/v1/markets/${market}/ohlcv-history/${interval}`)
    }
}

// 新增：訂單 API
export const ordersAPI = {
    /**
     * 下限價訂單
     * @param {string} market - 交易對，例如 'ETH-USDT'
     * @param {Object} orderData - 訂單資料
     * @param {number} orderData.side - 0=買單(Bid), 1=賣單(Ask)
     * @param {number} orderData.order_type - 0=限價單, 1=市價單
     * @param {number} orderData.mode - 0=掛單方(Maker), 1=吃單方(Taker)
     * @param {number} orderData.price - 限價 (限價單必填)
     * @param {number} orderData.size - 數量 (限價單必填)
     * @param {number} orderData.quote_amount - 總金額 (市價單必填)
     */
    placeOrder(market, orderData) {
        return apiClient.post(`/api/v1/orders/${market}`, orderData)
    },

    /**
     * 下限價買單
     * @param {string} market - 交易對
     * @param {number} price - 限價
     * @param {number} size - 數量
     * @param {number} mode - 0=Maker, 1=Taker，預設為1(Taker)
     */
    placeLimitBuyOrder(market, price, size, mode = 1) {
        return this.placeOrder(market, {
            side: 0,
            order_type: 0,
            mode: mode,
            price: price,
            size: size
        })
    },

    /**
     * 下限價賣單
     * @param {string} market - 交易對
     * @param {number} price - 限價
     * @param {number} size - 數量
     * @param {number} mode - 0=Maker, 1=Taker，預設為1(Taker)
     */
    placeLimitSellOrder(market, price, size, mode = 1) {
        return this.placeOrder(market, {
            side: 1,
            order_type: 0,
            mode: mode,
            price: price,
            size: size
        })
    },

    /**
     * 下市價買單
     * @param {string} market - 交易對
     * @param {number} quoteAmount - 總成本金額
     */
    placeMarketBuyOrder(market, quoteAmount) {
        return this.placeOrder(market, {
            side: 0,
            order_type: 1,
            quote_amount: quoteAmount
        })
    },

    /**
     * 下市價賣單
     * @param {string} market - 交易對
     * @param {number} size - 賣出數量
     */
    placeMarketSellOrder(market, size) {
        return this.placeOrder(market, {
            side: 1,
            order_type: 1,
            size: size
        })
    },

    /**
     * 取消訂單
     * @param {string} orderId - 訂單ID
     */
    cancelOrder(orderId) {
        return apiClient.delete(`/api/v1/orders/${orderId}`)
    },

    /**
     * 查詢訂單
     * @param {Object} params - 查詢參數
     * @param {string} params.market - 交易對 (可選)
     * @param {number} params.side - 0=買單, 1=賣單 (可選)
     * @param {string} params.type - 'OPENING'=未完成, 'CLOSED'=已完成 (必填)
     * @param {number} params.page_size - 每頁數量 (可選，預設10)
     * @param {number} params.current_page - 當前頁數 (可選，預設1)
     */
    getOrders(params) {
        return apiClient.get('/api/v1/orders', { params })
    },

    /**
     * 獲取未完成訂單
     * @param {string} market - 交易對 (可選)
     * @param {number} side - 0=買單, 1=賣單 (可選)
     * @param {number} pageSize - 每頁數量
     * @param {number} currentPage - 當前頁數
     */
    getOpenOrders(market = null, side = null, pageSize = 10, currentPage = 1) {
        const params = {
            type: 'OPENING',
            page_size: pageSize,
            current_page: currentPage
        }
        if (market) params.market = market
        if (side !== null) params.side = side

        return this.getOrders(params)
    },

    /**
     * 獲取已完成訂單
     * @param {string} market - 交易對 (可選)
     * @param {number} side - 0=買單, 1=賣單 (可選)
     * @param {number} pageSize - 每頁數量
     * @param {number} currentPage - 當前頁數
     */
    getClosedOrders(market = null, side = null, pageSize = 10, currentPage = 1) {
        const params = {
            type: 'CLOSED',
            page_size: pageSize,
            current_page: currentPage
        }
        if (market) params.market = market
        if (side !== null) params.side = side

        return this.getOrders(params)
    }
}

// 錯誤處理攔截器
apiClient.interceptors.response.use(
    response => response,
    error => {
        console.error('API Error:', error)
        return Promise.reject(error)
    }
)

// 請求攔截器 - 自動添加 token
apiClient.interceptors.request.use(
    config => {
        const token = authUtils.getToken()
        if (token) {
            config.headers.Authorization = token
        }
        return config
    },
    error => {
        return Promise.reject(error)
    }
)