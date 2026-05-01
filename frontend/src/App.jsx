import { StoreProvider, useStore } from './store'
import AuthPage from './pages/AuthPage'
import ChatPage from './pages/ChatPage'

function App() {
  const { state } = useStore()
  return state.token ? <ChatPage /> : <AuthPage />
}

export default function Root() {
  return (
    <StoreProvider>
      <App />
    </StoreProvider>
  )
}
