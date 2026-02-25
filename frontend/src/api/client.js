import axios from 'axios'

// Use VITE_API_BASE to point to the Go backend.
// Default to current host with backend port 8080 to support remote access.
const fallbackBase = typeof window !== 'undefined'
  ? `${window.location.protocol}//${window.location.hostname}:8080/api`
  : 'http://localhost:8080/api'
const apiBase = import.meta.env.VITE_API_BASE || fallbackBase

const client = axios.create({
  baseURL: apiBase,
  timeout: 15000
})

export default client
