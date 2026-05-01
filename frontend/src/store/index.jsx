import { createContext, useContext, useReducer } from 'react'

const Ctx = createContext(null)

const init = {
  user: JSON.parse(localStorage.getItem('user') || 'null'),
  token: localStorage.getItem('token') || '',
  // conversations: { [id]: { id, name, type: 'single'|'group', messages: [], unread: 0 } }
  conversations: {},
  activeId: null,
}

function reducer(state, action) {
  switch (action.type) {
    case 'LOGIN':
      localStorage.setItem('token', action.token)
      localStorage.setItem('user', JSON.stringify(action.user))
      return { ...state, user: action.user, token: action.token }

    case 'LOGOUT':
      localStorage.clear()
      return { ...init, user: null, token: '' }

    case 'SET_ACTIVE':
      return {
        ...state,
        activeId: action.id,
        conversations: {
          ...state.conversations,
          [action.id]: { ...state.conversations[action.id], unread: 0 },
        },
      }

    case 'ENSURE_CONV': {
      if (state.conversations[action.id]) return state
      return {
        ...state,
        conversations: {
          ...state.conversations,
          [action.id]: { id: action.id, name: action.name, type: action.convType, messages: [], unread: 0 },
        },
      }
    }

    case 'ADD_MESSAGE': {
      const { convId, msg } = action
      const conv = state.conversations[convId] || { id: convId, name: String(convId), type: 'single', messages: [], unread: 0 }
      const isActive = state.activeId === convId
      return {
        ...state,
        conversations: {
          ...state.conversations,
          [convId]: {
            ...conv,
            messages: [...conv.messages, msg],
            unread: isActive ? 0 : conv.unread + 1,
          },
        },
      }
    }

    case 'SET_HISTORY': {
      const { convId, messages } = action
      const conv = state.conversations[convId] || { id: convId, name: String(convId), type: 'single', messages: [], unread: 0 }
      return {
        ...state,
        conversations: {
          ...state.conversations,
          [convId]: { ...conv, messages },
        },
      }
    }

    default:
      return state
  }
}

export function StoreProvider({ children }) {
  const [state, dispatch] = useReducer(reducer, init)
  return <Ctx.Provider value={{ state, dispatch }}>{children}</Ctx.Provider>
}

export const useStore = () => useContext(Ctx)
