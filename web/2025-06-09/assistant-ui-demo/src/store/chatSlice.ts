import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
  streaming?: boolean
  widgets?: UIWidget[]
}

export interface TodoItem {
  id: string
  text: string
  completed: boolean
}

export interface DropdownOption {
  value: string
  label: string
}

export interface UIWidget {
  type: 'todo' | 'dropdown'
  id: string
  title: string
  data: TodoItem[] | DropdownOption[]
}

interface ChatState {
  messages: ChatMessage[]
  isConnected: boolean
  isTyping: boolean
}

const initialState: ChatState = {
  messages: [],
  isConnected: false,
  isTyping: false,
}

const chatSlice = createSlice({
  name: 'chat',
  initialState,
  reducers: {
    setConnected: (state, action: PayloadAction<boolean>) => {
      state.isConnected = action.payload
    },
    addMessage: (state, action: PayloadAction<ChatMessage>) => {
      const existingIndex = state.messages.findIndex(msg => msg.id === action.payload.id)
      if (existingIndex >= 0) {
        state.messages[existingIndex] = { ...action.payload, streaming: false }
      } else {
        state.messages.push(action.payload)
      }
      state.isTyping = false
    },
    updateStreamingMessage: (state, action: PayloadAction<{ id: string; chunk: string }>) => {
      const existingIndex = state.messages.findIndex(msg => msg.id === action.payload.id)
      if (existingIndex >= 0) {
        state.messages[existingIndex].content += action.payload.chunk
      } else {
        state.messages.push({
          id: action.payload.id,
          role: 'assistant',
          content: action.payload.chunk,
          timestamp: new Date(),
          streaming: true
        })
      }
    },
    addWidgetToMessage: (state, action: PayloadAction<{ messageId: string; widget: UIWidget }>) => {
      const message = state.messages.find(m => m.id === action.payload.messageId)
      if (message) {
        if (!message.widgets) {
          message.widgets = []
        }
        message.widgets.push(action.payload.widget)
      }
    },
    updateTodoItem: (state, action: PayloadAction<{ messageId: string; widgetId: string; todoId: string; completed: boolean }>) => {
      const message = state.messages.find(m => m.id === action.payload.messageId)
      if (message?.widgets) {
        const widget = message.widgets.find(w => w.id === action.payload.widgetId)
        if (widget && widget.type === 'todo') {
          const todos = widget.data as TodoItem[]
          const todo = todos.find(t => t.id === action.payload.todoId)
          if (todo) {
            todo.completed = action.payload.completed
          }
        }
      }
    },
    setTyping: (state, action: PayloadAction<boolean>) => {
      state.isTyping = action.payload
    },
  },
})

export const { 
  setConnected, 
  addMessage, 
  updateStreamingMessage, 
  addWidgetToMessage, 
  updateTodoItem,
  setTyping 
} = chatSlice.actions

export default chatSlice.reducer
