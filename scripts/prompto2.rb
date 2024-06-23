# frozen_string_literal: true

require 'shellwords'

class TextFragment
  def initialize(text)
    @text = text
  end

  def render
    @text
  end
end

class FileFragment
  def initialize(title:, path:, description: nil)
    @title = title
    @path = path
    @description = description
  end

  def render
    <<~MARKDOWN
      # #{@title}
      #{@description}
      #{File.read(@path)}
      ---
    MARKDOWN
  end
end

class ScriptFragment
  def initialize(title:, path:, arguments: [], description: nil)
    @title = title
    @path = path
    @arguments = arguments
    @description = description
  end

  def render
    output = `#{@path} #{@arguments.join(' ')}`
    <<~MARKDOWN
      # #{@title}
      #{@description}
      #{output}
    MARKDOWN
  end
end

class RubyFragment
  def initialize(title:, description: nil, &block)
    @title = title
    @block = block
    @description = description
  end

  def render
    output = @block.call
    <<~MARKDOWN
      # #{@title}
      #{@description}
      #{output}
    MARKDOWN
  end
end

class EmbeddedShellFragment
  def initialize(title:, script:, description: nil)
    @title = title
    @script = script
    @description = description
  end

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

class Prompto
  class << self
    def fragments
      @fragments ||= []
    end

    def text(content)
      fragments << TextFragment.new(content)
    end

    def file(title, path:, description: nil)
      fragments << FileFragment.new(title: title, path: path, description: description)
    end

    def script(title, path:, arguments: [], description: nil)
      fragments << ScriptFragment.new(title: title, path: path, arguments: arguments, description: description)
    end

    def ruby(title, description: nil, &block)
      fragments << RubyFragment.new(title: title, description: description, &block)
    end

    def shell(title, description: nil, &block)
      script = block.call.strip
      fragments << EmbeddedShellFragment.new(title: title, script: script, description: description)
    end

    def render
      fragments.map(&:render).join("\n")
    end

    def reset
      @fragments = []
    end
  end
end

class TestPrompto < Prompto
  text 'Hello, world!'

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
end

puts TestPrompto.render if __FILE__ == $PROGRAM_NAME
