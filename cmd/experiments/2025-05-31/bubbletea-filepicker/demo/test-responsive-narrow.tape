# VHS Script: Test Responsive Layout with Narrow Width
Set FontSize 12
Set Width 800
Set Height 600
Set Theme "Dracula"
Output demo/test-responsive-narrow.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Take screenshot showing responsive layout
Screenshot demo/responsive-narrow-layout.txt

Type "q"
