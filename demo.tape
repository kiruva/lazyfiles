# Demo recording script for lazyfiles.
# Install VHS (https://github.com/charmbracelet/vhs), then run:
#     vhs demo.tape
# It produces demo.gif, referenced by the README.

Output demo.gif

Set FontSize 16
Set Width 1200
Set Height 720
Set Padding 20

# Build and launch, hidden from the recording.
Hide
Type "go build -o /tmp/lazyfiles . && cd /tmp && /tmp/lazyfiles"
Enter
Sleep 1s
Show

# Move around the active pane.
Sleep 500ms
Down@300ms 4
Sleep 500ms

# Switch to the other pane and back.
Tab
Sleep 700ms
Down@300ms 2
Tab
Sleep 500ms

# Mark a couple of files.
Space
Sleep 300ms
Down@300ms 1
Space
Sleep 700ms

# Open the help overlay, then close it.
Type "?"
Sleep 2s
Escape
Sleep 500ms

# Quit.
Type "q"
Sleep 500ms
