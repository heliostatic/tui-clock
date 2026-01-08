# VHS tape for generating README screenshots
# Run with: vhs demo/demo.tape

# Terminal settings
Set Shell "bash"
Set FontSize 22
Set Width 1024
Set Height 500
Set Theme "Catppuccin Mocha"
Set Padding 20

# Build and run with demo config (hidden from output)
Hide
Type "cd /Users/bc/Code/tui-clock && go build -o tui-clock && ./tui-clock -config demo/config.yaml"
Enter
Sleep 2s
Show

# Screenshot 1: Normal mode
Sleep 500ms
Screenshot assets/normal-mode.png

# Screenshot 2: Timeline individual mode
Sleep 500ms
Type "t"
Sleep 1s
Screenshot assets/timeline-individual.png

# Screenshot 3: Timeline shared mode
Sleep 500ms
Type "m"
Sleep 1s
Screenshot assets/timeline-shared.png

# Exit
Type "q"
