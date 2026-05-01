import { useEffect, useRef, useState } from 'react'
import { useStore } from '../store'
import { getHistory, getGroupHistory } from '../api'
import { Send, Hash, User } from 'lucide-react'

export default function ChatWindow({ send }) {
  const { state, dispatch } = useStore()
  const { activeId, conversations, user } = state
  const conv = conversations[activeId]
  const [input, setInput] = useState('')
  const bottomRef = useRef(null)
  const seqRef = useRef(1)

  // 加载历史消息
  useEffect(() => {
    if (!activeId || !conv) return
    const load = async () => {
      try {
        const res = conv.type === 'group'
          ? await getGroupHistory(activeId)
          : await getHistory(user.uid, activeId)
        dispatch({ type: 'SET_HISTORY', convId: activeId, messages: res.data || [] })
      } catch {}
    }
    load()
  }, [activeId])

  // 滚动到底部
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [conv?.messages])

  const sendMsg = e => {
    e.preventDefault()
    if (!input.trim() || !activeId) return
    const msg = {
      msg_type: 1,
      target_type: conv.type === 'group' ? 2 : 1,
      to_id: activeId,
      content: input.trim(),
      timestamp: Date.now(),
      seq: seqRef.current++,
    }
    send(msg)
    // 本地先显示
    dispatch({
      type: 'ADD_MESSAGE',
      convId: activeId,
      msg: { ...msg, from_uid: user.uid, msg_id: Date.now() },
    })
    setInput('')
  }

  if (!activeId || !conv) {
    return (
      <div className="flex-1 flex items-center justify-center bg-slate-850">
        <div className="text-center text-slate-600">
          <div className="w-20 h-20 bg-slate-800 rounded-3xl flex items-center justify-center mx-auto mb-4">
            <Send className="w-8 h-8 opacity-30" />
          </div>
          <p className="text-lg font-medium text-slate-500">选择一个会话开始聊天</p>
          <p className="text-sm text-slate-600 mt-1">或点击左侧 + 新建会话</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 flex flex-col bg-slate-850 min-w-0">
      {/* 顶部栏 */}
      <div className="px-6 py-4 border-b border-slate-700/50 bg-slate-900 flex items-center gap-3">
        <div className={`w-9 h-9 rounded-xl flex items-center justify-center ${
          conv.type === 'group' ? 'bg-purple-500/20 text-purple-400' : 'bg-indigo-500/20 text-indigo-400'
        }`}>
          {conv.type === 'group' ? <Hash className="w-4 h-4" /> : <User className="w-4 h-4" />}
        </div>
        <div>
          <p className="text-white font-medium">{conv.name}</p>
          <p className="text-slate-500 text-xs">{conv.type === 'group' ? '群聊' : `UID: ${conv.id}`}</p>
        </div>
      </div>

      {/* 消息列表 */}
      <div className="flex-1 overflow-y-auto px-6 py-4 space-y-3">
        {conv.messages.map((msg, i) => {
          const isMine = msg.from_uid === user.uid
          const time = new Date(msg.timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
          return (
            <div key={msg.msg_id || i} className={`flex items-end gap-2 ${isMine ? 'flex-row-reverse' : 'flex-row'}`}>
              {/* 头像 */}
              <div className={`w-8 h-8 rounded-xl flex items-center justify-center text-xs font-bold flex-shrink-0 ${
                isMine ? 'bg-indigo-500 text-white' : 'bg-slate-700 text-slate-300'
              }`}>
                {isMine ? user.username[0]?.toUpperCase() : String(msg.from_uid || conv.name[0])?.toUpperCase()}
              </div>
              {/* 气泡 */}
              <div className={`max-w-xs lg:max-w-md ${isMine ? 'items-end' : 'items-start'} flex flex-col gap-1`}>
                <div className={`px-4 py-2.5 rounded-2xl text-sm leading-relaxed ${
                  isMine
                    ? 'bg-indigo-500 text-white rounded-br-sm'
                    : 'bg-slate-700 text-slate-100 rounded-bl-sm'
                }`}>
                  {msg.content}
                </div>
                <span className="text-slate-600 text-xs px-1">{time}</span>
              </div>
            </div>
          )
        })}
        <div ref={bottomRef} />
      </div>

      {/* 输入框 */}
      <div className="px-6 py-4 border-t border-slate-700/50 bg-slate-900">
        <form onSubmit={sendMsg} className="flex items-center gap-3">
          <input
            className="flex-1 bg-slate-700 border border-slate-600 rounded-xl px-4 py-3 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition text-sm"
            placeholder={`发送消息给 ${conv.name}...`}
            value={input}
            onChange={e => setInput(e.target.value)}
          />
          <button
            type="submit"
            disabled={!input.trim()}
            className="w-11 h-11 bg-indigo-500 hover:bg-indigo-600 disabled:opacity-40 disabled:cursor-not-allowed rounded-xl flex items-center justify-center transition-all shadow-lg shadow-indigo-500/25"
          >
            <Send className="w-4 h-4 text-white" />
          </button>
        </form>
      </div>
    </div>
  )
}
