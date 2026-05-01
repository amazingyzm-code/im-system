import { useState } from 'react'
import { useStore } from '../store'
import { joinGroup } from '../api'
import { MessageCircle, Users, LogOut, Plus, Hash } from 'lucide-react'

export default function Sidebar() {
  const { state, dispatch } = useStore()
  const { conversations, activeId, user } = state
  const [showAdd, setShowAdd] = useState(false)
  const [addForm, setAddForm] = useState({ type: 'single', id: '', name: '' })

  const convList = Object.values(conversations).sort((a, b) => {
    const aLast = a.messages.at(-1)?.timestamp || 0
    const bLast = b.messages.at(-1)?.timestamp || 0
    return bLast - aLast
  })

  const handleAdd = async e => {
    e.preventDefault()
    const id = parseInt(addForm.id)
    if (!id) return
    if (addForm.type === 'group') {
      await joinGroup(id, user.uid)
    }
    dispatch({ type: 'ENSURE_CONV', id, name: addForm.name || String(id), convType: addForm.type })
    dispatch({ type: 'SET_ACTIVE', id })
    setShowAdd(false)
    setAddForm({ type: 'single', id: '', name: '' })
  }

  return (
    <div className="w-72 bg-slate-900 flex flex-col border-r border-slate-700/50">
      {/* Header */}
      <div className="p-4 border-b border-slate-700/50">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 bg-indigo-500 rounded-xl flex items-center justify-center font-bold text-white text-sm">
              {user?.username?.[0]?.toUpperCase()}
            </div>
            <div>
              <p className="text-white font-medium text-sm">{user?.username}</p>
              <p className="text-slate-500 text-xs">UID: {user?.uid}</p>
            </div>
          </div>
          <button
            onClick={() => dispatch({ type: 'LOGOUT' })}
            className="text-slate-500 hover:text-red-400 transition p-1 rounded-lg hover:bg-slate-800"
          >
            <LogOut className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* 会话列表标题 */}
      <div className="flex items-center justify-between px-4 py-3">
        <span className="text-slate-400 text-xs font-semibold uppercase tracking-wider">消息</span>
        <button
          onClick={() => setShowAdd(true)}
          className="text-slate-500 hover:text-indigo-400 transition p-1 rounded-lg hover:bg-slate-800"
        >
          <Plus className="w-4 h-4" />
        </button>
      </div>

      {/* 会话列表 */}
      <div className="flex-1 overflow-y-auto space-y-0.5 px-2">
        {convList.length === 0 && (
          <div className="text-center text-slate-600 text-sm py-12">
            <MessageCircle className="w-8 h-8 mx-auto mb-2 opacity-50" />
            <p>点击 + 开始聊天</p>
          </div>
        )}
        {convList.map(conv => {
          const lastMsg = conv.messages.at(-1)
          const isActive = activeId === conv.id
          return (
            <button
              key={conv.id}
              onClick={() => dispatch({ type: 'SET_ACTIVE', id: conv.id })}
              className={`w-full flex items-center gap-3 px-3 py-3 rounded-xl transition-all text-left ${
                isActive ? 'bg-indigo-500/20 border border-indigo-500/30' : 'hover:bg-slate-800'
              }`}
            >
              <div className={`w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 ${
                conv.type === 'group' ? 'bg-purple-500/20 text-purple-400' : 'bg-indigo-500/20 text-indigo-400'
              }`}>
                {conv.type === 'group' ? <Hash className="w-5 h-5" /> : <span className="font-bold text-sm">{conv.name[0]?.toUpperCase()}</span>}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <span className="text-white text-sm font-medium truncate">{conv.name}</span>
                  {conv.unread > 0 && (
                    <span className="bg-indigo-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center flex-shrink-0">
                      {conv.unread > 9 ? '9+' : conv.unread}
                    </span>
                  )}
                </div>
                <p className="text-slate-500 text-xs truncate mt-0.5">
                  {lastMsg ? lastMsg.content : '暂无消息'}
                </p>
              </div>
            </button>
          )
        })}
      </div>

      {/* 添加会话弹窗 */}
      {showAdd && (
        <div className="absolute inset-0 bg-black/60 flex items-center justify-center z-50" onClick={() => setShowAdd(false)}>
          <div className="bg-slate-800 rounded-2xl p-6 w-80 border border-slate-700 shadow-2xl" onClick={e => e.stopPropagation()}>
            <h3 className="text-white font-semibold mb-4">新建会话</h3>
            <form onSubmit={handleAdd} className="space-y-3">
              <div className="flex bg-slate-700 rounded-xl p-1">
                <button type="button" onClick={() => setAddForm({ ...addForm, type: 'single' })}
                  className={`flex-1 py-2 rounded-lg text-sm transition-all flex items-center justify-center gap-1 ${addForm.type === 'single' ? 'bg-indigo-500 text-white' : 'text-slate-400'}`}>
                  <MessageCircle className="w-3.5 h-3.5" /> 单聊
                </button>
                <button type="button" onClick={() => setAddForm({ ...addForm, type: 'group' })}
                  className={`flex-1 py-2 rounded-lg text-sm transition-all flex items-center justify-center gap-1 ${addForm.type === 'group' ? 'bg-purple-500 text-white' : 'text-slate-400'}`}>
                  <Users className="w-3.5 h-3.5" /> 群聊
                </button>
              </div>
              <input
                className="w-full bg-slate-700 border border-slate-600 rounded-xl px-4 py-2.5 text-white text-sm placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                placeholder={addForm.type === 'single' ? '对方 UID' : '群组 ID'}
                value={addForm.id}
                onChange={e => setAddForm({ ...addForm, id: e.target.value })}
                required
              />
              <input
                className="w-full bg-slate-700 border border-slate-600 rounded-xl px-4 py-2.5 text-white text-sm placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                placeholder="备注名（可选）"
                value={addForm.name}
                onChange={e => setAddForm({ ...addForm, name: e.target.value })}
              />
              <div className="flex gap-2 pt-1">
                <button type="button" onClick={() => setShowAdd(false)}
                  className="flex-1 py-2.5 rounded-xl text-slate-400 hover:text-white bg-slate-700 hover:bg-slate-600 text-sm transition">取消</button>
                <button type="submit"
                  className="flex-1 py-2.5 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white text-sm transition">确认</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
