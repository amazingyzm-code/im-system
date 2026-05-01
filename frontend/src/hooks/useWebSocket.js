import { useEffect, useRef, useCallback } from 'react'

const MSG_TYPE_TEXT      = 1
const MSG_TYPE_ACK       = 3
const MSG_TYPE_HEARTBEAT = 4
const MSG_TYPE_AUTH      = 5

export function useWebSocket(token, onMessage) {
  const ws = useRef(null)
  const heartbeat = useRef(null)
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  const connect = useCallback(() => {
    if (!token) return
    const socket = new WebSocket(`ws://${location.host}/ws`)
    ws.current = socket

    socket.onopen = () => {
      // 第一包发鉴权
      socket.send(JSON.stringify({ msg_type: MSG_TYPE_AUTH, token }))
      // 心跳
      heartbeat.current = setInterval(() => {
        if (socket.readyState === WebSocket.OPEN) {
          socket.send(JSON.stringify({ msg_type: MSG_TYPE_HEARTBEAT }))
        }
      }, 25000)
    }

    socket.onmessage = e => {
      try {
        const msg = JSON.parse(e.data)
        if (msg.msg_type !== MSG_TYPE_ACK) {
          onMessageRef.current?.(msg)
        }
      } catch {}
    }

    socket.onclose = () => {
      clearInterval(heartbeat.current)
      // 3秒后重连
      setTimeout(connect, 3000)
    }
  }, [token])

  useEffect(() => {
    connect()
    return () => {
      clearInterval(heartbeat.current)
      ws.current?.close()
    }
  }, [connect])

  const send = useCallback((msg) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(msg))
    }
  }, [])

  return { send }
}
