// src/utils/auth.js
export const authUtils = {
    // 設定認證資料
    setAuthData(token, username) {
        localStorage.setItem('cryptoex_token', token)
        localStorage.setItem('cryptoex_username', username)
    },

    // 獲取 token
    getToken() {
        return localStorage.getItem('cryptoex_token')
    },

    // 獲取用戶名
    getUsername() {
        return localStorage.getItem('cryptoex_username')
    },

    // 檢查是否已登入
    isAuthenticated() {
        return !!this.getToken()
    },

    // 清除認證資料
    clearAuthData() {
        localStorage.removeItem('cryptoex_token')
        localStorage.removeItem('cryptoex_username')
        localStorage.removeItem('user_profile')
    },

    // 格式化日期
    formatDate(timestamp) {
        return new Date(timestamp).toLocaleString()
    },

    // 格式化費率
    formatFeeRate(rate) {
        return (rate * 100).toFixed(3) + '%'
    },
    setUserProfile(userProfile) {
        localStorage.setItem('user_profile', JSON.stringify(userProfile));
    },

    getUserProfile() {
        const data = localStorage.getItem('user_profile');
        console.log('Retrieving user profile from localStorage:', data)
        try {
            return data ? JSON.parse(data) : null;
        } catch (e) {
            console.error('Failed to parse user profile from localStorage', e);
            return null;
        }
    }
}