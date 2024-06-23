class Test2Prompto < Prompto
  text 'Test string', text: 'Greetings from Test2Prompto!'
  ruby 'Data Analysis', description: 'Performs a simple data analysis.' do
    data = [1, 2, 3, 4, 5]
    "Mean: #{data.sum.to_f / data.size}, Median: #{data.sort[data.size / 2]}"
  end
  shell 'System Info', description: 'Displays basic system information.' do
    <<~BASH
      #!/bin/bash
      echo "OS Information:"
      uname -a
      echo "CPU Information:"
      lscpu | grep "Model name"
      echo "Memory Information:"
      free -h
    BASH
  end

  define_thor_command :test2
end