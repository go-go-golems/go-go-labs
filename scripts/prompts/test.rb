# Example usage of Prompto

class TestPrompto < Prompto
  text 'Test string', text: 'Hello, world!'

  file 'Example File',
       path: 'test/example.txt',
       description: 'This is an example file.'

  script 'Example Script',
         path: 'scripts/example.sh',
         arguments: %w[arg1 arg2],
         description: 'This is an example script.'

  ruby 'Example Ruby Block', description: 'This is an example Ruby block.' do
    (1..5).map { |i| "Number #{i}" }.join(', ')
  end

  shell 'Embedded Shell Script', description: 'This is an embedded shell script.' do
    <<~BASH
      #!/bin/bash
      echo "Current directory:"
      pwd
      echo "Files in the current directory:"
      ls -la
    BASH
  end

  define_thor_command :test
end