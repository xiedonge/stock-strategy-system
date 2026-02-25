<template>
  <div class="chart-card">
    <div class="chart-header">
      <div>
        <h3>回测权益曲线</h3>
        <p v-if="subtitle">{{ subtitle }}</p>
      </div>
      <span class="badge">回测</span>
    </div>
    <div ref="chartEl" class="chart"></div>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  data: { type: Array, default: () => [] },
  subtitle: { type: String, default: '' }
})

const chartEl = ref(null)
let chartInstance = null

const buildOption = (rows) => {
  const categories = rows.map((row) => new Date(row.time).toLocaleDateString())
  const values = rows.map((row) => row.equity)

  return {
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis' },
    grid: { left: 16, right: 16, top: 30, bottom: 24, containLabel: true },
    xAxis: {
      type: 'category',
      data: categories,
      axisLabel: { color: '#55606a' },
      axisLine: { lineStyle: { color: '#cbd2d9' } }
    },
    yAxis: {
      scale: true,
      axisLabel: { color: '#55606a' },
      splitLine: { lineStyle: { color: '#e7ebf0' } }
    },
    series: [
      {
        type: 'line',
        data: values,
        smooth: true,
        symbol: 'none',
        lineStyle: { width: 3, color: '#2c6ce5' },
        areaStyle: { color: 'rgba(44, 108, 229, 0.12)' }
      }
    ]
  }
}

const renderChart = () => {
  if (!chartInstance || !chartEl.value) return
  chartInstance.setOption(buildOption(props.data || []), true)
}

onMounted(() => {
  if (!chartEl.value) return
  chartInstance = echarts.init(chartEl.value)
  renderChart()
  window.addEventListener('resize', resizeChart)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeChart)
  if (chartInstance) chartInstance.dispose()
})

const resizeChart = () => {
  if (chartInstance) chartInstance.resize()
}

watch(() => props.data, () => renderChart(), { deep: true })
</script>

<style scoped>
.chart-card {
  background: #fff;
  border-radius: 18px;
  padding: 18px;
  box-shadow: 0 20px 40px -28px rgba(15, 23, 42, 0.35);
}

.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.chart-header h3 {
  margin: 0;
  font-size: 1.1rem;
}

.chart-header p {
  margin: 4px 0 0;
  color: #6b7280;
  font-size: 0.85rem;
}

.badge {
  padding: 4px 10px;
  border-radius: 999px;
  background: #f2f6ff;
  color: #2c6ce5;
  font-size: 0.75rem;
  font-weight: 600;
}

.chart {
  height: 260px;
}
</style>
