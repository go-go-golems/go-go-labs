# frozen_string_literal: true

require 'shellwords'
require 'thor'

# frozen_string_literal: true

# Represents a simple text fragment in the prompt
class TextFragment
  def initialize(title:, text:, description: nil)
    @title = title
    @text = text
    @description = description
  end

  # Renders the text fragment with title and description
  # @return [String] Formatted text content in Markdown
  def render
    <<~MARKDOWN
      # #{@title}

      #{@description}

      #{@text}

      ---

    MARKDOWN
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

      ```
      #{output}
      ```

      ---

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

      ```ruby
      #{output}
      ```

      ---

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
      ```
      #{output}
      ```

      ---

    MARKDOWN
  end
end

# Main class for building prompts with various types of content
class Prompto < Thor
  class << self
    # @return [Array] List of fragments in the prompt
    def fragments
      @fragments ||= []
    end

    # Initializes a new @fragments array for each Prompto subclass.
    #
    # This method is automatically called by Ruby when Prompto is subclassed.
    # It ensures that each subclass has its own independent collection of fragments,
    # preventing unintended sharing of fragments between different Prompto subclasses.
    #
    # @param subclass [Class] The newly created subclass of Prompto
    #
    # @example
    #   class MyPrompto < Prompto
    #     # This subclass will automatically get its own @fragments array
    #   end
    def inherited(subclass)
      super
      subclass.instance_variable_set(:@fragments, [])
    end

    # Adds a text fragment to the prompt
    # @param title [String] The title of the text section
    # @param text [String] The text content
    # @param description [String, nil] Optional description of the text
    def text(title, text:, description: nil)
      fragments << TextFragment.new(title: title, text: text, description: description)
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

    # Defines a Thor command for rendering the Prompto
    # @param command_name [Symbol] The name of the Thor command
    def define_thor_command(command_name, default: true)
      # capture fragments for the thor method definition
      fragments_ = @fragments

      desc command_name.to_s, "Render the #{name} Prompto"
      method_option :about, type: :boolean, desc: 'Display metadata about the Prompto'
      fragments_.each do |fragment|
        method_option fragment.class.name.downcase.to_sym, type: :boolean, default: true,
                                                           desc: "Include #{fragment.class.name} fragments"
      end

      define_method(command_name) do
        prompto_class = Object.const_get(self.class.name.split('::').first)
        if options[:about]
          puts "Prompto: #{prompto_class.name}"
          puts 'Available fragments:'
          prompto_class.fragments.each do |fragment|
            puts "  - #{fragment.class.name}: #{fragment.instance_variable_get(:@title)}"
          end
        else
          filtered_fragments = prompto_class.fragments.select do |fragment|
            options[fragment.class.name.downcase.to_sym]
          end
          puts filtered_fragments.map(&:render).join("\n")
        end
      end

      return unless default

      default_command command_name
    end
  end
end

# Dynamically load Prompto classes from ~/.prompts directory
module PromptoDynamicLoader
  def self.load_prompto_classes
    prompts_dir = File.expand_path('scripts/prompts')
    Dir.glob(File.join(prompts_dir, '*.rb')).sort.each do |file|
      require file
    end

    # Find all subclasses of Prompto
    prompto_classes = ObjectSpace.each_object(Class).select { |klass| klass < Prompto }

    # Register each Prompto subclass as a subcommand of the main CLI class
    prompto_classes.each do |klass|
      command_name = klass.name.gsub(/Prompto$/, '').downcase
      CLI.desc command_name, "Run #{klass.name}"
      CLI.subcommand command_name, klass
    end
  end
end

class CLI < Thor
  # The existing subcommands will be dynamically added here
end

# Load and register Prompto classes
PromptoDynamicLoader.load_prompto_classes

CLI.start(ARGV) if __FILE__ == $PROGRAM_NAME
