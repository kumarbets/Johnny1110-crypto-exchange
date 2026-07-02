<template>
  <table class="market-table">
    <thead>
    <tr>
      <th>Trading Pair</th>
      <th>Price (USDT)</th>
      <th>24h Change</th>
      <th>Volume</th>
    </tr>
    </thead>
    <tbody>
    <tr v-for="market in markets" :key="market.market_name" @click="selectMarket(market)">
      <td>{{ market.market_name }}</td>
      <td>{{ formatPrice(market.latest_price) }}</td>
      <td :style="{ color: market.price_change_24h >= 0 ? '#00ff99' : '#ff3366' }">
        {{ formatChange(market.price_change_24h) }}
      </td>
      <td>{{ formatVolume(market.total_volume_24h) }}</td>
    </tr>
    </tbody>
  </table>


  <CommandWindow :push-data="apiResponse" />


</template>

<script>
import router from "@/router";
import {marketAPI} from "@/services/apiService";
import CommandWindow from "@/components/CommandWindow.vue";

export default {
  name: 'MarketTable',
  components: {CommandWindow},
  data() {
    return {
      markets: [],
      apiResponse: {}
    }
  },

  mounted() {
    this.fetchMarketData()
    this.startAutoRefresh()
  },

  methods: {
    async fetchMarketData() {
      this.loading = true
      this.error = null

      try {
        const response = await marketAPI.getAllMarkets()

        if (response.data.code === '0000000') {
          // eslint-disable-next-line vue/no-mutating-props
          this.markets = response.data.data
          this.apiResponse = response.data
          this.lastUpdated = new Date().toLocaleTimeString()
        } else {
          throw new Error(response.data.message || 'Failed to fetch market data')
        }
      } catch (error) {
        this.error = error.response?.data?.message || error.message || 'Network error occurred'
        console.error('Error fetching market data:', error)
      } finally {
        this.loading = false
      }
    },

    async onMarketSelected(market) {
      this.selectedMarket = market

      try {
        const response = await marketAPI.getMarketData(market.market_name)
        if (response.data.code === '0000000') {
          // 更新選中市場的詳細數據
          this.selectedMarket = response.data.data
        }
      } catch (error) {
        console.error('Error fetching market details:', error)
      }
    },

    startAutoRefresh() {
      this.refreshInterval = setInterval(() => {
        this.fetchMarketData()
      }, 30000) // 每30秒更新一次
    },

    formatPrice(price) {
      return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 8
      }).format(price)
    },

    formatChange(change) {
      const sign = change >= 0 ? '+' : ''
      return `${sign}${(change * 100).toFixed(2)}%`
    },

    formatVolume(volume) {
      if (volume >= 1000000) {
        return `${(volume / 1000000).toFixed(2)}M`
      } else if (volume >= 1000) {
        return `${(volume / 1000).toFixed(2)}K`
      }
      return volume.toFixed(2)
    },

    selectMarket(market) {
      console.log(market)
      router.push('/markets/' + market.market_name)
    }
  }
}
</script>

<style scoped>
.market-table {
  width: 100%;
  border-collapse: collapse;
  background: rgba(255, 255, 255, 0.1);
  border: 2px solid #ff99ff;
  margin: 15px 0;
}

.market-table th, .market-table td {
  border: 1px solid #ff66cc;
  padding: 8px;
  text-align: left;
  font-size: 10px;
  color: #ffffff;
}

.market-table th {
  background: linear-gradient(90deg, #cc00ff, #ff33cc);
  text-shadow: 1px 1px #330033;
}

.market-table tbody tr {
  cursor: pointer;
  transition: background-color 0.2s;
}

.market-table tbody tr:hover {
  background: rgba(255, 102, 204, 0.2);
}
</style>