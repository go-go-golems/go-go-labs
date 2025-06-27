# VHS Script: Test Very Narrow Width
Set FontSize 12
Set Width 600
Set Height 600
Set Theme "Dracula"
Output demo/test-responsive-very-narrow.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Take screenshot showing very narrow responsive layout
Screenshot demo/responsive-very-narrow-layout.txt

Type "q"
