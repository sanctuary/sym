# sym

Parse Playstation 1 symbol files (`*.SYM`).

## Installation

```bash
go get -u github.com/sanctuary/sym/cmd/sym_dump
```

## Usage

```bash
sym_dump DIABPSX.SYM
# Output:
#
# Header : MND version 1
# Target unit 0
# 000008: $800b031c overlay length $000009e4 id $4
# 000015: $800b031c overlay length $00000004 id $5
# 000022: $80139bf8 overlay length $00023234 id $b
# 00002f: $80139bf8 overlay length $00029dcc id $c
# 00003c: $80139bf8 overlay length $0002a228 id $d
# 000049: $80139bf8 overlay length $0001ec70 id $e
# 000056: $00000000 94 Def class TPDEF type UCHAR size 0 name u_char
```

The output of `sym_dump` is identical to the `DUMPSYM.EXE` tool of the [Psy-Q SDK](http://www.psxdev.net/help/psyq_install.html).
