import { useStore } from '../store'
import { useWebSocket } from '../hooks/useWebSocket'
import Sidebar from '../components/Sidebar'
import ChatWindow from '../components/ChatWindow'

export default function ChatPage() {
  const { state, dispatch } = useStore()
  const { token, user, conversations } = state

  const { send } = useWebSocket(token, msg => {
    // 收到消息，判断会话 ID
    const convId = msg.target_type === 2 ? msg.to_id : msg.from_uid
    // 如果会话不存在，自动创建
    if (!conversations[convId]) {
      dispatch({
        type: 'ENSURE_CONV',
        id: convId,
        name: msg.target_type === 2 ? `群组 ${convId}` : `用户 ${convId}`,
        convType: msg.target_type === 2 ? 'group' : 'single',
      })
    }
    dispatch({ type: 'ADD_MESSAGE', convId, msg })
  })

  return (
    <div className="h-screen flex bg-slate-900 overflow-hidden">
      <Sidebar />
      <ChatWindow send={send} />
    </div>
  )
}
