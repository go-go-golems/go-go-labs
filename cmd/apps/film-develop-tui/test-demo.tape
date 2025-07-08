Output test-demo.txt
Set Width 120
Set Height 120
Set Shell "bash"

Type "./film-develop-tui --log-level=warn"
Enter
Sleep 2s
Screenshot test-demo.txt
Sleep 1s
Type "f"
Sleep 1s
Screenshot film-selection.txt
Sleep 1s
Type "1"
Sleep 1s
Screenshot ei-selection.txt
Sleep 1s
Type "2"
Sleep 1s
Screenshot roll-selection.txt
Sleep 1s
Type "a"
Sleep 1s
Screenshot calculated-results.txt
Sleep 1s
Type "q"
Sleep 1s
