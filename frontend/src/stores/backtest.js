import { defineStore } from 'pinia'
import api from '../api/client'

// Backtest store keeps last backtest results.
export const useBacktestStore = defineStore('backtest', {
  state: () => ({
    summary: null,
    points: [],
    trades: [],
    loading: false
  }),
  actions: {
    async runBacktest(payload) {
      this.loading = true
      try {
        const { data } = await api.post('/backtest', payload)
        this.summary = data.summary
        this.points = data.points
        this.trades = data.trades
        return data
      } finally {
        this.loading = false
      }
    }
  }
})
