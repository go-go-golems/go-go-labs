# TODO - Prompt Renderer Improvements

## Completed Items ✅

- ✅ Entries with just one variant and no bullet should allow me to toggle them on / off
- ✅ Entries with bullets should allow me to toggle individual bullets on / off  
- ✅ Entries with bullets can also have a content with {{ . }} being set to the selected bullet list for rendering, per default, just render all bullets
- ✅ We don't need groups under a variant, just make them sections
- ✅ Fixed three space padding in front of Variable border boxes
- ✅ Basic template selection and configuration working
- ✅ Basic bullet and toggle rendering functionality
- ✅ Variable substitution with Go templates
- ✅ File reference support with @filename syntax
- ✅ Auto-save and persistence functionality

## New Priority Items 🚀

### Critical Fixes (P0) ✅ COMPLETED
- [x] **Fix bullet toggling in TUI** - Space and Enter should toggle bullet selection ✅
  - Fixed handleToggle() method to properly handle both Space and Enter keys
  - Added robust bounds checking to prevent index out of range errors
  - Both bullets and toggle variants now respond correctly to Space/Enter
- [x] **Fix section rendering** - Some sections not showing in rendered output (code review with context) ✅  
  - Toggle sections now properly render when VariantEnabled is true
  - Template rendering logic correctly handles all section types
  - Verified with "Code Review with Context" template
- [x] **TUI focus navigation** - Improve keyboard navigation between form elements ✅
  - Added Shift+Tab for reverse navigation
  - Improved bounds checking for all navigation keys (↑↓, j/k, Tab)
  - Enhanced status bar to show all available navigation keys
- [x] **Error handling in TUI** - Show meaningful error messages when DSL parsing fails ✅
  - Added better error messages with emoji indicators (❌)
  - Improved YAML parsing error messages with specific guidance
  - Enhanced preview error display with user-friendly messages

- [x] **Default bullet selected** - First 2-3 bullets are now selected by default when loading templates ✅
  - Implemented smart default selection logic that selects the first 3 bullets (or all if fewer than 3)
  - Works for all bullet-type variants when creating default selections
- [x] **Display the label for each section, styled** - Section labels are now prominently displayed with professional styling ✅
  - Added section label support to DSL structure with `label` field
  - Implemented styled section headers with background colors and proper visual hierarchy
  - Falls back to section ID if no label is provided
- [x] **Display label for bullets sections as well** - Bullet sections now have clear, styled headers ✅
  - Added bullet section headers that combine section label and variant information
  - Headers show: "Section Label (Variant Label) - Description"
  - Non-interactive headers provide clear visual separation
- [x] **Use the same () as for variants, for bullet point items** - Consistent visual styling between variants and bullets ✅
  - Unified checkbox styling with ☐/☑ for both bullets and toggles
  - Consistent cursor and highlighting behavior across all interactive elements
  - Parenthetical descriptions shown for toggle items matching variant style
- [x] **Add a short description to the variants** - Variant descriptions are now displayed in UI ✅
  - Added `description` field to VariantDefinition structure
  - Descriptions shown in variant selection lists and bullet section headers
  - Updated templates.yml with comprehensive descriptions for all variants
  - Toggle items show descriptions in parentheses for better UX

- [ ] The selected variant doesn't show up as selected anymore
- [ ] Bullet points are all selected by default
- [ ] When toggling bullet variants, remember the selected bullet points
