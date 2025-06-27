# VHS Script: Test History Navigation Help Display
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"
Output demo/test-history-help.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Show help to see if back/forward keys are displayed
Type "?"
Sleep 2s

# Take screenshot of help
Screenshot demo/help-with-history.txt

# Close help
Type "?"
Sleep 1s

Type "q"
