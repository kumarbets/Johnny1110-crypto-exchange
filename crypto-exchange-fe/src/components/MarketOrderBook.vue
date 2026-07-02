<template>
  <div>

    <div class="trading-section">

      <div class="chart-container">
        <p class="latest-price">Latest Price: {{ latestPrice }} {{ quoteAsset }}</p>
        <h3>K-Line Chart ({{ baseAsset }}/ {{ quoteAsset }})</h3>
        <KlineChart :market="market" :interval="chartInterval"/>
      </div>


      <div class="orderbook-container">
        <h3>Order Book</h3>

        <h4>Asks</h4>
        <table class="orderbook-table ask">
          <thead>
          <tr>
            <th>Price ({{ quoteAsset }})</th>
            <th>Amount ({{ baseAsset }})</th>
            <th>Volume Bar</th>
            <th>Total ({{ quoteAsset }})</th>
          </tr>
          </thead>
          <tbody>
          <!--          把最小的放在最前面用sort-->
          <tr v-for="(ask, index) in askSide" :key="'ask-' + index">
            <td>{{ ask.price.toFixed(3) }}</td>
            <td>{{ ask.volume.toFixed(3) }}</td>
            <td>
              <div
                  class="volume-bar"
                  :style="{ width: `${(ask.volume / maxAskVolume) * 100}px` }"
              ></div>
            </td>
            <td>{{ (ask.price * ask.volume).toFixed(2) }}</td>
          </tr>
          </tbody>
        </table>

        <h4>Bids</h4>
        <table class="orderbook-table bid">
          <thead>
          <tr>
            <th>Price ({{ quoteAsset }})</th>
            <th>Amount ({{ baseAsset }})</th>
            <th>Volume Bar</th>
            <th>Total ({{ quoteAsset }})</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="(bid, index) in bidSide" :key="'bid-' + index">
            <td>{{ bid.price.toFixed(3) }}</td>
            <td>{{ bid.volume.toFixed(3) }}</td>
            <td>
              <div
                  class="volume-bar"
                  :style="{ width: `${(bid.volume / maxBidVolume) * 100}px` }"
              ></div>
            </td>
            <td>{{ (bid.price * bid.volume).toFixed(2) }}</td>
          </tr>

          </tbody>
        </table>
      </div>

      <div class="order-form-container">
        <h3>Place Order <b>({{ orderType }})</b></h3>

        <div class="tab-bar button-group">
          <div
              class="tab"
              :class="{ active: orderType === 'limit' }"
              @click="orderType = 'limit'">Limit
          </div>
          <div
              class="tab"
              :class="{ active: orderType === 'market' }"
              @click="orderType = 'market'">Market
          </div>
        </div>

        <div class="button-group">
          <button
              :class="{ active: placeOrderBtn === 'Buy' }"
              class="buy"
              @click="changePlaceOrderBtn('Buy')">
            Buy
          </button>
          <button
              :class="{ active: placeOrderBtn === 'Sell' }"
              class="sell"
              @click="changePlaceOrderBtn('Sell')">
            Sell
          </button>
        </div>


        <!-- Limit Order -->
        <div v-show="orderType === 'limit'" class="tab-content">
          <label for="limit-price">Price ({{ quoteAsset }}):</label>
          <input
              id="limit-price"
              type="number"
              v-model="limitPrice"
              placeholder="Enter price"
              step="0.01"
          >

          <label for="limit-amount">Amount ({{ baseAsset }}):</label>
          <input
              id="limit-amount"
              type="number"
              v-model="limitAmount"
              placeholder="Enter amount"
              step="0.00001"
          >

        </div>

        <!-- Market Order -->
        <div v-show="orderType === 'market'" class="market-container">
          <div class="tab-content" v-show="placeOrderBtn === 'Sell'">
            <label for="sell-size">Sell Size ({{ baseAsset }}):</label>
            <input
                id="sell-size"
                type="number"
                v-model="marketSellSize"
                placeholder="Enter Sell Size"
                step="0.00001"
            >
            <!--            <div class="button-group">-->
            <!--              <button @click="placeMarketOrder('sell')" class="sell-btn">Sell</button>-->
            <!--            </div>-->
          </div>
          <div class="tab-content" v-show="placeOrderBtn === 'Buy'">
            <label for="buy-amount">Buy Amount ({{ quoteAsset }}):</label>
            <input
                id="buy-amount"
                type="number"
                v-model="marketBuyAmount"
                placeholder="Enter Buy amount"
                step="0.00001"
            >

          </div>
        </div>

        <div class="range">
          <input
              type="range"
              min="0"
              max="99.9"
              step="0.1"
              v-model="orderPercentage"
          >
          <div class="range-button">
            <button class="range-button" @click="placeOrder(placeOrderBtn)">Confrim</button>
          </div>
        </div>

        <div class="balance-section">
          <h3>Percentage:{{ orderPercentage}} % </h3>
          <h3>Base Balance ({{ baseAsset }}): {{ baseBalance }}</h3>
          <h3>Quote Balance ({{ quoteAsset }}): {{ quoteBalance }}</h3>
        </div>
      </div>


    </div>
    <div>
      <h3>Open Orders</h3>
      <table class="orders-table">
        <thead>
        <tr>
          <th>Order ID</th>
          <th>Type</th>
          <th>Side</th>
          <th>Price ({{ quoteAsset }})</th>
          <th>Original Size ({{ baseAsset }})</th>
          <th>Filled ({{ baseAsset }})</th>
          <th>Status</th>
          <th>Cancel</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="(order, index) in openOrders" :key="'open-' + index">
          <td>{{ order.id }}</td>
          <td>{{ order.type === 0 ? 'LIMIT' : 'MARKET' }}</td>
          <td>{{ order.side === 0 ? 'BUY' : 'SELL' }}</td>
          <td>{{ order.price == undefined ? '-' : order.price }}</td>
          <td>{{ order.original_size }}</td>
          <td>{{ order.original_size - order.remaining_size }}</td>
          <td>{{ order.status }}</td>
          <td>
            <button @click="cancelOrder(order.id)">X</button>
          </td>
        </tr>
        </tbody>
      </table>
    </div>

    <div>
      <h3>Order History</h3>
      <table class="orders-table">
        <thead>
        <tr>
          <th>Order ID</th>
          <th>Type</th>
          <th>Side</th>
          <th>Price ({{ quoteAsset }})</th>
          <th>Amount ({{ baseAsset }})</th>
          <th>Average Price ({{ quoteAsset }})</th>
          <th>Fees</th>
          <th>Fee Asset</th>
          <th>Status</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="(order, index) in orderHistory" :key="'history-' + index">
          <td>{{ order.id }}</td>
          <td>{{ order.type === 0 ? 'LIMIT' : 'MARKET' }}</td>
          <td>{{ order.side === 0 ? 'BUY' : 'SELL' }}</td>
          <td>{{ order.price == undefined ? order.avg_dealt_price : order.price }}</td>
          <td>{{ order.original_size }}</td>
          <td>{{ order.avg_dealt_price }}</td>
          <td>{{ order.fees }}</td>
          <td>{{ order.fee_asset }}</td>
          <td>{{ order.status }}</td>
        </tr>
        </tbody>
      </table>
    </div>

    <div class="cmd-window" id="cmdOutput">
      <div class="cmd-output">
        <span v-if="cmdOutputList.length === 0">No command output yet.</span>
        <ul>
          <div v-for="(output, index) in cmdOutputList" :key="'output-' + index">{{ output }} <span
              class="cursor">_</span></div>
        </ul>
      </div>
    </div>
    <CommonModal
        :visible="showModal"
        @close="showModal = false"
        :commonData="commonData.data"
    />
  </div>
