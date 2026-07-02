<!--<template>-->
<!--  <div class="crypto-market">-->
<!--    <div class="container">-->
<!--      <div class="title-bar">-->
<!--        <span>CryptoEx 2000 - Cryptocurrency Exchange</span>-->
<!--        <div>-->
<!--          <button>_</button>-->
<!--          <button>□</button>-->
<!--          <button>X</button>-->
<!--        </div>-->
<!--      </div>-->

<!--      <div class="nav-bar">-->
<!--        <router-link to="/">Home</router-link>-->
<!--        <a href="#" @click.prevent="activeTab = 'trade'" :class="{ active: activeTab === 'trade' }">Trade</a>-->
<!--        <router-link to="/balances">Wallet</router-link>-->
<!--        <a href="#" @click.prevent="activeTab = 'account'" :class="{ active: activeTab === 'account' }">Account</a>-->
<!--        <a href="#" @click.prevent="isLoggedIn ? handleLogout() : showLogin()" class="auth-link">-->
<!--          {{ isLoggedIn ? 'Logout' : 'Login' }}-->
<!--        </a>-->
<!--      </div>-->

<!--      &lt;!&ndash; User Profile Section &ndash;&gt;-->
<!--      <UserProfile-->
<!--          v-if="isLoggedIn && userProfile"-->
<!--          :user="userProfile"-->
<!--          @logout="handleLogout"-->
<!--      />-->

<!--      <div class="loading" v-if="loading">-->
<!--        <div class="loading-text">Loading market data...</div>-->
<!--      </div>-->

<!--      <div class="error" v-if="error">-->
<!--        <div class="error-text">{{ error }}</div>-->
<!--        <button @click="fetchMarketData" class="retry-btn">Retry</button>-->
<!--      </div>-->

<!--      <template v-if="!loading && !error">-->
<!--        <CommandWindow :selected-market="selectedMarket" />-->

<!--        <MarketTable-->
<!--            :markets="markets"-->
<!--            @market-selected="onMarketSelected"-->
<!--        />-->



<!--      </template>-->

<!--      <div class="footer">-->
<!--        CryptoEx 2000 © 2025 - All Rights Reserved |-->
<!--        Last Updated: {{ lastUpdated }}-->
<!--      </div>-->
<!--    </div>-->

<!--    &lt;!&ndash; Login Modal &ndash;&gt;-->
<!--    <LoginModal-->
<!--        :visible="showLoginModal"-->
<!--        @close="showLoginModal = false"-->
<!--        @login-success="onLoginSuccess"-->
<!--    />-->
<!--  </div>-->
<!--</template>-->

<!--<script>-->
<!--import { marketAPI, userAPI } from '@/services/apiService'-->
<!--import MarketTable from './MarketTable.vue'-->
<!--import CommandWindow from './CommandWindow.vue'-->
<!--import LoginModal from './LoginModal.vue'-->
<!--import UserProfile from './UserProfile.vue'-->
<!--import {authUtils} from "@/services/auth";-->

