import React from 'react'
import { useDispatch } from 'react-redux'
import { UIWidget, TodoItem, updateTodoItem } from '../store/chatSlice'

interface Props {
  widget: UIWidget
  messageId: string
}

const TodoWidget: React.FC<Props> = ({ widget, messageId }) => {
  const dispatch = useDispatch()
  const todos = widget.data as TodoItem[]

  const handleToggle = (todoId: string, completed: boolean) => {
    dispatch(updateTodoItem({
      messageId,
      widgetId: widget.id,
      todoId,
      completed
    }))
  }

  return (
    <div className="widget">
      <div className="widget-title">{widget.title}</div>
      <div className="todo-list">
        {todos.map((todo) => (
          <div key={todo.id} className="todo-item">
            <input
              type="checkbox"
              className="todo-checkbox form-check-input"
              checked={todo.completed}
              onChange={(e) => handleToggle(todo.id, e.target.checked)}
            />
            <span className={`todo-text ${todo.completed ? 'completed' : ''}`}>
              {todo.text}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

export default TodoWidget
