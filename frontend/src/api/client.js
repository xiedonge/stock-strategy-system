import axios from 'axios'

// Use VITE_API_BASE to point to the Go backend.
const apiBase = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api'

const client = axios.create({
  baseURL: apiBase,
  timeout: 15000
})

export default client