<!--export default {-->
<!--  name: 'CryptoMarket',-->
<!--  components: {-->
<!--    MarketTable,-->
<!--    CommandWindow,-->
<!--    LoginModal,-->
<!--    UserProfile-->
<!--  },-->
<!--  data() {-->
<!--    return {-->
<!--      markets: [],-->
<!--      selectedMarket: null,-->
<!--      loading: false,-->
<!--      error: null,-->
<!--      activeTab: 'home',-->
<!--      lastUpdated: null,-->
<!--      refreshInterval: null,-->
<!--      // 認證相關-->
<!--      isLoggedIn: false,-->
<!--      userProfile: null,-->
<!--      showLoginModal: false-->
<!--    }-->
<!--  },-->
<!--  async mounted() {-->
<!--    await this.checkAuthStatus()-->
<!--    await this.fetchMarketData()-->
<!--    this.startAutoRefresh()-->
<!--  },-->
<!--  beforeUnmount() {-->
<!--    if (this.refreshInterval) {-->
<!--      clearInterval(this.refreshInterval)-->
<!--    }-->
<!--  },-->
<!--  methods: {-->
<!--    // 檢查登入狀態-->
<!--    async checkAuthStatus() {-->
<!--      const token = authUtils.getToken()-->
<!--      const username = authUtils.getUsername()-->

<!--      if (token && username) {-->
<!--        try {-->
<!--          const response = await userAPI.getProfile()-->
<!--          if (response.data.code === '0000000') {-->
<!--            this.isLoggedIn = true-->
<!--            this.userProfile = response.data.data-->
<!--          } else {-->
<!--            // Token 無效，清除本地儲存-->
<!--            this.clearAuthData()-->
<!--          }-->
<!--        } catch (error) {-->
<!--          console.error('Auth check failed:', error)-->
<!--          this.clearAuthData()-->
<!--        }-->
<!--      }-->
<!--    },-->

<!--    // 清除認證資料-->
<!--    clearAuthData() {-->
<!--      authUtils.clearAuthData()-->
<!--      this.isLoggedIn = false-->
<!--      this.userProfile = null-->
<!--    },-->

<!--    // 顯示登入彈窗-->
<!--    showLogin() {-->
<!--      this.showLoginModal = true-->
<!--    },-->

<!--    // 登入成功處理-->
<!--    async onLoginSuccess(userData) {-->
<!--      this.isLoggedIn = true-->

<!--      console.log("login user:", userData)-->

<!--      // 獲取用戶詳細資料-->
<!--      try {-->
<!--        const response = await userAPI.getProfile()-->
<!--        if (response.data.code === '0000000') {-->
<!--          this.userProfile = response.data.data-->
<!--        }-->
<!--      } catch (error) {-->
<!--        console.error('Failed to fetch user profile:', error)-->
<!--      }-->
<!--    },-->

<!--    // 登出處理-->
<!--    async handleLogout() {-->
<!--      try {-->
<!--        if (this.isLoggedIn) {-->
<!--          await userAPI.logout()-->
<!--        }-->
<!--      } catch (error) {-->
<!--        console.error('Logout error:', error)-->
<!--      } finally {-->
<!--        this.clearAuthData()-->
<!--      }-->
<!--    },-->
<!--    async fetchMarketData() {-->
<!--      this.loading = true-->
<!--      this.error = null-->

<!--      try {-->
<!--        const response = await marketAPI.getAllMarkets()-->

<!--        if (response.data.code === '0000000') {-->
<!--          this.markets = response.data.data-->
<!--          this.lastUpdated = new Date().toLocaleTimeString()-->
<!--        } else {-->
<!--          throw new Error(response.data.message || 'Failed to fetch market data')-->
<!--        }-->
<!--      } catch (error) {-->
<!--        this.error = error.response?.data?.message || error.message || 'Network error occurred'-->
<!--        console.error('Error fetching market data:', error)-->
<!--      } finally {-->
<!--        this.loading = false-->
<!--      }-->
<!--    },-->

<!--    async onMarketSelected(market) {-->
<!--      this.selectedMarket = market-->

<!--      try {-->
<!--        const response = await marketAPI.getMarketData(market.market_name)-->
<!--        if (response.data.code === '0000000') {-->
<!--          // 更新選中市場的詳細數據-->
<!--          this.selectedMarket = response.data.data-->
<!--        }-->
<!--      } catch (error) {-->
<!--        console.error('Error fetching market details:', error)-->
<!--      }-->
<!--    },-->

<!--    startAutoRefresh() {-->
<!--      this.refreshInterval = setInterval(() => {-->
<!--        this.fetchMarketData()-->
<!--      }, 30000) // 每30秒更新一次-->
<!--    }-->
<!--  }-->
<!--}-->
<!--</script>-->

<!--<style scoped>-->
<!--@import url('https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap');-->

