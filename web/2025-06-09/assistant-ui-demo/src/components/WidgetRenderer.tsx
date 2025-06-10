import React from 'react'
import { useDispatch } from 'react-redux'
import { UIWidget, TodoItem, DropdownOption, updateTodoItem } from '../store/chatSlice'
import TodoWidget from './TodoWidget'
import DropdownWidget from './DropdownWidget'

interface Props {
  widget: UIWidget
  messageId: string
}

const WidgetRenderer: React.FC<Props> = ({ widget, messageId }) => {
  switch (widget.type) {
    case 'todo':
      return <TodoWidget widget={widget} messageId={messageId} />
    case 'dropdown':
      return <DropdownWidget widget={widget} messageId={messageId} />
    default:
      return null
  }
}

export default WidgetRenderer
