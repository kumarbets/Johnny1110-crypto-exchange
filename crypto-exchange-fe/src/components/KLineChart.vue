<template>
  <div class="ohlcv-chart-container">
    <div class="connection-status" :class="connectionStatus.class">
      {{ connectionStatus.text }}
    </div>

    <!-- Interval 切換按鈕區域 -->
    <div class="interval-controls">
      <button
          v-for="option in intervalOptions"
          :key="option.value"
          :class="['interval-btn', { active: currentInterval === option.value }]"
          @click="changeInterval(option.value)"
          :disabled="isLoading"
      >
        {{ option.label }}
      </button>
    </div>

    <div ref="chartContainer" class="chart"></div>
  </div>
</template>

<script setup>
import {ref, onMounted, onUnmounted, watch, computed} from 'vue'
import {createChart} from 'lightweight-charts'
import {ohlcvAPI} from '@/services/apiService'
import websocketService from '@/services/websocketService'

// eslint-disable-next-line no-undef
const props = defineProps({
  market: {type: String, default: 'ETH-USDT'},
  interval: {type: String, default: '15m'},
})

const chartContainer = ref(null)
const chart = ref(null)
const candleSeries = ref(null)
const volumeSeries = ref(null)
const isLoading = ref(false)
const wsConnected = ref(false)
const unsubscribeWs = ref(null)

// 內部狀態管理當前選中的 interval
const currentInterval = ref(props.interval)

// Interval 選項配置
const intervalOptions = [
  { value: '15m', label: '15m' },
  { value: '1h', label: '1h' },
  { value: '1d', label: '1d' },
  { value: '1w', label: '1w' }
]

// 計算連線狀態顯示
const connectionStatus = computed(() => {
  if (wsConnected.value) {
    return {
      text: '● Live',
      class: 'connected'
    }
  } else {
    return {
      text: '○ connecting...',
      class: 'connecting'
    }
  }
})

function transformData(raw) {
  const candleData = []
  const volumeData = []

  for (let i = 0; i < raw.t.length; i++) {
    const time = Math.floor(raw.t[i]) // 秒為單位
    candleData.push({
      time,
      open: raw.o[i],
      high: raw.h[i],
      low: raw.l[i],
      close: raw.c[i],
    })
    volumeData.push({
      time,
      value: Math.abs(raw.v[i]), // 確保 volume 值為正數
      color: raw.c[i] >= raw.o[i] ? '#26a69a' : '#ef5350',
    })
  }

  return {candleData, volumeData}
}

async function fetchData(market, interval) {
  try {
    const res = await ohlcvAPI.getOhlcvHistory(market, interval)
    if (res?.data?.code === '0000000' && res.data?.data?.t) {
      return transformData(res.data.data)
    } else {
      console.warn('API 回傳格式不正確', res)
      return {candleData: [], volumeData: []}
    }
  } catch (err) {
    console.error('API 請求失敗:', err)
    return {candleData: [], volumeData: []}
  }
}

async function initChart() {
  if (!chartContainer.value) return

  // 如果圖表已存在，先清理
  if (chart.value) {
    chart.value.remove()
  }

  chart.value = createChart(chartContainer.value, {
    width: chartContainer.value.clientWidth,
    height: 500,
    layout: {
      background: {color: '#1e1e2f'},
      textColor: '#ffffff',
    },
    grid: {
      vertLines: {color: '#2b2b43'},
      horzLines: {color: '#363c4e'},
    },
    timeScale: {
      timeVisible: true,
    },
    // 只啟用右側價格軸，隱藏左側
    leftPriceScale: {
      visible: false,
    },
    rightPriceScale: {
      visible: true,
      borderColor: '#485c7b',
    }
  })

  // K 線圖使用右側價格軸
  candleSeries.value = chart.value.addCandlestickSeries({
    upColor: '#26a69a',
    downColor: '#ef5350',
    wickUpColor: '#26a69a',
    wickDownColor: '#ef5350',
    borderVisible: false,
    priceScaleId: 'right',
  })

  // Volume 圖使用獨立的價格軸 ID
  volumeSeries.value = chart.value.addHistogramSeries({
    color: '#26a69a',
    priceFormat: {type: 'volume'},
    priceScaleId: 'volume',
  })

  // 配置右側價格軸（K線）- 只顯示正數，占據上半部
  chart.value.priceScale('right').applyOptions({
    scaleMargins: {
      top: 0.1,
      bottom: 0.5, // 為 volume 留出下半部空間
    },
    autoScale: true,
    mode: 1, // 1 = logarithmic mode，有助於只顯示正數
    invertScale: false,
  })

  // 配置 Volume 價格軸 - 顯示在右側下半部，只保留正數
  chart.value.priceScale('volume').applyOptions({
    scaleMargins: {
      top: 0.6, // volume 圖在下半部
      bottom: 0.05,
    },
    autoScale: true,
    position: 'right', // 明確指定顯示在右側
    visible: true,
    mode: 1, // 1 = logarithmic mode，有助於只顯示正數
    invertScale: false,
  })

  // 載入初始數據
  await loadData(props.market, currentInterval.value)
}

async function loadData(market, interval) {
  if (!candleSeries.value || !volumeSeries.value) {
    console.warn('圖表尚未初始化')
    return
  }

  isLoading.value = true
  try {
    const {candleData, volumeData} = await fetchData(market, interval)
    candleSeries.value.setData(candleData)
    volumeSeries.value.setData(volumeData)
  } catch (error) {
    console.error('載入數據失敗:', error)
  } finally {
    isLoading.value = false
  }
}

