# VHS Script: Test History Navigation (Back/Forward)
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"
Output demo/test-history-navigation.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Take initial screenshot
Screenshot demo/history-initial.txt

# Navigate into documents directory
Down Down Down
Enter
Sleep 1s

# Take screenshot in documents
Screenshot demo/history-documents.txt

# Navigate into work subdirectory
Down Down
Enter
Sleep 1s

# Take screenshot in work
Screenshot demo/history-work.txt

# Navigate back using 'h' key
Type "h"
Sleep 1s

# Take screenshot after going back to documents
Screenshot demo/history-back-to-documents.txt

# Go back again using Alt+Left (should be at root)
Type "h"
Sleep 1s

# Take screenshot after going back to root
Screenshot demo/history-back-to-root.txt

# Go forward using 'l' key (should go to documents)
Type "l"
Sleep 1s

# Take screenshot after going forward
Screenshot demo/history-forward-to-documents.txt

# Go forward again using 'l' key (should go to work)
Type "l"
Sleep 1s

# Take screenshot after going forward again
Screenshot demo/history-forward-to-work.txt

Type "q"
