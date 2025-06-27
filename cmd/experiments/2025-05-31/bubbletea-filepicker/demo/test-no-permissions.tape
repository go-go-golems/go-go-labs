# VHS Script: Test Layout Without Permissions Column
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"
Output demo/test-no-permissions.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Take screenshot showing layout without permissions
Screenshot demo/no-permissions-layout.txt

# Navigate to show various files
Down Down Down Down Down Down Down
Sleep 1s

# Take another screenshot with selection
Screenshot demo/no-permissions-with-selection.txt

# Test multi-selection
Space
Sleep 300ms
Up
Space
Sleep 300ms

# Take screenshot showing multi-selection without permissions
Screenshot demo/no-permissions-multi-select.txt

Type "q"
