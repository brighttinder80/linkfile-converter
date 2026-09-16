# dotconv

Most dotfiles repos end up describing their symlinks in one of two
ways: a structured list of `target -> source` mappings, or a shell
script full of hand-written `ln -s` commands that grew line by line
over the years. `dotconv` converts between the two, so you can move a
repo from one style to the other without retyping every mapping.

## Formats

**linkfile** - one mapping per line, target on the left:

```
# ~ means $HOME; the right side is relative to the dotfiles repo
~/.zshrc = zsh/zshrc
~/.gitconfig = git/gitconfig
~/.config/nvim/init.vim = nvim/init.vim
```

**script** - a shell script of `ln -s` calls, the kind of thing people
write by hand in `install.sh`:

```sh
ln -s "$REPO/zsh/zshrc" "$HOME/.zshrc"
ln -sf "$REPO/git/gitconfig" "$HOME/.gitconfig"
```

## Usage

Turn a linkfile into a ready-to-run install script:

```
$ dotconv to-script dotfiles.linkfile -o install.sh
$ chmod +x install.sh
```

Pull a linkfile out of an existing install script, e.g. when adopting
a repo that only has one:

```
$ dotconv to-linkfile legacy-install.sh -o dotfiles.linkfile
```

Check a linkfile without generating anything, e.g. from a pre-commit
hook or CI:

```
$ dotconv validate dotfiles.linkfile
ok: 14 entries
```

## Strict by default

By default `dotconv` refuses to guess. A linkfile target must be
absolute or start with `~/`; a source may not be absolute or escape
the repo with `..`. A script line that starts with `ln` must be a
plain `ln -s`/`-sf`/`-fs` call with exactly two path arguments and no
shell variables - anything else stops the conversion with an error
naming the offending line, because a dotfiles migration that silently
drops a symlink is worse than one that fails loudly.

Pass `--lenient` to relax this: unparseable or unsafe lines are
skipped with a warning on stderr instead of aborting. This is meant
for one-off migrations of messy scripts, not for everyday use.

```
$ dotconv to-linkfile --lenient legacy-install.sh -o dotfiles.linkfile
skipping line 12: variable expansion is not supported
skipping line 20: expected exactly 2 paths after ln, got 3
```

## Limitations

The script parser understands one `ln` call per line with literal or
quoted arguments; it does not evaluate `$HOME`, globs, backticks, or
line continuations. `--lenient` skips lines it can't handle rather
than trying to guess at them.

## Building

Standard library only, no external dependencies:

```
$ go build -o dotconv .
```

## License

MIT, see [LICENSE](LICENSE).
