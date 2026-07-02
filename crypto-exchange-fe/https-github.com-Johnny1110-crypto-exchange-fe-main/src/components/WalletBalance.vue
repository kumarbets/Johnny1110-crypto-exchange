<template>
  <div class="wallet-balance">
    <div>
      <!-- 載入中狀態 -->
      <div class="loading" v-if="loading">
        <div class="loading-text">Loading wallet balances...</div>
      </div>

      <!-- 錯誤狀態 -->
      <div class="error" v-if="error">
        <div class="error-text">{{ error }}</div>
        <button @click="fetchBalances" class="retry-btn">Retry</button>
      </div>

      <!-- 餘額資訊 -->
      <template v-if="!loading && !error">
        <!-- 總覽資訊 -->
        <div class="info-section">
          <h3>Wallet Overview</h3>
          <div class="overview-stats">
            <div class="stat-item">
              <span class="stat-label">Total Assets:</span>
              <span class="stat-value">{{ totalAssets }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Non-Zero Balances:</span>
              <span class="stat-value">{{ nonZeroBalances }}</span>
             </div>
            <div class="stat-item">
              <span class="stat-label">Total Valuation:</span>
               <span class="stat-value">{{ totalValuation }}</span>
            </div>
          </div>
        </div>

        <!-- 餘額表格 -->
        <div class="info-section">
          <h3>Asset Balances</h3>
          <div class="balance-controls">
            <label class="checkbox-container">
              <input type="checkbox" v-model="hideZeroBalances">
              <span class="checkmark"></span>
              Hide Zero Balances
            </label>
            <button @click="refreshBalances" class="refresh-btn" :disabled="loading">
              {{ loading ? 'Refreshing...' : 'Refresh' }}
            </button>
          </div>

          <table class="balance-table">
            <thead>
            <tr>
              <th>Asset</th>
              <th>Available</th>
              <th>Locked</th>
              <th>Total</th>
              <th>Valuation</th>
              <th>Status</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="balance in filteredBalances" :key="balance.asset"
                :class="{ 'has-balance': balance.total > 0 }">
              <td class="asset-cell">
                <span class="asset-name">{{ balance.asset }}</span>
              </td>
              <td class="amount-cell">{{ formatAmount(balance.available) }}</td>
              <td class="amount-cell">{{ formatAmount(balance.locked) }}</td>
              <td class="amount-cell total">{{ formatAmount(balance.total) }}</td>
              <td class="amount-cell total">{{ formatAmount(balance.asset_valuation) + '    '+balance.valuation_currency }}</td>
              <td class="status-cell">
                  <span class="status-badge" :class="getStatusClass(balance)">
                    {{ getStatusText(balance) }}
                  </span>
              </td>
            </tr>
            </tbody>
          </table>
        </div>

        <CommandWindow :push-data="apiResponse" />

      </template>
    </div>
  </div>
</template>

<script>
import { authUtils } from '@/services/auth'
import { walletAPI } from '@/services/apiService'
import CommandWindow from "@/components/CommandWindow.vue";

export default {
  name: 'WalletBalance',
  components: {CommandWindow},
  emits: ['navigate', 'logout'],
  data() {
    return {
      apiResponse: {},
      balances: [],
      loading: false,
      error: null,
      lastUpdated: null,
      hideZeroBalances: false,
      refreshInterval: null
    }
  },
  computed: {
    filteredBalances() {
      if (this.hideZeroBalances) {
        return this.balances.filter(balance => balance.total > 0)
      }
      return this.balances
    },
    totalAssets() {
      return this.balances.length
    },
    nonZeroBalances() {
      return this.balances.filter(balance => balance.total > 0).length
    },
    totalValuation() {
      return this.balances.reduce((sum, balance) => sum + parseFloat(balance.asset_valuation || 0), 0).toFixed(8)
    }
  },
  async mounted() {
    await this.fetchBalances()
    this.startAutoRefresh()
  },
  beforeUnmount() {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval)
    }
  },
  methods: {
    async fetchBalances() {
      this.loading = true
      this.error = null

      try {
        // 檢查是否已登入
        if (!authUtils.isAuthenticated()) {
          throw new Error('Please login to view wallet balances')
        }

        const response = await walletAPI.getBalances()

        if (response.data.code === '0000000') {
          this.balances = response.data.data
          this.apiResponse = response.data
          this.lastUpdated = new Date().toLocaleTimeString()
        } else {
          throw new Error(response.data.message || 'Failed to fetch balances')
        }
      } catch (error) {
        this.error = error.response?.data?.message || error.message || 'Network error occurred'
        console.error('Error fetching balances:', error)
      } finally {
        this.loading = false
      }
    },

    async refreshBalances() {
      await this.fetchBalances()
    },

    formatAmount(amount) {
      if (amount === 0) return '0.00000000'
      if (amount < 0.00000001) return '< 0.00000001'
      return parseFloat(amount).toFixed(8)
    },

    getStatusClass(balance) {
      if (balance.total === 0) return 'status-empty'
      if (balance.locked > 0) return 'status-locked'
      return 'status-available'
    },

    getStatusText(balance) {
      if (balance.total === 0) return 'Empty'
      if (balance.locked > 0) return 'Has Open Order'
      return 'Available'
    },

    formatBalancesForCmd() {
      const cmdData = {
        summary: {
          totalAssets: this.totalAssets,
          nonZeroBalances: this.nonZeroBalances,
          timestamp: new Date().toISOString()
        },
        balances: this.balances.map(balance => ({
          asset: balance.asset,
          available: balance.available,
          locked: balance.locked,
          total: balance.total
        }))
      }
      return JSON.stringify(cmdData, null, 2)
    },

    startAutoRefresh() {
      this.refreshInterval = setInterval(() => {
        this.fetchBalances()
      }, 30000) // 每60秒更新一次
    }
  }
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap');

