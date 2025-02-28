import axios from 'axios'

var client = axios.create({
  baseURL: '/api/v2',
  headers: { common: { 'Content-Type': 'application/json' } }
})

export default client