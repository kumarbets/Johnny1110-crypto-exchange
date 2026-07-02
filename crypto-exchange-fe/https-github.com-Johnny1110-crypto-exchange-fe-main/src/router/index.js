import { createRouter, createWebHistory } from 'vue-router'

// 匯入頁面元件
import Home from '@/views/Home.vue'
import WalletBalance from '@/components/WalletBalance.vue'
import MarketOrderBook from '@/components/MarketOrderBook.vue'
import MarketTable from "@/components/MarketTable.vue";
import KLineChart from "@/components/KLineChart.vue";

const routes = [
    { path: '/', name: 'Home', component: Home },
    { path: '/balances', name: 'Balances', component: WalletBalance },
    { path: '/markets/:marketName', name: 'Markets', component: MarketOrderBook },
    { path: '/markets/list', name: 'MarketList', component: MarketTable },
    { path: '/test/kline', name: 'testKline', component: KLineChart },
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

export default router