</template>


<script>
import {authUtils} from '@/services/auth'
import {walletAPI, orderBooksAPI, ordersAPI} from '@/services/apiService'
import CommonModal from "@/components/common/CommonModal.vue";
import KlineChart from "@/components/KLineChart.vue";

export default {
  name: 'MarketOrderBook',
  emits: ['navigate', 'logout'],
  components: {CommonModal, KlineChart},
  watch: {
    placeOrderBtn() {
      this.orderPercentage = 0
    },
    orderType(newVal) {
      this.orderPercentage = 0
      if (newVal === 'limit') {
        this.marketBuyAmount = 0
      } else {
        this.limitAmount = 0
        this.limitPrice = this.latestPrice
      }
    },

    orderPercentage(newVal) {
      const percentage = newVal / 100;
      const isBuy = this.placeOrderBtn === "Buy";
      const isLimit = this.orderType === 'limit';

      if (isLimit) {
        if (isBuy) {
          this.limitAmount = (Number(this.quoteBalance) * percentage) / this.limitPrice;
        } else {
          this.limitAmount = Number(this.baseBalance) * percentage;
        }
      } else {
        // market order
        if (isBuy) {
          this.marketBuyAmount = Number(this.quoteBalance) * percentage;
        } else {
          this.marketSellSize = Number(this.baseBalance) * percentage;
        }
      }
    }
  },
  data() {
    return {
      chartInterval: "15m",
      latestPrice: 0.0,
      openOrders: [],
      orderHistory: [],
      orderType: 'limit',
      market: "",
      placeOrderBtn: "Buy",
      baseAsset: "",
      quoteAsset: "",
      baseBalance: "",
      quoteBalance: "",
      orderPercentage: 0,
      limitPrice: 0.0,
      limitAmount: 0.0,
      marketSellSize: 0.0,
      marketBuyAmount: 0.0,
      balances: [],
      bidSide: [],
      askSide: [],
      maxBidVolume: 1,
      maxAskVolume: 1,
      cmdOutputList: [],
      showModal: false,
      commonData: {
        data: {
          isLoggedIn: false,
          "context": "",
          "title": "",
        }
      },
    };
  },

  async mounted() {
    const marketName = this.$route.params.marketName // 從路由中取得參數
    if (!marketName) {
      console.error('No market name in route')
      return
    }

    this.market = marketName
    var assets = marketName.split('-')
    this.baseAsset = assets[0]
    this.quoteAsset = assets[1]

    await this.fetchOrderBook()

    this.limitPrice = this.latestPrice

    await this.fetchOpenOrders()
    await this.fetchClosedOrders()
    await this.fetchBalances()
    this.startAutoRefresh()
    const baseAsset = 'ETH'; // Example dynamic data
    const quoteAsset = 'USD'; // Example dynamic data
    this.cmdOutputList.push(`C:\\CryptoEx> trading ${baseAsset}/${quoteAsset}
    Loading market data...`);
    if (authUtils.isAuthenticated()) {
      //加入but最貴價格
      this.limitPrice = this.askSide[0] ? this.askSide[0].price : 0.0;
      //先拿到登入後價錢,如果沒登入不顯示
      this.orderPercentage = 0;
    }
  },


  computed: {
    totalAssets() {
      return this.balances.length
    },
    nonZeroBalances() {
      return this.balances.filter(balance => balance.total > 0).length
    },

  },

  beforeUnmount() {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval)
    }
  },
  methods: {

    async cancelOrder(orderId) {
      try {
        // 檢查是否已登入
        if (!authUtils.isAuthenticated()) {
          return
        }

        const response = await ordersAPI.cancelOrder(orderId)

        if (response.data.code === '0000000') {
          alert("order canceled")
          this.fetchOpenOrders()
          this.fetchClosedOrders()
          this.refreshBalances()
        } else {
          throw new Error(response.data.message || 'failed cancel')
        }
      } catch (error) {
        this.error = error.response?.data?.message || error.message || 'Network error occurred'
        console.error('failed cancel:', error)
      } finally {
        this.loading = false
      }
    },

    async fetchOpenOrders() {
      try {
        // 檢查是否已登入
        if (!authUtils.isAuthenticated()) {
          return
        }

        const response = await ordersAPI.getOpenOrders(this.market)

        if (response.data.code === '0000000') {
          this.openOrders = response.data.data.result
          console.log(this.openOrders)
        } else {
          throw new Error(response.data.message || 'Failed to fetch open orders')
        }
      } catch (error) {
        this.error = error.response?.data?.message || error.message || 'Network error occurred'
        console.error('Error fetching orders:', error)
      } finally {
        this.loading = false
      }
    },

    async fetchClosedOrders() {
      try {
        // 檢查是否已登入
        if (!authUtils.isAuthenticated()) {
          return
        }

        const response = await ordersAPI.getClosedOrders(this.market)

        if (response.data.code === '0000000') {
          this.orderHistory = response.data.data.result
          console.log(this.orderHistory)
        } else {
          throw new Error(response.data.message || 'Failed to fetch open orders')
        }
      } catch (error) {
        this.error = error.response?.data?.message || error.message || 'Network error occurred'
        console.error('Error fetching orders:', error)
      } finally {
        this.loading = false
      }
    },


    async placeOrder(side) {
      if (this.orderType === 'limit') {
        await this.placeLimitOrder(side);
      } else {
        await this.placeMarketOrder(side);
      }

    },
    async placeLimitOrder(side) {

      var response

      if (side === 'Buy') {
        if (this.limitAmount <= 0) {
          alert("[WARNING]: size must greater than 0")
          return
        }
        if (this.limitPrice <= 0) {
          alert("[WARNING]: limit price must greater than 0")
          return
        }
        response = await ordersAPI.placeLimitBuyOrder(this.market, this.limitPrice, this.limitAmount)

      } else {
        if (this.limitAmount > this.baseBalance) {
          alert("insufficient " + this.baseAsset + " balance")
          return
        }
        response = await ordersAPI.placeLimitSellOrder(this.market, this.limitPrice, this.limitAmount)
      }

      if (response.data.code === '0000000') {
        this.showModal = true
        this.commonData.data.context = "Successfully placed " + side + " order for " + this.market + ".\n" +
            "Price: " + this.limitPrice + ", Amount: " + this.limitAmount + ".\n" +
            "Please check your orders.";
        this.commonData.data.title = "Limit Order Confirmation"
        this.fetchOpenOrders()
        this.fetchClosedOrders()
        this.refreshBalances()
        this.cmdOutputList.push(`C:\\CryptoEx> trading ${this.market} - ${side} order placed.\n` +
            `Price: ${this.limitPrice}, Amount: ${this.limitAmount}.\n`);
      } else {
        this.showModal = true
        this.commonData.data.context = "!Failed placed " + side + " order for " + this.market + ".\n" +
            "Reason: " + response.data.message;
        this.commonData.data.title = "Limit Order Failed"
      }



    },

    async placeMarketOrder(side) {
      console.log('Placing market order:', side, this.market);
      var response
      if (side === 'Buy') {
        if (this.marketBuyAmount <= 0) {
            alert("[WARNING]: quote amount must greater than 0")
            return
          }
          if (this.marketBuyAmount > this.quoteBalance) {
            alert("insufficient " + this.quoteAsset + " balance")
            return
          }
        response = await ordersAPI.placeMarketBuyOrder(this.market, this.marketBuyAmount)
      } else {
        if (this.marketSellSize <= 0) {
          alert("[WARNING]: quote amount must greater than 0")
          return
        }

        if (this.marketSellSize > this.baseBalance) {
          alert("insufficient " + this.baseAsset + " balance")
          return
        }

        response = await ordersAPI.placeMarketSellOrder(this.market, this.marketSellSize)
        if (response.data.code === '0000000') {
          this.commonData.data.title = "Market Order Confirmation"
          this.commonData.data.context = "Successfully placed " + side + " order for " + this.market + ".\n" +
              " Sell Size: " + this.marketSellSize + ".\n" +
              "Please check your orders.";
        } else {
          this.commonData.data.title = "Market Order Failed"
          this.commonData.data.context = "!Failed placed " + side + " order for " + this.market + ".\n" +
              "Reason:" + response.data.message;
        }
      }

      this.showModal = true

      this.fetchOpenOrders()
      this.fetchClosedOrders()
      this.refreshBalances()
      this.cmdOutputList.push(`C:\\CryptoEx> trading ${this.market} - ${side} order placed.\n`);
    },

    async fetchOrderBook() {
      try {
        const res = await orderBooksAPI.getOrderBook(this.market);
        const data = res.data.data;

        this.latestPrice = data.latest_price;


// 假設 data.bid_side 和 data.ask_side 的結構如下：
// [{ price: 100 }, { price: 200 }, ...]

        let bidSide = data.bid_side;
        let askSide = data.ask_side;

// 取得最大前五個的 bid_side
        let topFiveBid = [...bidSide]
            .sort((a, b) => b.price - a.price) // 由大到小排序
            .slice(0, 5); // 取前五個

// 取得最小前五個的 ask_side
        let bottomFiveAsk = [...askSide]
            .sort((a, b) => a.price - b.price) // 由小到大排序
            .slice(0, 5) // 取前五個
            .reverse();

// 返回結果到 this
        this.bidSide = topFiveBid;
        this.askSide = bottomFiveAsk;
        // 計算最大量，用於 bar 寬度百分比
        this.maxBidVolume = Math.max(...this.bidSide.map((b) => b.volume));
        this.maxAskVolume = Math.max(...this.askSide.map((a) => a.volume));
      } catch (error) {
        console.error("Failed to fetch order book:", error);
      }
    },

    async handlePriceUpdate(data) {
      console.log("handlePriceUpdate:", data)
    },

    async fetchBalances() {
      this.loading = true
      this.error = null

      try {
        // 檢查是否已登入
        if (!authUtils.isAuthenticated()) {
          return
        }

        const response = await walletAPI.getBalances()

        if (response.data.code === '0000000') {
          this.balances = response.data.data

          this.balances.forEach(b => {
            console.log(b)
            if (b.asset === this.baseAsset) {
              this.baseBalance = parseFloat(b.total).toFixed(4); // 顯示小數點後 4 位
            }
            if (b.asset === this.quoteAsset) {
              this.quoteBalance = parseFloat(b.total).toFixed(2); // 顯示小數點後 2 位
            }
          })

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
      if (balance.locked > 0) return 'Locked'
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
        //this.fetchBalances()
        this.fetchOrderBook()
      }, 1000) // 每5秒更新一次
    },
    changePlaceOrderBtn(btnName) {
      this.placeOrderBtn = btnName;
    }
  }
}
</script>