<!--.crypto-market {-->
<!--  background: linear-gradient(180deg, #ff66cc, #9900cc);-->
<!--  font-family: 'Press Start 2P', cursive;-->
<!--  color: #ffffff;-->
<!--  margin: 0;-->
<!--  padding: 20px;-->
<!--  min-height: 100vh;-->
<!--  background-image: url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAYAAACp8Z5+AAAAG0lEQVR4AWMAAv+BAgICAgICiAECAwICAgICAgICBtgBc4AAAAASUVORK5CYII=');-->
<!--  background-repeat: repeat;-->
<!--}-->

<!--.container {-->
<!--  background: rgba(51, 0, 51, 0.8);-->
<!--  border: 3px solid #ff99ff;-->
<!--  width: 800px;-->
<!--  margin: 0 auto;-->
<!--  padding: 15px;-->
<!--  box-shadow: 0 0 10px #ff66cc, 0 0 20px #9900cc;-->
<!--  border-radius: 5px;-->
<!--}-->

<!--.title-bar {-->
<!--  background: linear-gradient(90deg, #ff33cc, #cc00ff);-->
<!--  color: #ffffff;-->
<!--  padding: 8px;-->
<!--  font-size: 12px;-->
<!--  display: flex;-->
<!--  justify-content: space-between;-->
<!--  align-items: center;-->
<!--  border: 2px solid #ff99ff;-->
<!--  text-shadow: 1px 1px 2px #330033;-->
<!--}-->

<!--.title-bar button {-->
<!--  background: #ff66cc;-->
<!--  border: 2px solid #ff99ff;-->
<!--  padding: 3px 10px;-->
<!--  cursor: pointer;-->
<!--  font-family: 'Press Start 2P', cursive;-->
<!--  font-size: 10px;-->
<!--  color: #ffffff;-->
<!--  text-shadow: 1px 1px #330033;-->
<!--  transition: all 0.2s;-->
<!--}-->

<!--.title-bar button:hover {-->
<!--  background: #cc00ff;-->
<!--  box-shadow: 0 0 5px #ff66cc;-->
<!--}-->

<!--.nav-bar {-->
<!--  background: rgba(51, 0, 51, 0.8);-->
<!--  border: 2px solid #ff99ff;-->
<!--  padding: 8px;-->
<!--  margin-bottom: 15px;-->
<!--  box-shadow: 0 0 8px #ff66cc;-->
<!--}-->

<!--.nav-bar a {-->
<!--  margin-right: 12px;-->
<!--  color: #ffccff;-->
<!--  text-decoration: none;-->
<!--  font-size: 10px;-->
<!--  text-shadow: 1px 1px #330033;-->
<!--  cursor: pointer;-->
<!--}-->

<!--.nav-bar a:hover,-->
<!--.nav-bar a.active {-->
<!--  color: #ff66cc;-->
<!--  text-shadow: 0 0 5px #ff66cc;-->
<!--}-->

<!--.nav-bar a.auth-link {-->
<!--  color: #ff66cc;-->
<!--  font-weight: bold;-->
<!--}-->

<!--.nav-bar a.auth-link:hover {-->
<!--  color: #ffffff;-->
<!--  text-shadow: 0 0 8px #ff66cc;-->
<!--}-->

<!--.loading {-->
<!--  text-align: center;-->
<!--  padding: 50px;-->
<!--}-->

<!--.loading-text {-->
<!--  font-size: 12px;-->
<!--  color: #ff66cc;-->
<!--  animation: pulse 1.5s infinite;-->
<!--}-->

<!--.error {-->
<!--  text-align: center;-->
<!--  padding: 20px;-->
<!--}-->

<!--.error-text {-->
<!--  color: #ff3366;-->
<!--  font-size: 10px;-->
<!--  margin-bottom: 10px;-->
<!--}-->

<!--.retry-btn {-->
<!--  background: #ff66cc;-->
<!--  border: 2px solid #ff99ff;-->
<!--  padding: 5px 15px;-->
<!--  cursor: pointer;-->
<!--  font-family: 'Press Start 2P', cursive;-->
<!--  font-size: 8px;-->
<!--  color: #ffffff;-->
<!--}-->

<!--.retry-btn:hover {-->
<!--  background: #cc00ff;-->
<!--}-->

<!--.footer {-->
<!--  text-align: center;-->
<!--  font-size: 8px;-->
<!--  color: #ffccff;-->
<!--  margin-top: 15px;-->
<!--  text-shadow: 1px 1px #330033;-->
<!--}-->

<!--@keyframes pulse {-->
<!--  0%, 100% { opacity: 1; }-->
<!--  50% { opacity: 0.5; }-->
<!--}-->
<!--</style>-->