.wallet-balance {
  background: linear-gradient(180deg, #ff66cc, #9900cc);
  font-family: 'Press Start 2P', cursive;
  color: #ffffff;
  margin: 0;
  padding: 20px;
  min-height: 100vh;
  background-image: url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAYAAACp8Z5+AAAAG0lEQVR4AWMAAv+BAgICAgICiAECAwICAgICAgICBtgBc4AAAAASUVORK5CYII=');
  background-repeat: repeat;
}

.container {
  background: rgba(51, 0, 51, 0.8);
  border: 3px solid #ff99ff;
  max-width: 900px;
  margin: 0 auto;
  padding: 15px;
  box-shadow: 0 0 10px #ff66cc, 0 0 20px #9900cc;
  border-radius: 5px;
}

.title-bar {
  background: linear-gradient(90deg, #ff33cc, #cc00ff);
  color: #ffffff;
  padding: 8px;
  font-size: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 2px solid #ff99ff;
  text-shadow: 1px 1px 2px #330033;
}

.title-bar button {
  background: #ff66cc;
  border: 2px solid #ff99ff;
  padding: 3px 10px;
  cursor: pointer;
  font-family: 'Press Start 2P', cursive;
  font-size: 10px;
  color: #ffffff;
  text-shadow: 1px 1px #330033;
  transition: all 0.2s;
}

.title-bar button:hover {
  background: #cc00ff;
  box-shadow: 0 0 5px #ff66cc;
}

.nav-bar {
  background: rgba(51, 0, 51, 0.8);
  border: 2px solid #ff99ff;
  padding: 8px;
  margin-bottom: 15px;
  box-shadow: 0 0 8px #ff66cc;
}

.nav-bar a {
  margin-right: 12px;
  color: #ffccff;
  text-decoration: none;
  font-size: 10px;
  text-shadow: 1px 1px #330033;
  cursor: pointer;
}

.nav-bar a:hover,
.nav-bar a.active {
  color: #ff66cc;
  text-shadow: 0 0 5px #ff66cc;
}

.nav-bar a.auth-link {
  color: #ff66cc;
  font-weight: bold;
}

.info-section {
  padding: 5px;
  margin-bottom: 5px;
}

.info-section h3 {
  font-size: 12px;
  margin: 10px 0;
  border-bottom: 2px solid #ff99ff;
  color: #ffccff;
  text-shadow: 1px 1px #330033;
  padding-bottom: 5px;
}

.overview-stats {
  display: flex;
  gap: 30px;
  margin: 5px 0;
  justify-content: center;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
  width: 60vh
}

.stat-label {
  font-size: 8px;
  color: #ffccff;
}

.stat-value {
  font-size: 12px;
  color: #ff66cc;
  text-shadow: 0 0 5px #ff66cc;
}

.balance-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 15px 0;
}

.checkbox-container {
  display: flex;
  align-items: center;
  cursor: pointer;
  font-size: 8px;
  color: #ffccff;
}

.checkbox-container input {
  margin-right: 8px;
}

.refresh-btn {
  background: #ff66cc;
  border: 2px solid #ff99ff;
  padding: 5px 12px;
  cursor: pointer;
  font-family: 'Press Start 2P', cursive;
  font-size: 8px;
  color: #ffffff;
  transition: all 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background: #cc00ff;
  box-shadow: 0 0 5px #ff66cc;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.balance-table {
  width: 100%;
  border-collapse: collapse;
  background: rgba(255, 255, 255, 0.1);
  border: 2px solid #ff99ff;
  margin: 15px 0;
}

.balance-table th,
.balance-table td {
  border: 1px solid #ff66cc;
  padding: 8px;
  text-align: left;
  font-size: 10px;
  color: #ffffff;
}

.balance-table th {
  background: linear-gradient(90deg, #cc00ff, #ff33cc);
  text-shadow: 1px 1px #330033;
  text-align: center;
}

.balance-table tr.has-balance {
  background: rgba(255, 102, 204, 0.1);
}

.asset-cell {
  font-weight: bold;
  color: #ff66cc;
}

/*.amount-cell {*/
/*  text-align: right;*/
/*  font-family: 'Courier New', monospace;*/
/*}*/

/*.amount-cell.total {*/
/*  font-weight: bold;*/
/*  color: #ffccff;*/
/*}*/

.status-cell {
  text-align: center;
}

.status-badge {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 8px;
  font-weight: bold;
}

.status-badge.status-empty {
  background: #666;
  color: #ccc;
}

.status-badge.status-locked {
  background: #e7e337;
  color: #070707;
}

.status-badge.status-available {
  background: #66ff66;
  color: #000;
}

.cmd-window {
  background: #1a001a;
  color: #ff66cc;
  font-family: 'Courier New', monospace;
  padding: 12px;
  margin: 15px 0;
  border: 2px solid #ff99ff;
  box-shadow: inset 0 0 10px #9900cc;
}

.cmd-header {
  color: #ffccff;
  font-size: 10px;
  margin-bottom: 10px;
  border-bottom: 1px solid #ff66cc;
  padding-bottom: 5px;
}

.cmd-content {
  font-size: 10px;
  max-height: 300px;
  overflow-y: auto;
}

.loading {
  text-align: center;
  padding: 50px;
}

.loading-text {
  font-size: 12px;
  color: #ff66cc;
  animation: pulse 1.5s infinite;
}

.error {
  text-align: center;
  padding: 20px;
}

.error-text {
  color: #ff3366;
  font-size: 10px;
  margin-bottom: 10px;
}

.retry-btn {
  background: #ff66cc;
  border: 2px solid #ff99ff;
  padding: 5px 15px;
  cursor: pointer;
  font-family: 'Press Start 2P', cursive;
  font-size: 8px;
  color: #ffffff;
}

.retry-btn:hover {
  background: #cc00ff;
}

.footer {
  text-align: center;
  font-size: 8px;
  color: #ffccff;
  margin-top: 15px;
  text-shadow: 1px 1px #330033;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>