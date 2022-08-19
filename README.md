# ansigo

- [ansigo](#ansigo)
  - [Overview](#overview)
  - [Quick Start](#quick-start)
  - [Future plans (Short term)](#future-plans-short-term)
  - [Future plans (Long term)](#future-plans-long-term)

## Overview

_ansigo_ is a helper library that allows to you use common tags inside of text that result in [ANSI escape code](https://en.wikipedia.org/wiki/ANSI_escape_code#Colors) (color). Currently only one directional parsing is possible ( *tagged strings* ⮕ *color escaped strings* )

- [ansigo.go](ansigo.go) Contains the code and structs for the entire module, noteably [ansigo.Parse()](https://github.com/Volte6/ansigo/blob/master/ansigo.go#L53).
- [ansigo_test.go](ansigo_test.go) Contains unit tests, benchmarks, etc
- [testdata/ansigo_test.yaml](testdata/ansigo_test.yaml) Contains unit test data with input & expect output. The ANSI _Control Sequence Introducer_ should be represented by a unicode escaped value - `\u001b` (Octal `33`, Hexadecimal `1b`, Decimal `27`)

## Quick Start

Import the module:

    import "github.com/Volte6/ansigo"

Parse and print a string:

    fmt.Println( ansigo.Parse("This is a <ansi fg='red' bg='blue'>white text on a blue background</ansi>") )

Result:

![alt text](https://user-images.githubusercontent.com/143822/185706504-99d32ed5-37cc-4266-b682-c74b719e4790.png)

## Future plans (Short term)

- Expand to [8 bit color](https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit) support
- CSI sequence support such as cursor position

## Future plans (Long term)

- Stripping out color codes / tags from strings
- Generating color-styled HTML from color codes / tags
- Load color string-to-code map from flat file to allow extending easily. Possibly include other alias functionality.
