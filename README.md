# BranchKit Windows

Window snapping, space switching, and moving windows between spaces for
[BranchKit](https://branchkit.dev). MIT licensed.

Uses built-in OS primitives — no tiling engine, no window server hacks.

## What it provides

**Actions** (`windows.*`): `snap` (left, right, maximize, center, next, prev),
`move_to_space`, `desk_switch`.

**Collections**: `plugin.windows.snap_mode` and `plugin.windows.desk_mode` —
both **exclusive tag gates**.

Requires `windows`, `input`, `shell`, and `display`. macOS.

## The interesting part: exclusive gates

`snap_mode` is the smallest complete example of the subtlest thing in the
platform model, which is why it is worth reading even if you never want a
window manager.

An exclusive gate does two things at once. At the matcher, while the tag is
active, every command that does not require an active exclusive tag is
suppressed — so in snap mode only the directional commands are eligible.
At the recognition engine, the same derivation narrows the grammar to that
command set. One declaration, two effects, and they cannot drift apart because
they are projections of the same derivation.

The converse matters too: because `requires_tags` is a conjunction, commands
behind an exclusive gate contribute no structure to the free-context grammar.
This plugin's bare `<number>` in desk mode is why "five" alone does not become
decodable outside desk mode.

Contrast with a non-exclusive gate, which only augments — everything else stays
live. High-churn contexts must be non-exclusive.

## Reading this as an example

This is a **reference implementation, not a tutorial.** It is a real shipped
plugin, carrying real scar tissue: one comment records a panic mid-drag that
left the left mouse button physically held system-wide. Read it to see how the
platform is actually used, not to learn house style.

For idiom, read
[branchkit-plugin-helloworld-go](https://github.com/branchkit/branchkit-plugin-helloworld-go)
or scaffold with `branchkit-cli dev init`.

## Build

```bash
cd src && go build -o ../windows-plugin .
```

Install into a running BranchKit:

```bash
branchkit-cli plugin install . --build
```

## Platform documentation

```bash
branchkit-cli docs sync
grep -rl "exclusive" "$(branchkit-cli docs path)"
```
