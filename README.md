# ansitags

- [ansitags](#ansitags)
  - [Overview](#overview)
  - [Quick Start](#quick-start)
  - [Future plans (Short term)](#future-plans-short-term)
  - [Future plans (Long term)](#future-plans-long-term)

## Overview

_ansitags_ is a helper library that allows to you use common tags inside of text that result in [ANSI escape code](https://en.wikipedia.org/wiki/ANSI_escape_code#Colors) (color). Currently only one directional parsing is possible ( *tagged strings* ⮕ *color escaped strings* )

- [ansitags.go](ansitags.go) Contains the code and structs for the basic parsing logic and flow of data, noteably `ansitags.Parse()` and `ansitags.ParseStreaming()`.
- [ansiproperties.go](ansiproperties.go) handles basic ansi properties/tag parsing and conversion into valid escape codes.
- [tagmatcher.go](tagmatcher.go) basic helper struct to simplify finding ansi "tag" matches.
- [ansitags_test.go](ansitags_test.go) Contains unit tests, benchmarks, etc
- [testdata/ansitags_test.yaml](testdata/ansitags_test.yaml) Contains unit test data with input & expect output. The ANSI _Control Sequence Introducer_ should be represented by a unicode escaped value - `\u001b` (Octal `33`, Hexadecimal `1b`, Decimal `27`)

## Quick Start

Import the module:

    import "github.com/GoMudEngine/ansitags"

Parse and print a string:

    fmt.Println( ansitags.Parse("This is a <ansi fg='red' bg='blue'>white text on a blue background</ansi>") )

Result:

![alt text](https://user-images.githubusercontent.com/143822/185706504-99d32ed5-37cc-4266-b682-c74b719e4790.png)

Note: You can switch between 256 color mode and 8 color mode (The default is 8):

    ansitags.SetColorMode(ansitags.Color8)
    ansitags.SetColorMode(ansitags.Color256)


## Future plans (Short term)

- CSI sequence support such as cursor position

## Future plans (Long term)

- Stripping out color codes / tags from strings
- Generating color-styled HTML from color codes / tags
