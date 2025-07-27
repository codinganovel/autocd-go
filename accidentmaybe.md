# The Accidental Directory History Feature

## The Context

I was concerned that some users might develop workflows where they heavily use autocd across multiple applications, causing shells to build up on top of each other and potentially hinder performance. To address this, I wanted to implement shell depth warnings that would alert users when they had accumulated many nested shells.

While implementing these warnings for my autocd-go library, I needed to detect when users had reached problematic shell depths. This led me to investigate the `SHLVL` environment variable and understand exactly how autocd creates shell nesting.

## What is SHLVL?

`SHLVL` (Shell Level) is an environment variable that tracks how many nested shells you're currently in:
- `SHLVL=1`: Your original shell when you opened the terminal
- `SHLVL=2`: You started a subshell (like running `bash` inside bash)
- `SHLVL=3`: You started another subshell inside that one
- And so on...

Each time you start a new shell, it increments SHLVL by 1. When you `exit` a shell, you return to the previous shell level.

## How AutoCD Creates Shell Nesting

Every time you use autocd to exit a program with directory inheritance, it:
1. Creates a temporary script that navigates to the target directory
2. Replaces the current process with that script using `syscall.Exec`
3. The script then starts a new shell in the target directory
4. This new shell has `SHLVL` incremented by 1
5. The previous shell remains "underneath" in the shell stack

This is the simplest way to achieve directory inheritance without requiring shell wrappers or complex integration.

## The Discovery

While tracking SHLVL for the warnings, I realized something fascinating: **each shell level represents a previous location in your navigation history**.

## A Step-by-Step Example

Let's trace what happens when you use autocd multiple times:

```bash
# Start in your home directory
SHLVL=1  pwd: /home/user

# Use vim with autocd, exit to /projects/app1
vim /projects/app1/main.go
# (autocd creates new shell in /projects/app1)
SHLVL=2  pwd: /projects/app1

# Use ranger with autocd, exit to /projects/app2  
ranger
# (autocd creates new shell in /projects/app2)
SHLVL=3  pwd: /projects/app2

# Use another app with autocd, exit to /etc
some-app
# (autocd creates new shell in /etc)
SHLVL=4  pwd: /etc
```

Now you have a shell stack where:
- `SHLVL=4`: `/etc` (current location)
- `SHLVL=3`: `/projects/app2` (where ranger was)
- `SHLVL=2`: `/projects/app1` (where vim was)  
- `SHLVL=1`: `/home/user` (original terminal)

You can navigate backwards through this history by typing `exit`.

## The Accidental Feature

If you've used autocd across multiple applications and now have SHLVL=15, you effectively have a **15-level directory history** stored in your shell stack:

```
SHLVL=15: /current/working/directory     ← You are here
SHLVL=14: /previous/location            ← exit once to go back
SHLVL=13: /location/before/that         ← exit twice to go back
SHLVL=12: /even/earlier/location        ← exit three times
...
SHLVL=1:  /original/starting/directory  ← exit 14 times to reach start
```

## Navigation Through History

Users can navigate backwards through their directory history by simply typing `exit`:

```bash
pwd                    # /projects/current-app
exit                   # Back to previous directory
pwd                    # /home/user/documents  
exit                   # Back further
pwd                    # /usr/local/bin
exit                   # Back to where you started before using autocd apps
```

Each `exit` command pops you back one level in your navigation history.

## The Reframing

This discovery completely reframes the "shell nesting problem":

### Traditional View (Performance Problem)
- **Problem**: Deep shell nesting hurts performance
- **Solution**: Start fresh terminal to "clean up"
- **User Action**: Lose all navigation context

### New View (Accidental Feature)
- **Feature**: Implicit directory history through shell stack
- **Benefit**: Navigate backwards through your path without tools
- **Trade-off**: Performance vs. navigation utility

## Implications for Warnings

The current shell depth warnings focus purely on performance:

```
💡 Tip: You have 18 nested shells from navigation.
For better performance, consider opening a fresh terminal.
```

But they could acknowledge the navigation benefit:

```
💡 Tip: You have 18 nested shells from navigation.
You can use 'exit' to navigate back through your directory history.
For better performance, consider opening a fresh terminal when done exploring.
```

## The Design Philosophy Question

This raises an interesting question about autocd's design philosophy:

1. **Is this a bug or a feature?**
   - Technically it's an implementation side-effect
   - But it provides genuine utility
   - Many users might prefer it to external history tools

2. **Could I optimize it away?**
   - Not really - avoiding shell nesting would require shell wrappers
   - That's exactly what autocd was designed to avoid
   - The current approach is the simplest solution that works

3. **Should I promote this as a feature?**
   - It only works if all your apps use autocd
   - Most users would have a mix of autocd and non-autocd applications
   - It's more of a cool unintended side-effect than a reliable feature

## Real-World Usage Patterns

This suggests users might naturally develop workflows like:

```bash
# Work in project A
vim /projects/project-a/src/main.go
# (autocd exits vim to /projects/project-a/src/)

# Navigate to project B  
cd /projects/project-b/docs
ranger
# (autocd exits ranger to /projects/project-b/docs/)

# Need to go back to project A quickly?
exit  # Back to /projects/project-b (where ranger was launched)
exit  # Back to /projects/project-a/src/ (where vim was launched)
```

## The Unintended Consequence

Sometimes simple solutions have unexpected side effects. The autocd library's shell nesting approach:

- **Solves the primary problem**: Directory inheritance after program exit
- **Uses simple, portable mechanics**: Standard shell behavior
- **Creates unintended behavior**: Implicit directory history (but only for autocd apps)
- **Requires no additional complexity**: Works with existing shell features

This is an example of how focusing on **simple, correct implementation** can lead to **interesting unintended consequences**.

## Conclusion

What started as a "performance problem" to warn users about might actually be **cool unintended behavior** that some users find more valuable than the performance cost.

The shell nesting in autocd isn't just a side-effect—it's an **unintentional directory history system**, though it only works reliably if all your applications use autocd.

Maybe the real question isn't "how do I warn users about shell nesting?" but rather "how do I help users understand this interesting behavior so they can decide if it's useful for their workflow?"