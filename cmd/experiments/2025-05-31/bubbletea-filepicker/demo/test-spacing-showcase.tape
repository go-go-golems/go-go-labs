# VHS Script: Showcase Size-Date Spacing Improvements
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"
Output demo/spacing-showcase.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Take screenshot showing improved spacing
Screenshot demo/size-date-spacing.txt

# Navigate to show different file sizes
Down Down Down Down Down Down
Sleep 1s

# Take another screenshot
Screenshot demo/various-file-sizes.txt

Type "q"
