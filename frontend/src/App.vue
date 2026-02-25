<template>
  <div class="app">
    <header class="header">
      <div>
        <h1>股票策略选股系统</h1>
        <p>日线 + 30分钟级别数据、本地策略管理、选股与回测一体化。</p>
      </div>
      <div class="header-actions">
        <button class="secondary" @click="seedDemo">导入示例数据</button>
        <button class="primary" @click="refreshAll">刷新数据</button>
      </div>
    </header>

    <section class="panel-grid">
      <div class="card">
        <h2>策略管理</h2>
        <div class="form-row">
          <label>策略名称</label>
          <input v-model="form.name" placeholder="例如：MA短期上穿" />
        </div>
        <div class="form-row">
          <label>策略描述</label>
          <textarea v-model="form.description" placeholder="说明策略逻辑与适用范围"></textarea>
        </div>
        <div class="form-row">
          <label>参数(JSON)</label>
          <input v-model="form.params" placeholder='{"short_window":5,"long_window":20}' />
        </div>
        <button class="primary" @click="createStrategy">新增策略</button>
      </div>

      <div class="card">
        <h2>策略执行</h2>
        <div class="form-row">
          <label>选择策略</label>
          <select v-model.number="selectedStrategyId">
            <option disabled value="">请选择策略</option>
            <option v-for="strategy in strategyStore.strategies" :key="strategy.id" :value="strategy.id">
              {{ strategy.name }} ({{ strategy.type }})
            </option>
          </select>
        </div>
        <button class="primary" @click="runScreening" :disabled="!selectedStrategyId">
          运行选股
        </button>
        <div class="result-list">
          <div v-for="item in screeningResults" :key="item.stock.code" class="result-item">
            <div>
              <strong>{{ item.stock.name }}</strong>
              <span>{{ item.stock.code }} · {{ item.stock.exchange }}</span>
            </div>
            <div class="metrics">
              <div>
                <div>理由</div>
                <strong>{{ item.reason }}</strong>
              </div>
            </div>
          </div>
          <div v-if="!screeningResults.length" class="footer-note">暂无命中结果。</div>
        </div>
      </div>

      <div class="card">
        <h2>回测配置</h2>
        <div class="form-row">
          <label>股票代码</label>
          <select v-model="selectedStockCode">
            <option disabled value="">请选择股票</option>
            <option v-for="stock in stockStore.stocks" :key="stock.code" :value="stock.code">
              {{ stock.name }} ({{ stock.code }})
            </option>
          </select>
        </div>
        <div class="form-row">
          <label>初始资金</label>
          <input v-model.number="initialCapital" type="number" min="10000" step="1000" />
        </div>
        <button class="primary" @click="runBacktest" :disabled="!selectedStrategyId || !selectedStockCode">
          启动回测
        </button>
        <div v-if="backtestStore.summary" class="metrics">
          <div>
            <div>收益率</div>
            <strong>{{ backtestStore.summary.return_pct.toFixed(2) }}%</strong>
          </div>
          <div>
            <div>终值</div>
            <strong>{{ backtestStore.summary.final_capital.toFixed(0) }}</strong>
          </div>
        </div>
      </div>

      <div class="card">
        <h2>AkShare 同步</h2>
        <div class="form-row">
          <label>股票代码（逗号分隔，留空则按数量取样）</label>
          <input v-model="syncForm.symbols" placeholder="000001,600519" />
        </div>
        <div class="form-row">
          <label>同步模式</label>
          <select v-model="syncForm.mode">
            <option value="all">日线 + 分钟线</option>
            <option value="daily">仅日线</option>
            <option value="minute">仅分钟线</option>
          </select>
        </div>
        <div class="form-row">
          <label>日线起始 (YYYYMMDD)</label>
          <input v-model="syncForm.startDate" placeholder="20240101" />
        </div>
        <div class="form-row">
          <label>日线结束 (YYYYMMDD)</label>
          <input v-model="syncForm.endDate" placeholder="20241231" />
        </div>
        <div class="form-row">
          <label>分钟起始 (YYYY-MM-DD HH:MM:SS)</label>
          <input v-model="syncForm.minStart" placeholder="2024-12-01 09:30:00" />
        </div>
        <div class="form-row">
          <label>分钟结束 (YYYY-MM-DD HH:MM:SS)</label>
          <input v-model="syncForm.minEnd" placeholder="2025-02-01 15:00:00" />
        </div>
        <div class="form-row">
          <label>分钟周期</label>
          <select v-model="syncForm.period">
            <option value="1">1 分钟</option>
            <option value="5">5 分钟</option>
            <option value="15">15 分钟</option>
            <option value="30">30 分钟</option>
            <option value="60">60 分钟</option>
          </select>
        </div>
        <div class="form-row">
          <label>默认数量（未指定代码时）</label>
          <input v-model.number="syncForm.limit" type="number" min="1" step="1" />
        </div>
        <button class="primary" @click="runAkshareSync" :disabled="syncLoading">
          {{ syncLoading ? '同步中...' : '启动同步' }}
        </button>
        <p class="footer-note">留空将使用脚本默认范围，首次同步会下载依赖。</p>
        <div v-if="syncError" class="footer-note" style="color: #b42318;">{{ syncError }}</div>
        <div v-if="syncResult" class="result-list">
          <div class="result-item">
            <div>
              <strong>同步结果</strong>
              <span>股票 {{ syncResult.stocks }} · 日线 {{ syncResult.daily_rows }} · 分钟线 {{ syncResult.minute_rows }}</span>
            </div>
            <div class="metrics">
              <div>
                <div>错误数</div>
                <strong>{{ syncErrors.length }}</strong>
              </div>
            </div>
          </div>
          <div v-for="(err, idx) in syncErrors" :key="idx" class="footer-note">
            {{ err.symbol }} · {{ err.mode }} · {{ err.error }}
          </div>
        </div>
      </div>
    </section>

    <section class="chart-grid">
      <KlineChart :data="stockStore.klines" :subtitle="selectedStockLabel" />
      <EquityChart :data="backtestStore.points" :subtitle="selectedStrategyLabel" />
    </section>

    <section class="panel-grid">
      <div class="card">
        <h2>交易记录</h2>
        <ul class="trade-list">
          <li v-for="trade in backtestStore.trades" :key="trade.time + trade.side">
            {{ new Date(trade.time).toLocaleDateString() }} · {{ trade.side }} · {{ trade.price.toFixed(2) }} · {{ trade.shares.toFixed(0) }} 股
          </li>
        </ul>
        <div v-if="!backtestStore.trades.length" class="footer-note">暂无交易记录。</div>
      </div>
      <div class="card">
        <h2>系统提示</h2>
        <p class="footer-note">
          默认策略为 MA 交叉逻辑。你可以通过新增策略修改短期/长期窗口，或在后端扩展更多策略类型。
        </p>
        <p class="footer-note">
          目前示例数据使用内置随机生成器模拟行情，方便本地验证流程。
        </p>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import api from './api/client'
