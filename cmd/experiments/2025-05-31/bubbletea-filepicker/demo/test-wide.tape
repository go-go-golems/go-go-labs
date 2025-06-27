# VHS Script: Test Very Wide Layout
Set FontSize 12
Set Width 1600
Set Height 600
Set Theme "Dracula"
Output demo/test-wide.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate to files
Down Down Down Down Down
Sleep 500ms

Screenshot demo/layout-very-wide.txt
Type "q"
