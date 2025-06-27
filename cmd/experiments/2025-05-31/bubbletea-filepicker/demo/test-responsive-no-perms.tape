# VHS Script: Test Responsive Layout Without Permissions
Set FontSize 12
Set Width 900
Set Height 600
Set Theme "Dracula"
Output demo/test-responsive-no-perms.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Take screenshot showing responsive layout without permissions
Screenshot demo/responsive-no-perms.txt

Type "q"