import { useStrategyStore } from './stores/strategy'
import { useStockStore } from './stores/stock'
import { useBacktestStore } from './stores/backtest'
import KlineChart from './components/KlineChart.vue'
import EquityChart from './components/EquityChart.vue'

const strategyStore = useStrategyStore()
const stockStore = useStockStore()
const backtestStore = useBacktestStore()

const selectedStrategyId = ref('')
const selectedStockCode = ref('')
const initialCapital = ref(100000)
const screeningResults = ref([])
const syncLoading = ref(false)
const syncResult = ref(null)
const syncError = ref('')

const form = reactive({
  name: '',
  description: '',
  params: '{"short_window":5,"long_window":20}'
})

const syncForm = reactive({
  symbols: '',
  mode: 'all',
  startDate: '',
  endDate: '',
  minStart: '',
  minEnd: '',
  period: '30',
  limit: 50
})

const selectedStrategyLabel = computed(() => {
  const strategy = strategyStore.strategies.find((item) => item.id === selectedStrategyId.value)
  return strategy ? strategy.name : '策略未选择'
})

const selectedStockLabel = computed(() => {
  const stock = stockStore.stocks.find((item) => item.code === selectedStockCode.value)
  return stock ? `${stock.name} (${stock.code})` : '股票未选择'
})

const syncErrors = computed(() => {
  if (!syncResult.value || !Array.isArray(syncResult.value.errors)) return []
  return syncResult.value.errors
})

const refreshAll = async () => {
  await Promise.all([strategyStore.fetchStrategies(), stockStore.fetchStocks()])
  if (!selectedStrategyId.value && strategyStore.strategies.length) {
    selectedStrategyId.value = strategyStore.strategies[0].id
  }
  if (!selectedStockCode.value && stockStore.stocks.length) {
    selectedStockCode.value = stockStore.stocks[0].code
    await stockStore.fetchKlines(selectedStockCode.value)
  }
}

const seedDemo = async () => {
  await api.post('/demo/seed')
  await refreshAll()
}

const createStrategy = async () => {
  if (!form.name) return
  await strategyStore.createStrategy({
    name: form.name,
    description: form.description,
    type: 'ma_crossover',
    params_json: form.params
  })
  form.name = ''
  form.description = ''
}

const runScreening = async () => {
  if (!selectedStrategyId.value) return
  const { data } = await api.post('/screen', { strategy_id: selectedStrategyId.value })
  screeningResults.value = data
}

const runBacktest = async () => {
  if (!selectedStrategyId.value || !selectedStockCode.value) return
  await backtestStore.runBacktest({
    strategy_id: selectedStrategyId.value,
    stock_code: selectedStockCode.value,
    initial_capital: initialCapital.value
  })
}

const runAkshareSync = async () => {
  syncError.value = ''
  syncResult.value = null
  syncLoading.value = true

  try {
    const symbols = syncForm.symbols
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean)
    const payload = {
      mode: syncForm.mode
    }
    if (symbols.length) {
      payload.symbols = symbols
    } else if (syncForm.limit) {
      payload.limit = syncForm.limit
    }
    if (syncForm.startDate) payload.start_date = syncForm.startDate
    if (syncForm.endDate) payload.end_date = syncForm.endDate
    if (syncForm.minStart) payload.min_start = syncForm.minStart
    if (syncForm.minEnd) payload.min_end = syncForm.minEnd
    if (syncForm.period) payload.period = syncForm.period

    const { data } = await api.post('/sync/akshare', payload, { timeout: 600000 })
    syncResult.value = data.summary
    await refreshAll()
  } catch (err) {
    syncError.value = err?.response?.data?.error || err?.message || '同步失败'
  } finally {
    syncLoading.value = false
  }
}

onMounted(refreshAll)

watch(selectedStockCode, async (code) => {
  await stockStore.fetchKlines(code)
})
</script>

<style scoped>
.footer-note {
  margin: 0;
}
</style>
