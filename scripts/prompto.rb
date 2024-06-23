# frozen_string_literal: true

class TextFragment
  def initialize(text)
    @text = text
  end

  def render
    @text
  end

end

class FileFragment
  def initialize(title, path, description = nil)
    @title = title
    @path = path
    @description = description
  end

  def render
    <<~MARKDOWN
      # #{@title}

      #{@description unless @description}

      #{File.read(@path)}

      ---
    MARKDOWN
  end
end

class ScriptFragment
  def initialize(title, path, arguments = [], description = nil)
    @title = title
    @path = path
    @arguments = arguments
    @description = description
  end

  def render
    # run script with given arguments
    output = `#{@path} #{@arguments}`

    <<~MARKDOWN
      # #{@title}

      #{@description unless @description}

      #{output}
    MARKDOWN
  end
end

class Prompto
  def initialize
    @fragments = []
  end

  def add_text(text)
    @fragments << TextFragment.new(text)
  end

  def add_file(title, path, description = nil)
    @fragments << FileFragment.new(title, path, description: description)
  end

  def add_script(title, path, arguments = [], description = nil)
    @fragments << ScriptFragment.new(title, path, arguments: arguments, description: description)
  end

  def render
    @fragments.map(&:render).join("\n")
  end

end


class TestPrompto < Prompto
  def initialize
    super
    add_text("Hello, world!")
    add_file("Example File", "example.txt", description: "This is an example file.")
    add_script("Example Script", "scripts/example.sh", ["arg1", "arg2"], description: "This is an example script.")
  end
end

if __FILE__ == $0
  puts TestPrompto.new.render
end