<style>
body {
  background: linear-gradient(180deg, #ff66cc, #9900cc);
  font-family: 'Press Start 2P', cursive;
  color: #ffffff;
  margin: 0;
  padding: 20px;
  background-image: url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAYAAACp8Z5+AAAAG0lEQVR4AWMAAv+BAgICAgICiAECAwICAgICAgICBtgBc4AAAAASUVORK5CYII=');
  background-repeat: repeat;
}

/*.container {*/
/*  background: rgba(51, 0, 51, 0.8);*/
/*  border: 3px solid #ff99ff;*/
/*  width: 900px;*/
/*  margin: 50px auto;*/
/*  padding: 15px;*/
/*  box-shadow: 0 0 10px #ff66cc, 0 0 20px #9900cc;*/
/*  border-radius: 5px;*/
/*}*/
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
  margin: 15px 0;
}

.nav-bar a {
  margin-right: 15px;
  color: #ffccff;
  text-decoration: none;
  font-size: 10px;
  text-shadow: 1px 1px #330033;
}

.nav-bar a:hover {
  color: #ff66cc;
  text-shadow: 0 0 5px #ff66cc;
}

.balance-section {
  margin: 15px 0;
  font-size: 10px;
  color: #ffccff;
  text-shadow: 1px 1px #330033;
}

