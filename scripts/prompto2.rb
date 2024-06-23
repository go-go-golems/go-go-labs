# frozen_string_literal: true

require 'shellwords'

# Represents a simple text fragment in the prompt
class TextFragment
  def initialize(text)
    @text = text
  end

  # Renders the text fragment
  # @return [String] The text content
  def render
    @text
  end
end

# Represents a file to be included in the prompt
class FileFragment
  def initialize(title:, path:, description: nil)
    @title = title
    @path = path
    @description = description
  end

  # Renders the file content with title and description
  # @return [String] Formatted file content in Markdown
  def render
    <<~MARKDOWN
      # #{@title}
      #{@description}
      #{File.read(@path)}
      ---
    MARKDOWN
  end
end

# Represents a script to be executed and included in the prompt
class ScriptFragment
  def initialize(title:, path:, arguments: [], description: nil)
    @title = title
    @path = path
    @arguments = arguments
    @description = description
  end

  # Executes the script and renders its output with title and description
  # @return [String] Formatted script output in Markdown
  def render
    output = `#{@path} #{@arguments.join(' ')}`
    <<~MARKDOWN
      # #{@title}
      #{@description}
      #{output}
    MARKDOWN
  end
end

# Represents a Ruby code block to be executed and included in the prompt
class RubyFragment
  def initialize(title:, description: nil, &block)
    @title = title
    @block = block
    @description = description
  end

  # Executes the Ruby block and renders its output with title and description
  # @return [String] Formatted Ruby output in Markdown
  def render
    output = @block.call
    <<~MARKDOWN
      # #{@title}
      #{@description}
      #{output}
    MARKDOWN
  end
end

# Represents an embedded shell script to be executed and included in the prompt
class EmbeddedShellFragment
  def initialize(title:, script:, description: nil)
    @title = title
    @script = script
    @description = description
  end

  # Executes the shell script and renders its content and output with title and description
  # @return [String] Formatted shell script and its output in Markdown
  def render
    output = `bash -c #{@script.shellescape}`
    <<~MARKDOWN
      # #{@title}
      #{@description}
      ```bash
      #{@script}
      ```
      Output:
      #{output}
    MARKDOWN
  end
end

# Main class for building prompts with various types of content
class Prompto
  class << self
    # @return [Array] List of fragments in the prompt
    def fragments
      @fragments ||= []
    end

    # Adds a text fragment to the prompt
    # @param content [String] The text content
    def text(content)
      fragments << TextFragment.new(content)
    end

    # Adds a file fragment to the prompt
    # @param title [String] The title of the file section
    # @param path [String] The path to the file
    # @param description [String, nil] Optional description of the file
    def file(title, path:, description: nil)
      fragments << FileFragment.new(title: title, path: path, description: description)
    end

    # Adds a script fragment to the prompt
    # @param title [String] The title of the script section
    # @param path [String] The path to the script
    # @param arguments [Array<String>] Optional arguments for the script
    # @param description [String, nil] Optional description of the script
    def script(title, path:, arguments: [], description: nil)
      fragments << ScriptFragment.new(title: title, path: path, arguments: arguments, description: description)
    end

    # Adds a Ruby code fragment to the prompt
    # @param title [String] The title of the Ruby section
    # @param description [String, nil] Optional description of the Ruby code
    # @yield The Ruby code block to be executed
    def ruby(title, description: nil, &block)
      fragments << RubyFragment.new(title: title, description: description, &block)
    end

    # Adds an embedded shell script fragment to the prompt
    # @param title [String] The title of the shell script section
    # @param description [String, nil] Optional description of the shell script
    # @yield The shell script content as a string
    def shell(title, description: nil, &block)
      script = block.call.strip
      fragments << EmbeddedShellFragment.new(title: title, script: script, description: description)
    end

    # Renders all fragments in the prompt
    # @return [String] The complete prompt content
    def render
      fragments.map(&:render).join("\n")
    end

    # Resets the prompt by clearing all fragments
    def reset
      @fragments = []
    end
  end
end

# Example usage of Prompto
class TestPrompto < Prompto
  text 'Hello, world!'

  file 'Example File',
       path: 'example.txt',
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
end

puts TestPrompto.render if __FILE__ == $PROGRAM_NAME
