import React, { useState } from 'react'
import { UIWidget, DropdownOption } from '../store/chatSlice'

interface Props {
  widget: UIWidget
  messageId: string
}

const DropdownWidget: React.FC<Props> = ({ widget, messageId }) => {
  const [selectedValue, setSelectedValue] = useState('')
  const options = widget.data as DropdownOption[]

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedValue(e.target.value)
    console.log('Selected:', e.target.value)
  }

  return (
    <div className="widget">
      <div className="widget-title">{widget.title}</div>
      <div className="dropdown-widget">
        <select
          className="form-select"
          value={selectedValue}
          onChange={handleChange}
        >
          <option value="">Select an option...</option>
          {options.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
        {selectedValue && (
          <div className="mt-2">
            <small className="text-muted">Selected: {selectedValue}</small>
          </div>
        )}
      </div>
    </div>
  )
}

export default DropdownWidget
