import { useState } from 'react'
import { login, register } from '../api'
import { useStore } from '../store'
import { MessageCircle } from 'lucide-react'

export default function AuthPage() {
  const { dispatch } = useStore()
  const [isLogin, setIsLogin] = useState(true)
  const [form, setForm] = useState({ username: '', password: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async e => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      if (!isLogin) await register(form.username, form.password)
      const res = await login(form.username, form.password)
      const token = res.data.token
      const payload = JSON.parse(atob(token.split('.')[1]))
      dispatch({ type: 'LOGIN', token, user: { uid: payload.uid, username: form.username } })
    } catch (err) {
      setError(err.response?.data || '操作失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-900 via-purple-900 to-slate-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-indigo-500 rounded-2xl mb-4 shadow-lg">
            <MessageCircle className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-white">IM System</h1>
          <p className="text-slate-400 mt-1">高性能即时通讯系统</p>
        </div>

        <div className="bg-slate-800 rounded-2xl shadow-2xl p-8 border border-slate-700">
          <div className="flex bg-slate-700 rounded-xl p-1 mb-6">
            <button
              onClick={() => setIsLogin(true)}
              className={`flex-1 py-2 rounded-lg text-sm font-medium transition-all ${isLogin ? 'bg-indigo-500 text-white shadow' : 'text-slate-400 hover:text-white'}`}
            >登录</button>
            <button
              onClick={() => setIsLogin(false)}
              className={`flex-1 py-2 rounded-lg text-sm font-medium transition-all ${!isLogin ? 'bg-indigo-500 text-white shadow' : 'text-slate-400 hover:text-white'}`}
            >注册</button>
          </div>

          <form onSubmit={submit} className="space-y-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1">用户名</label>
              <input
                className="w-full bg-slate-700 border border-slate-600 rounded-xl px-4 py-3 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
                placeholder="输入用户名"
                value={form.username}
                onChange={e => setForm({ ...form, username: e.target.value })}
                required
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">密码</label>
              <input
                type="password"
                className="w-full bg-slate-700 border border-slate-600 rounded-xl px-4 py-3 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
                placeholder="输入密码"
                value={form.password}
                onChange={e => setForm({ ...form, password: e.target.value })}
                required
              />
            </div>

            {error && <p className="text-red-400 text-sm bg-red-400/10 rounded-lg px-3 py-2">{error}</p>}

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-indigo-500 hover:bg-indigo-600 disabled:opacity-50 text-white font-medium py-3 rounded-xl transition-all shadow-lg shadow-indigo-500/25"
            >
              {loading ? '请稍候...' : isLogin ? '登录' : '注册'}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
