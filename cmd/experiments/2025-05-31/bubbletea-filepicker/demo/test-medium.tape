# VHS Script: Test Medium Width Layout
Set FontSize 12
Set Width 1000
Set Height 600
Set Theme "Dracula"
Output demo/test-medium.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate to files
Down Down Down Down Down
Sleep 500ms

Screenshot demo/layout-medium.txt
Type "q"
