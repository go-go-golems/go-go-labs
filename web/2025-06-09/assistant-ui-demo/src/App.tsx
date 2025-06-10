import React, { useEffect, useRef, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { RootState } from './store/store'
import { 
  setConnected, 
  addMessage, 
  updateStreamingMessage, 
  addWidgetToMessage, 
  setTyping,
  ChatMessage as ChatMessageType
} from './store/chatSlice'
import ChatMessage from './components/ChatMessage'
import WidgetRenderer from './components/WidgetRenderer'
import TypingIndicator from './components/TypingIndicator'

function App() {
  const dispatch = useDispatch()
  const { messages, isConnected, isTyping } = useSelector((state: RootState) => state.chat)
  const [input, setInput] = useState('')
  const [ws, setWs] = useState<WebSocket | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const websocket = new WebSocket(`ws://${window.location.host}/ws`)
    
    websocket.onopen = () => {
      console.log('WebSocket connected')
      dispatch(setConnected(true))
      setWs(websocket)
    }
    
    websocket.onclose = () => {
      console.log('WebSocket disconnected')
      dispatch(setConnected(false))
      setWs(null)
    }
    
    websocket.onmessage = (event) => {
      const data = JSON.parse(event.data)
      
      switch (data.type) {
        case 'message':
          dispatch(addMessage({
            ...data.content,
            timestamp: new Date(data.content.timestamp)
          }))
          break
        case 'chunk':
          dispatch(updateStreamingMessage(data.content))
          break
        case 'widget':
          dispatch(addWidgetToMessage({
            messageId: data.content.messageId,
            widget: data.content.widget
          }))
          break
      }
    }
    
    return () => {
      websocket.close()
    }
  }, [dispatch])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, isTyping])

  const sendMessage = () => {
    if (!input.trim() || !ws) return

    const message: ChatMessageType = {
      id: `msg_${Date.now()}`,
      role: 'user',
      content: input.trim(),
      timestamp: new Date()
    }

    dispatch(addMessage(message))
    dispatch(setTyping(true))

    ws.send(JSON.stringify({
      type: 'chat',
      content: message
    }))

    setInput('')
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  return (
    <div className="container-fluid chat-container">
      <div className="row h-100">
        <div className="col-12">
          <div className="d-flex flex-column h-100">
            {/* Header */}
            <div className="bg-primary text-white p-3">
              <h4 className="mb-0">Assistant UI Demo</h4>
              <small>
                Status: {isConnected ? 
                  <span className="badge bg-success">Connected</span> : 
                  <span className="badge bg-danger">Disconnected</span>
                }
              </small>
            </div>

            {/* Messages */}
            <div className="messages-container">
              {messages.map((message) => (
                <div key={message.id}>
                  <ChatMessage message={message} />
                  {message.widgets?.map((widget) => (
                    <WidgetRenderer key={widget.id} widget={widget} messageId={message.id} />
                  ))}
                </div>
              ))}
              
              {isTyping && <TypingIndicator />}
              <div ref={messagesEndRef} />
            </div>

            {/* Input */}
            <div className="input-container">
              <div className="input-group">
                <textarea
                  className="form-control"
                  placeholder="Type your message..."
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyPress={handleKeyPress}
                  rows={2}
                  disabled={!isConnected}
                />
                <button
                  className="btn btn-primary"
                  onClick={sendMessage}
                  disabled={!isConnected || !input.trim()}
                >
                  Send
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