.trading-section {

  display: flex;
  flex-wrap: wrap;
  gap: 15px;
}

.chart-container, .orderbook-container, .order-form-container {
  flex: 1;
  min-width: 280px;
  background: rgba(255, 255, 255, 0.1);
  border: 2px solid #ff99ff;
  padding: 15px;
  box-shadow: inset 0 0 5px #9900cc;
}

.chart-container canvas {
  max-height: 300px;
}

.latest-price {
  font-size: 12px;
  font-weight: bold;
  margin: 5px 0;
  color: #ffccff;
  text-shadow: 1px 1px #330033;
}

.orderbook-container h4 {
  font-size: 11px;
  margin: 10px 0;
  border-bottom: 1px solid #ff99ff;
  color: #ffcccc;
  text-shadow: 3px 3px #330033;
}

.orderbook-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 10px;
  margin-bottom: 10px;
  color: #ffccff;
}

.orderbook-table th, .orderbook-table td {
  border: 1px solid #ff99ff;
  padding: 3px;
  text-align: right;
  text-shadow: 1px 1px #330033;
}

.orderbook-table th {
  background: linear-gradient(90deg, #ff33cc, #cc00ff);
  color: #ffffff;
}

.orderbook-table.bid tr {
  background: rgba(0, 255, 0, 0.2);
}

.orderbook-table.ask tr {
  background: rgba(255, 0, 0, 0.2);
}

.volume-bar {
  display: inline-block;
  height: 8px;
  background: #ff66cc;
  border: 1px solid #ff99ff;
}

.order-form-container h3 {
  font-size: 13px;
  margin: 5px 0;
  color: #ffccff;
  text-shadow: 1px 1px #330033;
}

.tab-bar {
  display: flex;
  margin-bottom: 10px;
  justify-content: center;
}

.tab {
  background: #ff66cc;
  border: 2px solid #ff99ff;
  padding: 10px 10px;
  cursor: pointer;
  font-size: 12px;
  color: #ffffff;
  margin-top: 10px;
  margin-right: 10px;
  text-shadow: 1px 1px #330033;
  transition: all 1s;
}

.tab.active {
  background: #cc00ff;
  border: 2px solid #ff99ff;
  box-shadow: 0 0 5px #ff66cc;
  font-weight: bold;
}

.tab-content input[type="number"] {
  background-color: #fcd1ff; /* 淺粉紫色背景 */
  color: #5a0080; /* 深紫文字 */
  border: 4px solid #ff9eff; /* 粉紫色邊框 */
  font-family: 'Press Start 2P', cursive; /* 像素風字體 (需載入) */
  font-size: 14px;
  padding: 10px;
  outline: none;

  box-shadow: 4px 4px 0 #c93fff; /* 像素風陰影 */
  border-radius: 0; /* 保持硬邊像素風 */
  width: 90%;
  text-align: center;
}

/* 可選：hover 和 focus 效果強化遊戲感 */
.tab-content input[type="number"]:hover,
.tab-content input[type="number"]:focus {
  background-color: #ffe6ff;
  border-color: #ff00ff;
  box-shadow: 4px 4px 0 #ff00ff;
  color: #8000ff;
}

.tab-content.active {
  display: block;
}

.order-form-container label {
  display: block;
  font-size: 10px;
  margin: 10px 0 5px;
  color: #ffccff;
  text-shadow: 1px 1px #330033;
}

.order-form-container input {
  width: calc(100% - 14px);
  padding: 6px;
  border: 2px solid #ff99ff;
  background: rgba(255, 255, 255, 0.1);
  font-family: 'Press Start 2P', cursive;
  font-size: 10px;
  color: #ffffff;
  box-shadow: inset 0 0 5px #9900cc;
}

.order-form-container button {
  background: #ff66cc;
  border: 2px solid #ff99ff;
  padding: 6px 12px;
  cursor: pointer;
  font-family: 'Press Start 2P', cursive;
  font-size: 10px;
  color: #ffffff;
  width: 48%;
  margin: 10px 5px 0 0;
  box-shadow: 0 0 5px #ff66cc;
  text-shadow: 1px 1px #330033;
  transition: all 0.2s;
}

.order-form-container button:hover {
  background: #cc00ff;
  box-shadow: 0 0 8px #ff66cc;
}

.buy-btn {
  background: #00cc00;
}

.sell-btn {
  background: #cc0000;
}

.orders-table {
  width: 100%;
  border-collapse: collapse;
  background: rgba(255, 255, 255, 0.1);
  border: 2px solid #ff99ff;
  margin: 15px 0;
  font-size: 7px;
  color: #ffccff;
}

.orders-table th, .orders-table td {
  border: 1px solid #ff99ff;
  padding: 5px;
  text-align: left;
  text-shadow: 1px 1px #330033;
}

.orders-table th {
  background: linear-gradient(90deg, #ff33cc, #cc00ff);
  color: #ffffff;
}

.cmd-window {
  background: #1a001a;
  color: #ff66cc;
  font-family: 'Courier New', monospace;
  padding: 12px;
  margin: 15px 0;
  border: 2px solid #ff99ff;
  height: 150px;
  overflow-y: auto;
  box-shadow: inset 0 0 10px #9900cc;
  font-size: 12px;
}

.button-group {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}

.button-group button {
  background-color: #e2ec8d; /* 淺粉紫 */
  color: #5a0080;            /* 深紫文字 */
  border: 4px solid #ff9eff; /* 粉紫邊框 */
  font-family: 'Press Start 2P', cursive;
  font-size: 12px;
  padding: 10px 16px;
  cursor: pointer;

  box-shadow: 4px 4px 0 #c93fff; /* 像素陰影 */
  border-radius: 0;
  transition: all 0.3s ease-in-out;
}

/* 選中狀態 - 新增 */
.button-group button.active {
  background-color: #ff00ff; /* 亮粉紫背景 */
  color: #ffffff;            /* 白色文字 */
  border-color: #8000ff;     /* 深紫邊框 */
  box-shadow: 4px 4px 0 #4a0080; /* 更深的陰影 */
}

/* hover 效果 */
.button-group button:hover {
  background-color: #e468ef;
  border-color: #ff00ff;
  box-shadow: 4px 4px 0 #ff00ff;
  color: #8000ff;
}

/* 選中狀態的 hover 效果 - 新增 */
.button-group button.active:hover {
  background-color: #cc00cc; /* 稍微暗一點的粉紫 */
  border-color: #6600cc;
  box-shadow: 4px 4px 0 #330066;
}

/* 遊戲感：按下時稍微縮進 */
.button-group button:active {
  box-shadow: 2px 2px 0 #c93fff;
  transform: translate(2px, 2px);
}

/* 選中狀態按下效果 - 新增 */
.button-group button.active:active {
  box-shadow: 2px 2px 0 #4a0080;
  transform: translate(2px, 2px);
}

/* Buy 按鈕選中狀態 */
.button-group button.buy.active {
  background-color: #00ff88; /* 綠色系 */
  color: #004400;
  border-color: #00cc66;
  box-shadow: 4px 4px 0 #006633;
}

/* Sell 按鈕選中狀態 */
.button-group button.sell.active {
  background-color: #ff4488; /* 紅色系 */
  color: #ffffff;
  border-color: #cc0044;
  box-shadow: 4px 4px 0 #880033;
}

.market-container {
  display: flex;
  height: 27%;
}

.left-pane {
  flex: 1.5;
  padding: 1rem;
  border-right: 1px solid #ccc; /* 分隔左右的線條 */
}

.right-pane {
  flex: 1.5;
  padding: 1rem;
}

.range {
  width: 100%;
  max-width: 300px;
  margin: 20px auto;
}

.range input[type="range"] {
  -webkit-appearance: none;
  width: 100%;
  height: 8px;
  background: #ddd;
  border-radius: 4px;
  outline: none;
  transition: background 0.3s;
}

.range input[type="range"]::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #007bff;
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: background 0.3s, transform 0.2s;
}

.range input[type="range"]::-moz-range-thumb {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #007bff;
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: background 0.3s, transform 0.2s;
}

.range input[type="range"]:hover::-webkit-slider-thumb {
  background: #0056b3;
  transform: scale(1.1);
}

.range input[type="range"]:hover::-moz-range-thumb {
  background: #0056b3;
  transform: scale(1.1);
}

.range input[type="range"]:focus {
  background: #bbb;
}

.button-group {
  display: flex;
  margin-top: 10px;
  margin-bottom: 10px;
}

.range-button {
  font-size: 24px;
  height: 50px;
  width: 100%;
  margin: fill;
}
</style>