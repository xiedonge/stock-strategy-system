import { defineStore } from 'pinia'
import api from '../api/client'

// Strategy store keeps strategy data in memory.
export const useStrategyStore = defineStore('strategy', {
  state: () => ({
    strategies: [],
    loading: false
  }),
  actions: {
    async fetchStrategies() {
      this.loading = true
      try {
        const { data } = await api.get('/strategies')
        this.strategies = data
      } finally {
        this.loading = false
      }
    },
    async createStrategy(payload) {
      const { data } = await api.post('/strategies', payload)
      this.strategies.unshift(data)
      return data
    }
  }
})
