import axios from 'axios'

const http = axios.create({ baseURL: '/api' })

http.interceptors.request.use(cfg => {
  const token = localStorage.getItem('token')
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  return cfg
})

export const register = (username, password) =>
  http.post('/user/register', { username, password })

export const login = (username, password) =>
  http.post('/user/login', { username, password })

export const getHistory = (uid1, uid2, lastId = 0) =>
  http.get('/history/single', { params: { uid1, uid2, last_id: lastId, limit: 30 } })

export const getGroupHistory = (groupId, lastId = 0) =>
  http.get('/history/group', { params: { group_id: groupId, last_id: lastId, limit: 30 } })

export const joinGroup = (groupId, uid) =>
  http.post(`/group/join?group_id=${groupId}&uid=${uid}`)

export const leaveGroup = (groupId, uid) =>
  http.post(`/group/leave?group_id=${groupId}&uid=${uid}`)
