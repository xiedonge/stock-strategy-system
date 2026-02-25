import { defineStore } from 'pinia'
import api from '../api/client'

// Stock store manages stock list and kline data.
export const useStockStore = defineStore('stock', {
  state: () => ({
    stocks: [],
    klines: [],
    loading: false
  }),
  actions: {
    async fetchStocks() {
      this.loading = true
      try {
        const { data } = await api.get('/stocks')
        this.stocks = data
      } finally {
        this.loading = false
      }
    },
    async fetchKlines(code) {
      if (!code) return
      const { data } = await api.get(`/stocks/${code}/klines`, { params: { interval: '1d', limit: 200 } })
      this.klines = data
    }
  }
})
