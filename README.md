# ANSI SGR Process

**_Extract ANSI SGR Sequences from `string`_**

> [!NOTE]
> Tab`\t` will be replaced by 4 spaces` `

## Usage

- `Extract(string) -> (*ANSITableList, string)`  
  split `string with ansi` into `ansi sequences` and `raw string`
- `Search(bs []BoundStruct, position int) -> (res []int)`  
  binary search index given `Bound [2]int`, the result is index.  
  `len(res) <= 2` and `-1` means reach the start/end of the slice  
  `len(res) == 1` means the position is within `bs[res[0]].Bounds`  
  `len(res) == 2` means the position is between `bs[res[0]].Bounds[1]` and `bs[res[1]].Bounds[0]`
- `*ANSITableList SetStyle(style []byte, startIndex, endIndex int)`  
  set style for `string[startIndex:endIndex]`
