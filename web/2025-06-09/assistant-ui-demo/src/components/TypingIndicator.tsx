import React from 'react'

const TypingIndicator: React.FC = () => {
  return (
    <div className="message assistant">
      <div className="d-flex align-items-center">
        <span className="me-2">Assistant is typing</span>
        <div className="typing-indicator me-1"></div>
        <div className="typing-indicator me-1"></div>
        <div className="typing-indicator"></div>
      </div>
    </div>
  )
}

export default TypingIndicator
