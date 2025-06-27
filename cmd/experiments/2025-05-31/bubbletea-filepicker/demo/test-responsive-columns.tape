# VHS Script: Test Responsive Column Layout
# Tests column hiding at different window widths

# Test 1: Very wide window - should show all columns
Output demo/test-responsive-1-wide.txt
Set FontSize 12
Set Width 1600
Set Height 600
Set Theme "Dracula"

Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate to show files with larger sizes
Down Down Down Down Down
Sleep 500ms

Screenshot demo/responsive-very-wide.txt
Type "q"

# Test 2: Medium width - should hide permissions
Output demo/test-responsive-2-medium.txt  
Set Width 1000
Set Height 600

Type "./filepicker test-files"
Enter
Sleep 2s

Down Down Down Down Down
Sleep 500ms

Screenshot demo/responsive-medium.txt
Type "q"

# Test 3: Narrow width - should hide permissions and date
Output demo/test-responsive-3-narrow.txt
Set Width 700  
Set Height 600

Type "./filepicker test-files"
Enter
Sleep 2s

Down Down Down Down Down
Sleep 500ms

Screenshot demo/responsive-narrow.txt
Type "q"

# Test 4: Very narrow - should hide all detail columns
Output demo/test-responsive-4-very-narrow.txt
Set Width 500
Set Height 600

Type "./filepicker test-files"
Enter
Sleep 2s

Down Down Down Down Down
Sleep 500ms

Screenshot demo/responsive-very-narrow.txt
Type "q"
