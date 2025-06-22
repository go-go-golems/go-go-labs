# VHS Script to verify UI improvements in prompt-renderer
# This script demonstrates the new features:
# 1. Default bullet selections
# 2. Styled section labels 
# 3. Bullet section labels
# 4. Consistent visual styling
# 5. Variant descriptions

Output verify_ui_improvements.gif
Set FontSize 14
Set Width 1200  
Set Height 800
Set Theme "Dracula"

# Start the application
Type "./prompt-renderer"
Enter
Sleep 2s

# Show template list - demonstrates styled headers
Screenshot screenshot_template_list.png

# Select the first template (Code Review Assistant)
Enter
Sleep 2s

# Show the template configuration UI with improvements
Screenshot screenshot_improved_ui.png

# Navigate through sections to show improvements
# 1. Navigate to sections area  
Tab Tab Tab Tab
Sleep 1s

# Show Introduction Style section with descriptions
Screenshot screenshot_section_with_descriptions.png

# Navigate to Review Focus Areas section
Down Down Down Down Down Down
Sleep 1s

# Show bullet section headers and default selections
Screenshot screenshot_bullet_section_header.png

# Show that bullets have default selections (first 2-3 selected)
Sleep 2s

# Navigate through individual bullets to show consistent styling
Down Down Down
Sleep 1s
Screenshot screenshot_bullet_styling.png

# Toggle a bullet to show interaction
Space
Sleep 1s

# Navigate to toggle section to show consistent styling
Up Up Up Up Up Up Up Up Up Up
Down Down
Sleep 1s
Screenshot screenshot_toggle_styling.png

# Show the preview with default selections
Ctrl+r
Sleep 2s
Screenshot screenshot_preview_with_defaults.png

# Exit gracefully  
Escape
Sleep 1s
Type "q"
Sleep 1s