/**
 * 初始化 WebSocket 連線
 */
async function initWebSocket() {
  try {
    // 建立 WebSocket 連線
    await websocketService.connect('remote')
    wsConnected.value = true
    console.log('WebSocket 連線成功')

    // 訂閱 OHLCV 數據
    subscribeToOhlcv(props.market, currentInterval.value)

  } catch (error) {
    console.error('WebSocket 連線失敗:', error)
    wsConnected.value = false
  }
}

/**
 * 訂閱 OHLCV 數據
 */
function subscribeToOhlcv(market, interval) {
  // 如果已有訂閱，先取消
  if (unsubscribeWs.value) {
    unsubscribeWs.value()
  }

  // 訂閱新的 OHLCV 數據
  unsubscribeWs.value = websocketService.subscribeOhlcv(
      market,
      interval,
      handleRealtimeData
  )

  console.log(`訂閱 WebSocket OHLCV: ${market} ${interval}`)
}

/**
 * 處理即時數據更新
 */
function handleRealtimeData(data, timestamp) {
  if (!candleSeries.value || !volumeSeries.value) {
    console.warn('圖表尚未初始化，無法更新即時數據')
    return
  }

  const { candleData, volumeData } = data

  if (candleData.length > 0 && volumeData.length > 0) {
    // 獲取最新的數據點
    const latestCandle = candleData[candleData.length - 1]
    const latestVolume = volumeData[volumeData.length - 1]

    console.log('更新即時數據:', { latestCandle, latestVolume, timestamp })

    // 更新圖表數據
    candleSeries.value.update(latestCandle)
    volumeSeries.value.update(latestVolume)
  }
}

/**
 * 切換 interval
 */
function changeInterval(newInterval) {
  if (newInterval === currentInterval.value || isLoading.value) {
    return
  }

  currentInterval.value = newInterval
  console.log(`切換到 ${newInterval}`)

  // 重新載入數據和訂閱
  handleMarketChange(props.market, newInterval)
}

/**
 * 清理 WebSocket 訂閱
 */
function cleanupWebSocket() {
  if (unsubscribeWs.value) {
    unsubscribeWs.value()
    unsubscribeWs.value = null
  }
}

/**
 * 監聽 market 或 interval 變化，重新訂閱
 */
async function handleMarketChange(market, interval) {
  // 先載入歷史數據
  await loadData(market, interval)

  // 如果 WebSocket 已連線，重新訂閱
  if (wsConnected.value) {
    subscribeToOhlcv(market, interval)
  }
}

// 組件掛載時初始化
onMounted(async () => {
  await initChart()
  await initWebSocket()
})

// 組件卸載時清理資源
onUnmounted(() => {
  // 清理圖表
  if (chart.value) {
    chart.value.remove()
  }

  // 清理 WebSocket 訂閱
  cleanupWebSocket()

  // 斷開 WebSocket 連線（如果需要的話）
  // websocketService.disconnect()
})

// 監聽 props 變化
watch(
    () => [props.market, props.interval],
    async ([market, interval]) => {
      currentInterval.value = interval
      await handleMarketChange(market, interval)
    }
)

// 監聽內部 interval 變化
watch(currentInterval, async (newInterval) => {
  await handleMarketChange(props.market, newInterval)
})

// 監聽 WebSocket 連線狀態變化
watch(wsConnected, (connected) => {
  if (connected) {
    // 連線成功後訂閱當前市場數據
    subscribeToOhlcv(props.market, currentInterval.value)
  }
})
</script>

<style scoped>
.ohlcv-chart-container {
  padding: 10px;
  position: relative;
}

.connection-status {
  position: absolute;
  top: 15px;
  right: 15px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  z-index: 10;
  backdrop-filter: blur(4px);
}

.connection-status.connected {
  background: rgba(38, 166, 154, 0.2);
  color: #26a69a;
  border: 1px solid rgba(38, 166, 154, 0.3);
}

.connection-status.connecting {
  background: rgba(255, 193, 7, 0.2);
  color: #ffc107;
  border: 1px solid rgba(255, 193, 7, 0.3);
  animation: pulse 2s infinite;
}

.interval-controls {
  position: absolute;
  top: 15px;
  left: 15px;
  display: flex;
  gap: 8px;
  z-index: 10;
}

.interval-btn {
  padding: 6px 12px;
  background: rgba(30, 30, 47, 0.8);
  color: #ffffff;
  border: 1px solid rgba(72, 92, 123, 0.5);
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  backdrop-filter: blur(4px);
}

.interval-btn:hover:not(:disabled) {
  background: rgba(72, 92, 123, 0.3);
  border-color: rgba(38, 166, 154, 0.5);
  transform: translateY(-1px);
}

.interval-btn.active {
  background: rgba(38, 166, 154, 0.2);
  color: #26a69a;
  border-color: rgba(38, 166, 154, 0.6);
  box-shadow: 0 0 8px rgba(38, 166, 154, 0.3);
}

.interval-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@keyframes pulse {
  0% { opacity: 1; }
  50% { opacity: 0.6; }
  100% { opacity: 1; }
}

.chart {
  width: 100%;
  height: 500px;
}
</style>