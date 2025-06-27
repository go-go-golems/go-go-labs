# VHS Script: Test Narrow Width Layout
Set FontSize 12
Set Width 700
Set Height 600
Set Theme "Dracula"
Output demo/test-narrow.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate to files
Down Down Down Down Down
Sleep 500ms

Screenshot demo/layout-narrow.txt
Type "q"
