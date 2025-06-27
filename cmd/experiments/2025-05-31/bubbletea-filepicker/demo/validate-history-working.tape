# VHS Script: Validate History Navigation is Working
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"
Output demo/validate-history-working.txt

Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate to root - take screenshot showing path
Screenshot demo/step1-root.txt

# Navigate down to documents (2nd item)
Down Down
Enter
Sleep 1s

# Take screenshot in documents directory
Screenshot demo/step2-documents.txt

# Navigate into personal (2nd subdirectory)
Down Down
Enter
Sleep 1s  

# Take screenshot in personal directory
Screenshot demo/step3-personal.txt

# Use 'h' to go back to documents
Type "h"
Sleep 1s

# Take screenshot - should be back in documents
Screenshot demo/step4-back-to-documents.txt

# Use 'h' again to go back to root
Type "h"
Sleep 1s

# Take screenshot - should be back at root
Screenshot demo/step5-back-to-root.txt

# Use 'l' to go forward to documents
Type "l"
Sleep 1s

# Take screenshot - should be forward to documents
Screenshot demo/step6-forward-to-documents.txt

# Use 'l' to go forward to personal
Type "l"
Sleep 1s

# Take screenshot - should be forward to personal
Screenshot demo/step7-forward-to-personal.txt

Type "q"
