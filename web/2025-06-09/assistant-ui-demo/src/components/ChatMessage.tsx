import React from 'react'
import { ChatMessage as ChatMessageType } from '../store/chatSlice'

interface Props {
  message: ChatMessageType
}

const ChatMessage: React.FC<Props> = ({ message }) => {
  const messageClass = `message ${message.role} ${message.streaming ? 'streaming' : ''}`
  
  return (
    <div className={messageClass}>
      <div className="message-content">
        {message.content}
      </div>
      <small className="text-muted d-block mt-1">
        {message.timestamp.toLocaleTimeString()}
      </small>
    </div>
  )
}

export default ChatMessage
