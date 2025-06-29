# Demo: Content Sources
# Shows different ways to provide content: command line, piped input, clipboard simulation

Output demos/demo-content.gif

Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Dracula"

# Setup
Type "cd cmd/apps/add-research"
Enter
Sleep 500ms

Type "echo '# Demo: Content Sources' && echo"
Enter
Sleep 1s

# Demonstrate command line message
Type "echo '📝 Method 1: Command line message...'"
Enter
Sleep 1s

Type "./add-research --title 'Microservices Architecture' --no-links --message 'Research findings on microservices: benefits include independent deployment and scaling, challenges include distributed complexity, key patterns are API gateway and event-driven communication.'"
Enter
Sleep 2s

Type "echo '✅ Note created from command line!'"
Enter
Sleep 1s

# Demonstrate piped input
Type "echo '' && echo '📥 Method 2: Piped input...'"
Enter
Sleep 1s

# Create content file and pipe it
Type "echo 'Performance Benchmarks Report: Database SELECT operations average 150ms, INSERT operations average 45ms, API endpoints average 25-80ms response time. Recommendations include query caching and index optimization.' > /tmp/performance_data"
Enter

Type "cat /tmp/performance_data | ./add-research --title 'Performance Analysis Q1 2024' --no-links"
Enter
Sleep 2s

Type "echo '✅ Note created from piped input!'"
Enter
Sleep 1s

# Demonstrate clipboard simulation (using a temp file)
Type "echo '' && echo '📋 Method 3: Clipboard content...'"
Enter
Sleep 1s

Type "echo 'Creating clipboard content...'"
Enter

Type "echo 'Container Orchestration Study: Kubernetes key concepts include Pods, Services, Deployments. Best practices are resource limits, health checks, autoscaling. Security considerations include RBAC and network policies.' > /tmp/clipboard_content"
Enter

Type "echo '📄 Clipboard content ready:'"
Enter
Type "cat /tmp/clipboard_content"
Enter
Sleep 2s

# Create note from "clipboard" (temp file)  
Type "cat /tmp/clipboard_content | ./add-research --title 'Kubernetes Deep Dive' --no-links"
Enter
Sleep 2s

Type "echo '✅ Note created from clipboard content!'"
Enter
Sleep 1s

# Show all three methods' results
Type "echo '' && echo '📊 Summary of content sources:'"
Enter
Sleep 1s

Type "echo '1. Command line message:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Microservices-Architecture*' -exec cat {} \\;"
Enter
Sleep 2s

Type "echo '' && echo '2. Piped input:'"
Enter  
Type "find ~/code/wesen/obsidian-vault/research -name '*Performance-Analysis*' -exec cat {} \\;"
Enter
Sleep 2s

Type "echo '' && echo '3. Clipboard content:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Kubernetes-Deep-Dive*' -exec cat {} \\;"
Enter
Sleep 3s

# Show flexibility with mixed content
Type "echo '' && echo '🔄 Bonus: Combining message + file content...'"
Enter
Sleep 1s

Type "echo 'Combined content example' > summary.txt"
Enter

Type "./add-research --title 'Multi-Source Research' --file summary.txt --no-links --message 'This note demonstrates combining multiple content sources in one organized note.'"
Enter
Sleep 2s

Type "echo '📄 Multi-source note preview:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Multi-Source-Research*' -exec cat {} \\;"
Enter
Sleep 3s

# Show the different input methods
Type "echo '' && echo '⚡ Content input methods available:'"
Enter
Type "echo '• --message flag: Direct command line input'"
Enter
Type "echo '• Piped input: echo content | add-research'"
Enter
Type "echo '• --clip flag: Use clipboard content'"
Enter
Type "echo '• --file flag: Attach and embed files'"
Enter
Type "echo '• Interactive mode: Type content directly'"
Enter
Sleep 3s

# Clean up
Type "echo '' && echo '🧹 Cleaning up...'"
Enter
Type "rm -f /tmp/clipboard_content /tmp/performance_data summary.txt"
Enter
Sleep 1s

Type "echo '✅ Content sources demo completed!'"
Enter
Sleep 2s

Screenshot demos/demo-content.txt
